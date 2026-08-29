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
        | bounded subprocess per import
        v
Bundled converter (Python + PyMuPDF) ----> temporary images + toc.csv
```

The temporary conversion directory is removed on success and failure. Go treats every staged file as untrusted, validates names, continuity, MIME types, TOC shape/references, and size limits, then ingests the entire book in one transaction.

## Packages

- `internal/appconfig`: flags, default path discovery, and loopback binding policy.
- `internal/importer`: local file-reference resolution, converter subprocess execution, and staging validation.
- `internal/store`: DDL, domain models, validation, transactions, optimistic concurrency, and all canonical data access.
- `internal/mcpserver`: typed tool schemas, descriptions, annotations, structured results/errors, and image content.
- `converter`: the PyMuPDF renderer, frozen as `pdf-converter.exe` and invoked directly by the local application.
- `noggin-plugin`: the Noggin workflow that consumes the MCP API.

## State boundaries

A book has a stable generated UUID, ordered pages, optional ordered TOC entries, zero or more named sessions, and zero or more caller-named decks. There is no selected book or implicit session.

Session navigation checks `expected_page_index` in the same transaction as the cursor update. Deck metadata uses its own revision. Card updates insert immutable revisions; deletion removes the full logical card history only after checking the latest revision.

SQLite uses one application connection, foreign keys on, a busy timeout, and bounded transactions. IDs and titles have separate roles: UUIDs and `deck_id` values are stable identity; titles and session names are editable display metadata.

## MCP behavior

All ordinary results include concise text plus `structuredContent`. `current_page`, `next_page`, `prev_page`, and `goto_page` additionally include one MCP image content block without duplicating image bytes in structured data.

Every tool advertises inferred or explicit JSON input/output schemas. Unknown fields are rejected by schema validation. Read-only and destructive annotations match actual behavior; every tool is closed-world because version one affects only the local library.

Domain errors use stable codes from `MCP.md` and return `isError: true`, short text, and a structured `{code,message,details}` object. Conflicts expose actual state needed for reconciliation and never mutate state.

## Security

Version one is a trusted single-user local service with no authentication. It only binds to loopback addresses and cannot be exposed to other computers through configuration. It is built and run locally from this repository; there is no installation or deployment stage.

The Fyne executable is the only long-running application. Python is conversion machinery invoked by Go, not a service, controller, or canonical data owner.
