# Guided Study Architecture

## Runtime

`cmd/guided-study` owns the process lifecycle. It opens SQLite, constructs the importer and MCP service, starts one `net/http` server, then runs either the Fyne tray loop or a headless signal loop. Tray Quit cancels the process context, shuts down HTTP with a deadline, and closes SQLite.

```text
ChatGPT / MCP client
        |
        | Streamable HTTP /mcp
        v
Go MCP service ----> SQLite (canonical state and rendered page BLOBs)
        |
        | JSON request and result
        v
Bundled converter (Python + PyMuPDF) ----> SQLite transaction
```

The converter opens the PDF and SQLite database. It extracts the PDF outline
and inserts the book, lazily rendered page BLOBs, and flat outline in one
transaction. Failed extraction, rendering, or insertion rolls back the complete
import. No page-image or CSV staging files are created.

## Packages

- `internal/appconfig`: flags, default path discovery, and loopback binding policy.
- `internal/importer`: converter request encoding, subprocess execution, diagnostics, and result decoding.
- `internal/store`: DDL, domain models, runtime transactions, and canonical study-data access.
- `internal/mcpserver`: typed tool schemas, descriptions, annotations, structured results/errors, and image content.
- `converter`: the transactional PDF import writer, frozen as `pdf-converter.exe`.
- `noggin-plugin`: the Noggin workflow that consumes the MCP API.

## State boundaries

A book has a stable generated UUID, ordered pages, a flat extracted outline,
zero or more named sessions, and zero or more caller-named decks.
There is no selected book or implicit session.

Each session stores a page-window origin, physical checkpoint page, and nullable
agent-supplied heading. Batch windows are computed from the origin and are never
stored. `continue_reading` saves a checkpoint and preloads the following batch
atomically. Deck metadata uses its own revision. Card updates insert immutable
revisions; deletion removes the full logical card history only after checking
the latest revision.

Go uses one application SQLite connection. Each import opens one converter
connection with foreign keys enabled and a busy timeout. The importer serializes
converter processes so only one import transaction writes at a time. IDs and
display names have separate roles: UUIDs and `deck_id` values are stable
identity; book titles, deck titles, and session names are editable metadata.

## MCP behavior

All ordinary results include concise text plus `structuredContent`.
`read_pages`, `create_session`, `goto_page`, and `continue_reading`
additionally include labeled MCP image content blocks without duplicating image
bytes in structured data.

Every tool advertises inferred or explicit JSON input/output schemas. Unknown fields are rejected by schema validation. Read-only and destructive annotations match actual behavior; every tool is closed-world because version one affects only the local library.

Domain errors use stable codes from `MCP.md` and return `isError: true`, short text, and a structured `{code,message,details}` object. Conflicts expose actual state needed for reconciliation and never mutate state.

## Security

Version one is a trusted single-user local service with no authentication. It only binds to loopback addresses and cannot be exposed to other computers through configuration. It is built and run locally from this repository; there is no installation or deployment stage.

The Fyne executable is the only long-running application. Python is conversion
machinery invoked by Go and writes canonical import data transactionally.
