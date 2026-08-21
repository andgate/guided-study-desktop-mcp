import importlib.util
import tempfile
import unittest
from pathlib import Path

import pymupdf

MODULE_PATH = Path(__file__).with_name("prepare_book.py")
SPEC = importlib.util.spec_from_file_location("prepare_book", MODULE_PATH)
assert SPEC and SPEC.loader
converter = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(converter)


def make_pdf(path: Path):
    document = pymupdf.open()
    for _ in range(3):
        document.new_page(width=100, height=100)
    document.set_toc([[1, "Chapter", 1], [2, "Topic", 2]])
    document.save(path)
    document.close()


class ConverterTest(unittest.TestCase):
    def test_writes_contiguous_images_and_compact_toc(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.pdf"
            output = root / "output"
            output.mkdir()
            make_pdf(source)

            pages, entries = converter.prepare(
                source,
                output,
                converter.DEFAULT_DPI,
                converter.DEFAULT_JPEG_QUALITY,
            )

            self.assertEqual((pages, entries), (3, 2))
            self.assertEqual(
                sorted(path.name for path in output.glob("page-*.jpg")),
                ["page-0001.jpg", "page-0002.jpg", "page-0003.jpg"],
            )
            self.assertEqual(
                (output / "toc.csv").read_text(encoding="utf-8"),
                "position,depth,title,page_index\n1,0,Chapter,1\n2,1,Topic,2\n",
            )


if __name__ == "__main__":
    unittest.main()
