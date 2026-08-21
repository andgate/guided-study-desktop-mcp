#!/usr/bin/env python3
"""Render a PDF to page images and a compact table-of-contents CSV."""

from __future__ import annotations

import argparse
import csv
import logging
import sys
from pathlib import Path

import pymupdf

logger = logging.getLogger(__name__)


def outline_title(title: object) -> str:
    return str(title or "Untitled").strip() or "Untitled"


def prepare(pdf_path: Path, output_dir: Path, dpi: int, quality: int):
    if pdf_path.suffix.lower() != ".pdf" or not pdf_path.is_file():
        raise ValueError(f"Source is not a readable PDF: {pdf_path}")
    if not output_dir.is_dir() or any(output_dir.iterdir()):
        raise ValueError("Output directory must exist and be empty")
    if not 72 <= dpi <= 600:
        raise ValueError("DPI must be between 72 and 600")
    if not 1 <= quality <= 100:
        raise ValueError("JPEG quality must be between 1 and 100")

    with pymupdf.open(pdf_path) as document:
        page_count = document.page_count
        if page_count < 1:
            raise ValueError("PDF has no pages")

        for page_index, page in enumerate(document, start=1):
            pixmap = page.get_pixmap(dpi=dpi, colorspace=pymupdf.csRGB, alpha=False)
            pixmap.save(
                output_dir / f"page-{page_index:04d}.jpg",
                jpg_quality=quality,
            )

        rows = [
            (level - 1, outline_title(title), page_index)
            for level, title, page_index in document.get_toc()
            if 1 <= page_index <= page_count
        ]

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
    args = parser.parse_args()
    logging.basicConfig(level=logging.INFO, format="%(message)s")
    try:
        pages, entries = prepare(
            Path(args.input).resolve(),
            Path(args.output).resolve(),
            args.dpi,
            args.jpeg_quality,
        )
    except (OSError, RuntimeError, ValueError) as exc:
        logger.error("Conversion failed: %s", exc)
        return 1
    logger.info("Rendered %d pages with %d TOC entries", pages, entries)
    return 0


if __name__ == "__main__":
    sys.exit(main())
