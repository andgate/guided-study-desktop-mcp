import importlib.util
import sqlite3
import tempfile
import unittest
from pathlib import Path

import pymupdf

MODULE_PATH = Path(__file__).with_name("prepare_book.py")
SCHEMA_PATH = MODULE_PATH.parent.parent / "internal" / "store" / "schema.sql"
SPEC = importlib.util.spec_from_file_location("prepare_book", MODULE_PATH)
assert SPEC and SPEC.loader
converter = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(converter)


def make_pdf(path: Path, with_outline: bool = True):
    document = pymupdf.open()
    for _ in range(3):
        document.new_page(width=100, height=100)
    if with_outline:
        document.set_toc(
            [
                [1, "Opening", 1],
                [1, 'Names, "Quotes"', 3],
            ]
        )
    document.save(path)
    document.close()


def make_database(path: Path):
    connection = sqlite3.connect(path)
    connection.executescript(SCHEMA_PATH.read_text(encoding="utf-8"))
    connection.close()


def request(pdf: Path, database: Path):
    return {
        "database_path": str(database),
        "file_reference": str(pdf),
        "title": "Test Book",
        "render": {
            "dpi": converter.DEFAULT_DPI,
            "jpeg_quality": converter.DEFAULT_JPEG_QUALITY,
        },
    }


class ConverterTest(unittest.TestCase):
    def test_stores_book(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.pdf"
            database = root / "book.db"
            make_pdf(source)
            make_database(database)

            result = converter.convert(request(source, database))

            connection = sqlite3.connect(database)
            book = connection.execute(
                "SELECT book_id,title,page_count FROM books"
            ).fetchone()
            pages = connection.execute(
                "SELECT page_index,mime_type,length(image_data) "
                "FROM book_pages ORDER BY page_index"
            ).fetchall()
            outline = connection.execute(
                "SELECT outline_index,title,page_index "
                "FROM book_outline ORDER BY outline_index"
            ).fetchall()
            connection.close()

            self.assertEqual(book, (result["book_id"], "Test Book", 3))
            self.assertEqual(
                [page[:2] for page in pages],
                [(1, "image/jpeg"), (2, "image/jpeg"), (3, "image/jpeg")],
            )
            self.assertTrue(all(page[2] > 0 for page in pages))
            self.assertEqual(
                outline,
                [(0, "Opening", 1), (1, 'Names, "Quotes"', 3)],
            )

    def test_requires_outline(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.pdf"
            database = root / "book.db"
            make_pdf(source, with_outline=False)
            make_database(database)

            with self.assertRaises(converter.ImportFailure) as raised:
                converter.convert(request(source, database))

            self.assertEqual(raised.exception.code, converter.OUTLINE_REQUIRED)

            connection = sqlite3.connect(database)
            count = connection.execute("SELECT count(*) FROM books").fetchone()[0]
            connection.close()
            self.assertEqual(count, 0)

    def test_rolls_back_unusable_outline(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.pdf"
            database = root / "book.db"
            make_pdf(source)
            make_database(database)
            connection = sqlite3.connect(database)
            connection.execute(
                "CREATE TRIGGER reject_outline BEFORE INSERT ON book_outline "
                "BEGIN SELECT RAISE(ABORT, 'reject outline'); END"
            )
            connection.commit()
            connection.close()

            with self.assertRaises(converter.ImportFailure) as raised:
                converter.convert(request(source, database))

            self.assertEqual(raised.exception.code, converter.OUTLINE_UNUSABLE)

            connection = sqlite3.connect(database)
            count = connection.execute("SELECT count(*) FROM books").fetchone()[0]
            connection.close()
            self.assertEqual(count, 0)


if __name__ == "__main__":
    unittest.main()
