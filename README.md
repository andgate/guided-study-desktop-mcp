# Guided Study Desktop MCP

<img src="cmd/guided-study/assets/tray-icon.png" alt="Guided Study open-book tray icon" width="96">

Guided Study is a local Windows application that gives an LLM deterministic tools for preparing books, navigating rendered pages, persisting independent study sessions, and maintaining revisioned flashcard decks. The LLM teaches; this service owns storage and page delivery.

The version-one contract is implemented as:

- one Go/Fyne system-tray application;
- a Streamable HTTP MCP endpoint at `http://127.0.0.1:7331/mcp`;
- one canonical SQLite database under `%LOCALAPPDATA%\GuidedStudy`;
- a bundled converter executable that renders PDFs through PyMuPDF and emits temporary page images plus `toc.csv`;
- 27 focused MCP tools covering books, reading sessions, decks, and immutable card revisions;
- the MCP-native Noggin skill in [`noggin-plugin`](noggin-plugin/SKILL.md).

The source PDF is never modified, deleted, or retained in SQLite. Failed conversion or validation cannot expose a partial book.

## Prerequisites

- Windows 11
- [Git LFS](https://git-lfs.com/) for binary application assets
- Go 1.25 or newer for development
- [Task](https://taskfile.dev/) v3
- [uv](https://docs.astral.sh/uv/) 0.12.5 or newer
- Python 3.12 for development and builds
- A GCC toolchain for building Fyne on Windows

The build packages `converter\prepare_book.py`, Python, and PyMuPDF into `bin\pdf-converter.exe`. The application runs that executable directly and does not require a system Python installation. `--converter` overrides its path.

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

The build produces `bin\guided-study.exe` and `bin\pdf-converter.exe`. Keep them together when distributing the application.

The status window may be closed without stopping the server. Reopen it from the tray. Choose **Quit** from the tray for graceful HTTP and SQLite shutdown.

For terminal-only development:

```powershell
.\bin\guided-study.exe --headless
```

Useful options:

```text
--listen 127.0.0.1:7331
--database C:\path\guided-study.db
--converter C:\path\pdf-converter.exe
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

Run the executable from the repository root so it can use `converter\prepare_book.py`, or pass that script's path with `--converter`. Tests cover transactional storage, independent cursors, immutable card history, atomic batch deletion, MCP discovery/schemas/annotations, structured conflicts, and image content. The Python test verifies contiguous page output and compact TOC generation.

## Connect an MCP client

Use Streamable HTTP and the endpoint shown in the status window. MCP Inspector can connect to:

```text
http://127.0.0.1:7331/mcp
```

The health probe is `GET /healthz`.

The `import_book.file_reference` input is the absolute path to the local PDF supplied by the host.

## Install Noggin locally

Connecting the MCP server exposes storage and navigation tools, but it does not install the teaching workflow. Install or refresh Noggin for the current Windows user with:

```powershell
task plugin:install
```

The task runs `python scripts/install_plugin.py` in the global Python environment. It validates the skill from `noggin-plugin` and installs it at `%USERPROFILE%\.agents\skills\noggin`, making it available outside this repository. The installer does not require the Codex CLI.

Configure the MCP server separately in ChatGPT Desktop or Codex as a Streamable HTTP server named `guided_study` with URL `http://127.0.0.1:7331/mcp`. Start a new conversation after installing or refreshing the plugin so the client discovers the updated skill.

## Data and recovery

SQLite contains book metadata, TOC entries, rendered page BLOBs, named session cursors, deck metadata, and immutable card revisions. Use any SQLite viewer for inspection. Explicit exports are not part of version one.

Deleting a session, card, deck, or book is immediate and permanent. Destructive MCP annotations accurately mark those operations. Deleting an imported book never touches its original PDF.

See [ARCHITECTURE.md](docs/ARCHITECTURE.md), [SCHEMA.md](docs/SCHEMA.md), and [MCP.md](docs/MCP.md) for the implementation boundaries and version-one contracts.
