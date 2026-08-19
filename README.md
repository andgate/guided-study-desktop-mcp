# Guided Study Desktop MCP

Guided Study is a local Windows application that gives an LLM deterministic tools for preparing books, navigating rendered pages, persisting independent study sessions, and maintaining revisioned flashcard decks. The LLM teaches; this service owns storage and page delivery.

The version-one contract is implemented as:

- one Go/Fyne system-tray application;
- a Streamable HTTP MCP endpoint at `http://127.0.0.1:7331/mcp`;
- one canonical SQLite database under `%LOCALAPPDATA%\GuidedStudy`;
- a Python subprocess that renders PDFs through Poppler and emits temporary page images plus `toc.csv`;
- 30 focused MCP tools covering books, reading sessions/progress, decks, and immutable card revisions;
- an MCP-native teaching skill in [`plugin/skills/guide-book-study`](plugin/skills/guide-book-study/SKILL.md).

The source PDF is never modified, deleted, or retained in SQLite. Failed conversion or validation cannot expose a partial book.

## Prerequisites

- Windows 11
- Go 1.25 or newer for development
- [Task](https://taskfile.dev/) v3
- [uv](https://docs.astral.sh/uv/) 0.12.5 or newer
- Poppler's `pdftoppm.exe`
- A GCC toolchain for building Fyne on Windows

The application runs `converter\prepare_book.py` directly from the repository. It looks for the script next to the executable and under the current working directory; `--converter` overrides that path. Python and Poppler are discovered next to the executable, in the Codex bundled runtime, or on `PATH`; `--python` and `--pdftoppm` override their paths.

## Build and run

```powershell
task build
.\bin\guided-study.exe
```

The status window may be closed without stopping the server. Reopen it from the tray. Choose **Quit** from the tray for graceful HTTP and SQLite shutdown.

For terminal-only development:

```powershell
.\bin\guided-study.exe --headless
```

Useful options:

```text
--listen 127.0.0.1:7331
--database C:\path\guided-study.db
--python C:\path\python.exe
--converter C:\path\prepare_book.py
--pdftoppm C:\path\pdftoppm.exe
--headless
```

The server always refuses non-loopback addresses. It has no authentication and is strictly local.

## Test

```powershell
task setup
task test
```

Format the Go and Python source files with:

```powershell
task format
```

## VS Code setup

Run the project setup once from the repository root:

```powershell
task setup
```

Then configure VS Code:

1. Open this repository as the workspace folder.
2. Open the Extensions view and enter `@recommended` in its search box.
3. Install all three workspace recommendations: Go, Python, and Ruff.
4. Run **Developer: Reload Window** from the Command Palette.
5. Check the Python interpreter shown in the status bar. If it is not `converter\.venv\Scripts\python.exe`, run **Python: Select Interpreter** and choose that file.

The checked-in workspace settings format Go with `gofmt` and Python with Ruff when you save those source files. `task format:check` checks those two source languages.

Run the executable from the repository root so it can use `converter\prepare_book.py`, or pass that script's path with `--converter`. Tests cover transactional storage, independent cursors, guarded progress logs, immutable card history, atomic batch deletion, MCP discovery/schemas/annotations, structured conflicts, and image content. The Python test verifies contiguous page output and compact TOC generation.

## Connect an MCP client

Use Streamable HTTP and the endpoint shown in the status window. MCP Inspector can connect to:

```text
http://127.0.0.1:7331/mcp
```

The health probe is `GET /healthz`.

The `import_book.file_reference` bridge deliberately supports local filesystem paths and `file://` URLs first. Other opaque ChatGPT upload references return `file_reference_unsupported` until actual host behavior is verified. This is the one integration question intentionally left open by the contract.

## Data and recovery

SQLite contains book metadata, TOC entries, rendered page BLOBs, named session cursors/logs, deck metadata, and immutable card revisions. Use any SQLite viewer for inspection. Explicit exports are not part of version one.

Deleting a session, log tail, card, deck, or book is immediate and permanent. Destructive MCP annotations accurately mark those operations. Deleting an imported book never touches its original PDF.

See [ARCHITECTURE.md](ARCHITECTURE.md), [SCHEMA.md](SCHEMA.md), and [MCP.md](MCP.md) for the implementation boundaries and frozen contracts.
