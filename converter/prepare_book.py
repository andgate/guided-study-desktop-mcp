#!/usr/bin/env python3
"""Render a PDF to page images and a compact table-of-contents CSV."""

from __future__ import annotations

import argparse
import csv
import logging
import os
import re
import shutil
import subprocess
import sys
from collections.abc import Iterable
from pathlib import Path
from typing import Any

from pypdf import PdfReader


def destination_page(reader: PdfReader, item: Any) -> int | None:
    try:
        page = reader.get_destination_page_number(item)
    except Exception as exc:
        logging.warning("Skipping unreadable outline item: %s", exc)
        return None
    if page is None or page < 0 or page >= len(reader.pages):
        return None
    return page + 1


def outline_title(item: Any) -> str:
    title = getattr(item, "title", None)
    if title is None and isinstance(item, dict):
        title = item.get("/Title") or item.get("title")
    return str(title or "Untitled").strip() or "Untitled"


def flatten_outline(reader: PdfReader, items: Iterable[Any], depth: int = 0):
    last_item_had_children = False
    for item in items:
        if isinstance(item, list):
            child_depth = depth + 1 if last_item_had_children else depth
            yield from flatten_outline(reader, item, child_depth)
            last_item_had_children = False
            continue
        page = destination_page(reader, item)
        if page is not None:
            yield depth, outline_title(item), page
        last_item_had_children = True


def locate_pdftoppm(explicit: str | None) -> str:
    bundled = (
        Path.home()
        / ".cache/codex-runtimes/codex-primary-runtime/dependencies/native/poppler/Library/bin/pdftoppm.exe"
    )
    for value in (
        explicit,
        os.environ.get("PDFTOPPM_PATH"),
        shutil.which("pdftoppm"),
        str(bundled),
    ):
        if value and Path(value).is_file():
            return str(value)
    raise FileNotFoundError("pdftoppm was not found; install Poppler or set PDFTOPPM_PATH")


def prepare(pdf_path: Path, output_dir: Path, dpi: int, quality: int, pdftoppm: str | None):
    if pdf_path.suffix.lower() != ".pdf" or not pdf_path.is_file():
        raise ValueError(f"Source is not a readable PDF: {pdf_path}")
    if not output_dir.is_dir() or any(output_dir.iterdir()):
        raise ValueError("Output directory must exist and be empty")
    if not 72 <= dpi <= 600:
        raise ValueError("DPI must be between 72 and 600")
    if not 1 <= quality <= 100:
        raise ValueError("JPEG quality must be between 1 and 100")

    reader = PdfReader(str(pdf_path))
    page_count = len(reader.pages)
    if page_count < 1:
        raise ValueError("PDF has no pages")

    executable = locate_pdftoppm(pdftoppm)
    prefix = output_dir / "rendered"
    command = [
        executable,
        "-jpeg",
        "-r",
        str(dpi),
        "-jpegopt",
        f"quality={quality}",
        str(pdf_path),
        str(prefix),
    ]
    flags = subprocess.CREATE_NO_WINDOW if os.name == "nt" else 0
    subprocess.run(command, check=True, creationflags=flags)

    pattern = re.compile(r"rendered-(\d+)\.jpg$", re.IGNORECASE)
    rendered: dict[int, Path] = {}
    for path in output_dir.glob("rendered-*.jpg"):
        match = pattern.fullmatch(path.name)
        if match:
            rendered[int(match.group(1))] = path
    if set(rendered) != set(range(1, page_count + 1)):
        raise RuntimeError("pdftoppm did not render a contiguous page set")
    for page_index, path in rendered.items():
        path.rename(output_dir / f"page-{page_index:04d}.jpg")

    outline = getattr(reader, "outline", None) or getattr(reader, "outlines", None) or []
    rows = list(flatten_outline(reader, outline))
    with (output_dir / "toc.csv").open("w", encoding="utf-8", newline="") as stream:
        writer = csv.writer(stream, lineterminator="\n")
        writer.writerow(["position", "depth", "title", "page_index"])
        for position, (depth, title, page_index) in enumerate(rows, start=1):
            writer.writerow([position, depth, title, page_index])
    return page_count, len(rows)


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--input", required=True)
    parser.add_argument("--output", required=True)
    parser.add_argument("--dpi", type=int, default=200)
    parser.add_argument("--jpeg-quality", type=int, default=90)
    parser.add_argument("--pdftoppm")
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(message)s")
    try:
        pages, entries = prepare(
            Path(args.input).resolve(),
            Path(args.output).resolve(),
            args.dpi,
            args.jpeg_quality,
            args.pdftoppm,
        )
    except Exception as exc:
        logging.error("Conversion failed: %s", exc)
        return 1
    logging.info("Rendered %d pages with %d TOC entries", pages, entries)
    return 0


if __name__ == "__main__":
    sys.exit(main())
