from pathlib import Path
from shutil import copytree, rmtree

source = Path(__file__).parents[1] / "noggin-plugin" / "skills" / "noggin"
destination = Path.home() / ".agents" / "skills" / "noggin"

rmtree(destination, ignore_errors=True)
copytree(source, destination)
