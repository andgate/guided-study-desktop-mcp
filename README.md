# Guided Study Desktop MCP

<img src="cmd/guided-study/assets/tray-icon.png" alt="Guided Study open-book tray icon" width="96">

Guided Study is a local Windows application that gives an LLM deterministic tools for preparing books, navigating rendered pages, persisting independent study sessions, and maintaining revisioned flashcard decks. The LLM teaches; this service owns storage and page delivery.

The version-one contract is implemented as:

- one Go/Fyne system-tray application;
- a Streamable HTTP MCP endpoint at `http://127.0.0.1:7331/mcp`;
- one canonical SQLite database under `%LOCALAPPDATA%\GuidedStudy`;
- a Python subprocess that renders PDFs through PyMuPDF and emits temporary page images plus `toc.csv`;
- 30 focused MCP tools covering books, reading sessions/progress, decks, and immutable card revisions;
- the MCP-native Teach skill in [`teach/skills/teach`](teach/skills/teach/SKILL.md).

The source PDF is never modified, deleted, or retained in SQLite. Failed conversion or validation cannot expose a partial book.

## Prerequisites

- Windows 11
- [Git LFS](https://git-lfs.com/) for binary application assets
- Go 1.25 or newer for development
- [Task](https://taskfile.dev/) v3
- [uv](https://docs.astral.sh/uv/) 0.12.5 or newer
- Python 3.12
- A GCC toolchain for building Fyne on Windows

The application runs `converter\prepare_book.py` directly from the repository. It looks for the script next to the executable and under the current working directory; `--converter` overrides that path. Python is discovered next to the executable or on `PATH`; `--python` overrides its path.

After cloning, initialize Git LFS and materialize the application assets before building:

```powershell
git lfs install
git lfs pull
```

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
--headless
```

The server always refuses non-loopback addresses. It has no authentication and is strictly local.

## Test

```powershell
task test
```

Format the Go and Python source files with:

```powershell
task format
```

Run the executable from the repository root so it can use `converter\prepare_book.py`, or pass that script's path with `--converter`. Tests cover transactional storage, independent cursors, guarded progress logs, immutable card history, atomic batch deletion, MCP discovery/schemas/annotations, structured conflicts, and image content. The Python test verifies contiguous page output and compact TOC generation.

## Connect an MCP client

Use Streamable HTTP and the endpoint shown in the status window. MCP Inspector can connect to:

```text
http://127.0.0.1:7331/mcp
```

The health probe is `GET /healthz`.

The `import_book.file_reference` bridge deliberately supports local filesystem paths and `file://` URLs first. Other opaque ChatGPT upload references return `file_reference_unsupported` until actual host behavior is verified. This is the one integration question intentionally left open by the contract.

## Install Teach locally

Connecting the MCP server exposes storage and navigation tools, but it does not install the teaching workflow. Install or refresh Teach for the current Windows user with:

```powershell
task plugin:install
```

The task runs `python scripts/install_plugin.py` in the global Python environment. It validates the skill from `teach\skills\teach` and installs it at `%USERPROFILE%\.agents\skills\teach`, making it available outside this repository. The installer does not require the Codex CLI.

Configure the MCP server separately in ChatGPT Desktop or Codex as a Streamable HTTP server named `guided_study` with URL `http://127.0.0.1:7331/mcp`. Start a new conversation after installing or refreshing the plugin so the client discovers the updated skill.

## Data and recovery

SQLite contains book metadata, TOC entries, rendered page BLOBs, named session cursors/logs, deck metadata, and immutable card revisions. Use any SQLite viewer for inspection. Explicit exports are not part of version one.

Deleting a session, log tail, card, deck, or book is immediate and permanent. Destructive MCP annotations accurately mark those operations. Deleting an imported book never touches its original PDF.

See [ARCHITECTURE.md](docs/ARCHITECTURE.md), [SCHEMA.md](docs/SCHEMA.md), and [MCP.md](docs/MCP.md) for the implementation boundaries and version-one contracts.
