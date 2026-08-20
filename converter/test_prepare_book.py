import importlib.util
import tempfile
import unittest
from pathlib import Path
from unittest import mock

from pypdf import PdfWriter


MODULE_PATH = Path(__file__).with_name("prepare_book.py")
SPEC = importlib.util.spec_from_file_location("prepare_book", MODULE_PATH)
assert SPEC and SPEC.loader
converter = importlib.util.module_from_spec(SPEC)
SPEC.loader.exec_module(converter)


def make_pdf(path: Path):
    writer = PdfWriter()
    for _ in range(3):
        writer.add_blank_page(width=100, height=100)
    chapter = writer.add_outline_item("Chapter", 0)
    writer.add_outline_item("Topic", 1, parent=chapter)
    with path.open("wb") as stream:
        writer.write(stream)


class ConverterTest(unittest.TestCase):
    def test_writes_contiguous_images_and_compact_toc(self):
        with tempfile.TemporaryDirectory() as temp:
            root = Path(temp)
            source = root / "book.pdf"
            output = root / "output"
            output.mkdir()
            make_pdf(source)

            def fake_run(command, check, creationflags):
                del check, creationflags
                prefix = Path(command[-1])
                for page in range(1, 4):
                    prefix.with_name(f"{prefix.name}-{page}.jpg").write_bytes(b"jpeg")

            with mock.patch.object(converter, "locate_pdftoppm", return_value="pdftoppm"), mock.patch.object(converter.subprocess, "run", side_effect=fake_run):
                pages, entries = converter.prepare(source, output, 200, 90, None)

            self.assertEqual((pages, entries), (3, 2))
            self.assertEqual(sorted(path.name for path in output.glob("page-*.jpg")), ["page-0001.jpg", "page-0002.jpg", "page-0003.jpg"])
            self.assertEqual((output / "toc.csv").read_text(encoding="utf-8"), "position,depth,title,page_index\n1,0,Chapter,1\n2,1,Topic,2\n")


if __name__ == "__main__":
    unittest.main()
