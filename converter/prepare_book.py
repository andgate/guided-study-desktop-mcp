#!/usr/bin/env python3
"""Render and store one PDF atomically."""

from __future__ import annotations

import json
import sqlite3
import sys
import uuid
from collections.abc import Iterator
from pathlib import Path
from typing import Any

import pymupdf

MIN_DPI = 72
MAX_DPI = 600
DEFAULT_DPI = 200

MIN_JPEG_QUALITY = 1
MAX_JPEG_QUALITY = 100
DEFAULT_JPEG_QUALITY = 90

MIME_TYPE = "image/jpeg"
BUSY_TIMEOUT_MS = 5000
OUTLINE_REQUIRED = "outline_required"
OUTLINE_UNUSABLE = "outline_unusable"


class ImportFailure(Exception):
    """Report a stable import failure."""

    def __init__(self, code: str, message: str):
        super().__init__(message)
        self.code = code
        self.message = message


def page_rows(
    document: pymupdf.Document,
    book_id: str,
    dpi: int,
    quality: int,
) -> Iterator[tuple[str, int, str, bytes]]:
    for page_index, page in enumerate(document, start=1):
        pixmap = page.get_pixmap(dpi=dpi, colorspace=pymupdf.csRGB, alpha=False)
        image = pixmap.tobytes("jpeg", jpg_quality=quality)
        yield book_id, page_index, MIME_TYPE, image


def outline_rows(
    outline: list[list[Any]],
    book_id: str,
) -> Iterator[tuple[str, int, str, int]]:
    for outline_index, entry in enumerate(outline):
        yield (
            book_id,
            outline_index,
            entry[1],
            entry[2],
        )


def check_request(
    request: dict[str, Any],
) -> tuple[Path, Path, str, int, int]:
    pdf_path = Path(request["file_reference"]).resolve()
    database_path = Path(request["database_path"]).resolve()
    title = str(request["title"]).strip()
    render = request["render"]
    dpi = int(render["dpi"])
    quality = int(render["jpeg_quality"])

    if pdf_path.suffix.lower() != ".pdf" or not pdf_path.is_file():
        raise ValueError(f"Source is not a readable PDF: {pdf_path}")
    if not title:
        raise ValueError("Title must not be blank")
    if not MIN_DPI <= dpi <= MAX_DPI:
        raise ValueError(f"DPI must be between {MIN_DPI} and {MAX_DPI}")
    if not MIN_JPEG_QUALITY <= quality <= MAX_JPEG_QUALITY:
        raise ValueError(
            f"JPEG quality must be between {MIN_JPEG_QUALITY} and {MAX_JPEG_QUALITY}"
        )

    return pdf_path, database_path, title, dpi, quality


def convert(request: dict[str, Any]) -> dict[str, Any]:
    pdf_path, database_path, title, dpi, quality = check_request(request)
    book_id = str(uuid.uuid4())

    with pymupdf.open(pdf_path) as document:
        page_count = document.page_count
        if page_count < 1:
            raise ValueError("PDF has no pages")

        outline = document.get_toc(simple=True)
        if not outline:
            raise ImportFailure(OUTLINE_REQUIRED, "PDF outline is required.")

        connection = sqlite3.connect(database_path, isolation_level=None)
        try:
            connection.execute("PRAGMA foreign_keys = ON")
            connection.execute(f"PRAGMA busy_timeout = {BUSY_TIMEOUT_MS}")
            connection.execute("BEGIN")
            try:
                connection.execute(
                    "INSERT INTO books(book_id,title,page_count) VALUES(?,?,?)",
                    (book_id, title, page_count),
                )
                connection.executemany(
                    "INSERT INTO book_pages(book_id,page_index,mime_type,image_data) "
                    "VALUES(?,?,?,?)",
                    page_rows(document, book_id, dpi, quality),
                )
                try:
                    connection.executemany(
                        "INSERT INTO book_outline("
                        "book_id,outline_index,title,page_index"
                        ") VALUES(?,?,?,?)",
                        outline_rows(outline, book_id),
                    )
                except sqlite3.IntegrityError as exc:
                    raise ImportFailure(
                        OUTLINE_UNUSABLE,
                        "PDF outline cannot be stored.",
                    ) from exc
                connection.commit()
            except Exception:
                connection.rollback()
                raise
        finally:
            connection.close()

    return {"book_id": book_id, "title": title, "page_count": page_count}


def main() -> int:
    try:
        request = json.load(sys.stdin)
        result = convert(request)
    except ImportFailure as exc:
        json.dump(
            {"code": exc.code, "message": exc.message},
            sys.stderr,
            ensure_ascii=False,
        )
        sys.stderr.write("\n")
        return 1
    except (
        json.JSONDecodeError,
        KeyError,
        OSError,
        RuntimeError,
        sqlite3.Error,
        TypeError,
        ValueError,
    ):
        json.dump(
            {"code": "conversion_failed", "message": "PDF import failed."},
            sys.stderr,
        )
        sys.stderr.write("\n")
        return 1

    json.dump(result, sys.stdout, ensure_ascii=False)
    sys.stdout.write("\n")
    return 0


if __name__ == "__main__":
    sys.exit(main())
