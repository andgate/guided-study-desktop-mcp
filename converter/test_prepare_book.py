import importlib.util
import sqlite3
import tempfile
import unittest
import zipfile
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


def make_epub(path: Path):
    package = (
        '<?xml version="1.0"?><package xmlns="http://www.idpf.org/2007/opf" '
        'version="3.0" unique-identifier="i"><metadata '
        'xmlns:dc="http://purl.org/dc/elements/1.1/"><dc:identifier id="i">x'
        "</dc:identifier><dc:title>Test</dc:title><dc:language>en</dc:language>"
        '</metadata><manifest><item id="n" href="nav.xhtml" '
        'media-type="application/xhtml+xml" properties="nav"/><item id="c1" '
        'href="c1.xhtml" media-type="application/xhtml+xml"/><item id="c2" '
        'href="c2.xhtml" media-type="application/xhtml+xml"/></manifest>'
        '<spine><itemref idref="c1"/><itemref idref="c2"/></spine></package>'
    )
    nav = (
        '<?xml version="1.0"?><html xmlns="http://www.w3.org/1999/xhtml" '
        'xmlns:epub="http://www.idpf.org/2007/ops"><body><nav epub:type="toc">'
        '<ol><li><a href="c1.xhtml">Opening</a></li>'
        '<li><a href="c2.xhtml">Closing</a></li></ol></nav></body></html>'
    )
    with zipfile.ZipFile(path, "w") as archive:
        archive.writestr("mimetype", "application/epub+zip", zipfile.ZIP_STORED)
        archive.writestr(
            "META-INF/container.xml",
            '<?xml version="1.0"?><container version="1.0" '
            'xmlns="urn:oasis:names:tc:opendocument:xmlns:container"><rootfiles>'
            '<rootfile full-path="c.opf" '
            'media-type="application/oebps-package+xml"/></rootfiles></container>',
        )
        archive.writestr("c.opf", package)
        archive.writestr("nav.xhtml", nav)
        for name, heading in (("c1", "Opening"), ("c2", "Closing")):
            archive.writestr(
                f"{name}.xhtml",
                '<?xml version="1.0"?><html xmlns="http://www.w3.org/1999/xhtml">'
                f"<body><h1>{heading}</h1><p>Body text.</p></body></html>",
            )


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
            book = connection.execute("SELECT book_id,title,page_count FROM books").fetchone()
            pages = connection.execute(
                "SELECT page_index,mime_type,length(image_data) FROM book_pages ORDER BY page_index"
            ).fetchall()
            outline = connection.execute(
                "SELECT outline_index,title,page_index FROM book_outline ORDER BY outline_index"
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

    def test_stores_reflowable_book(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.epub"
            database = root / "book.db"
            make_epub(source)
            make_database(database)

            result = converter.convert(request(source, database))

            connection = sqlite3.connect(database)
            page_count = connection.execute("SELECT page_count FROM books").fetchone()[0]
            pages = connection.execute("SELECT count(*) FROM book_pages").fetchone()[0]
            outline = connection.execute(
                "SELECT title FROM book_outline ORDER BY outline_index"
            ).fetchall()
            connection.close()

            self.assertEqual(page_count, result["page_count"])
            self.assertEqual(pages, page_count)
            self.assertEqual([row[0] for row in outline], ["Opening", "Closing"])

    def test_rejects_unsupported_format(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.txt"
            database = root / "book.db"
            source.write_text("not a book", encoding="utf-8")
            make_database(database)

            with self.assertRaises(ValueError):
                converter.convert(request(source, database))


if __name__ == "__main__":
    unittest.main()
