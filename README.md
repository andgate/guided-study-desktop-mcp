# Guided Study Desktop MCP

<img src="cmd/noggin-mcp/assets/tray-icon.png" alt="Guided Study open-book tray icon" width="96">

Guided Study is a local desktop application that gives an LLM deterministic tools for preparing books, navigating rendered pages, persisting independent study sessions, and maintaining revisioned flashcard decks. The LLM teaches; this service owns storage and page delivery.

The version-one contract is implemented as:

- one Go/Fyne system-tray application;
- a Streamable HTTP MCP endpoint at `http://127.0.0.1:7331/mcp`;
- one canonical SQLite database in the current user's data directory;
- a bundled converter executable that imports PDF and EPUB books transactionally through PyMuPDF and SQLite;
- 25 focused MCP tools covering books, reading sessions, decks, and immutable card revisions;
- the MCP-native Noggin skill in [`noggin-plugin`](noggin-plugin/SKILL.md).

The source book is never modified, deleted, or retained in SQLite. Failed
rendering or insertion cannot expose a partial book.

## Prerequisites

- Windows 11 or macOS 13 or newer
- [Git LFS](https://git-lfs.com/) for binary application assets
- Go 1.25 or newer for development
- [Task](https://taskfile.dev/) v3
- [uv](https://docs.astral.sh/uv/) 0.12.5 or newer
- Python 3.12 for development and builds
- A C toolchain for building Fyne: GCC on Windows, Xcode Command Line Tools on macOS

The build packages `converter/prepare_book.py`, Python, and PyMuPDF into a single
converter executable in `bin`. The application runs that executable directly and
does not require a system Python installation. `--converter` overrides its path.

After cloning, initialize Git LFS and materialize the application assets before building:

```bash
git lfs install
git lfs pull
```

## Build and run

```bash
task build
```

The build produces `bin/noggin-mcp` and `bin/pdf-converter`, both with a `.exe`
suffix on Windows. Keep them together when distributing the application.

On macOS the build also assembles `bin/Noggin MCP.app`, which carries its own
copy of both executables. Open it from Finder, or drag it to `/Applications`. The
bundle is signed ad hoc, so it runs on the machine that built it without
notarization.

Run the executable from the repository root:

```bash
./bin/noggin-mcp
```

On Windows:

```powershell
.\bin\noggin-mcp.exe
```

The status window may be closed without stopping the server. Reopen it from the tray. Choose **Quit** from the tray for graceful HTTP and SQLite shutdown.

For terminal-only development:

```bash
./bin/noggin-mcp --headless
```

Useful options:

```text
--listen 127.0.0.1:7331
--database /path/to/guided-study.db
--converter /path/to/pdf-converter
--headless
```

The server always refuses non-loopback addresses. It has no authentication and is strictly local.

## Test

```bash
task test
```

Format the Go and Python source files with:

```bash
task format
```

Run the executable from the repository root so it can use
`converter/prepare_book.py`, or pass that script's path with `--converter`.
Tests cover transactional imports, page batching, independent sessions,
window resets, immutable card history, atomic batch deletion, MCP schemas,
structured errors, and image content.

## Connect an MCP client

Use Streamable HTTP and the endpoint shown in the status window. MCP Inspector can connect to:

```text
http://127.0.0.1:7331/mcp
```

The health probe is `GET /healthz`.

The `import_book.file_reference` input is the absolute path to the local PDF or EPUB supplied by the host.

EPUB is reflowable, so its pages are produced by laying the book out at a fixed
page size. Those page numbers belong to this library and do not match a printed
edition. Navigate reflowable books by heading.

## Install Noggin locally

Install or refresh Noggin for the current user with:

```bash
task plugin:install
```

The task copies `noggin-plugin/skills/noggin` to `~/.agents/skills/noggin`.

Configure the MCP server separately in ChatGPT Desktop or Codex as a Streamable HTTP server named `noggin_mcp` with URL `http://127.0.0.1:7331/mcp`.

## Data and recovery

SQLite contains book metadata, extracted outlines, rendered page BLOBs, durable
session progress, deck metadata, and immutable card revisions. The database
lives under `%LOCALAPPDATA%\GuidedStudy` on Windows and
`~/Library/Application Support/GuidedStudy` on macOS. Use any SQLite viewer for
inspection. Explicit exports are not part of version one.

Deleting a session, card, deck, or book is immediate and permanent. Destructive MCP annotations accurately mark those operations. Deleting an imported book never touches its original file.

See [ARCHITECTURE.md](docs/ARCHITECTURE.md), [SCHEMA.md](docs/SCHEMA.md), and [MCP.md](docs/MCP.md) for the implementation boundaries and version-one contracts.
