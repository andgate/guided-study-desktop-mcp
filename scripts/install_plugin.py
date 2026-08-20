from __future__ import annotations

import shutil
from pathlib import Path

import yaml

SKILL_NAME = "teach"
REPO_ROOT = Path(__file__).resolve().parents[1]
SOURCE_SKILL = REPO_ROOT / "teach-plugin"
USER_SKILLS = Path.home() / ".agents" / "skills"
DESTINATION = USER_SKILLS / SKILL_NAME


def validate_skill(path: Path) -> None:
    skill_path = path / "SKILL.md"
    parts = skill_path.read_text(encoding="utf-8").split("---", 2)
    if len(parts) != 3:
        raise ValueError(f"Missing YAML frontmatter in {skill_path}.")
    frontmatter = yaml.safe_load(parts[1])
    if frontmatter.get("name") != SKILL_NAME or not frontmatter.get("description"):
        raise ValueError(f"Invalid skill frontmatter in {skill_path}.")

    metadata_path = path / "agents" / "openai.yaml"
    metadata = yaml.safe_load(metadata_path.read_text(encoding="utf-8"))
    dependencies = metadata.get("dependencies", {}).get("tools", [])
    if not any(
        item.get("type") == "mcp" and item.get("value") == "guided_study"
        for item in dependencies
    ):
        raise ValueError(
            f"{metadata_path} must declare the guided_study MCP dependency."
        )


def install_plugin() -> None:
    validate_skill(SOURCE_SKILL)
    USER_SKILLS.mkdir(parents=True, exist_ok=True)

    staging = USER_SKILLS / f".{SKILL_NAME}.installing"
    if staging.exists():
        shutil.rmtree(staging)
    shutil.copytree(SOURCE_SKILL, staging)
    validate_skill(staging)

    if DESTINATION.exists():
        shutil.rmtree(DESTINATION)
    staging.rename(DESTINATION)

    print(f"Installed {SKILL_NAME} at {DESTINATION}")
    print(
        "Start a new Codex or ChatGPT Desktop conversation before testing the study workflow."
    )


if __name__ == "__main__":
    install_plugin()
