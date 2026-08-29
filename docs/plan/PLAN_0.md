# Guided Study System Plan (IMPLEMENTED)

## Status of this document

This document records the current replacement plan for Guided Study after the first full study-workflow experiment. It is intentionally separate from `README.md`, `ARCHITECTURE.md`, `study-agent.md`, and the existing skills, because those files still describe the old Google Drive, split-PDF, card-inbox, archive, and Quizlet workflow.

The labels below matter:

- **Decided** means the direction was explicitly established in discussion.
- **Current direction** means it is the leading design but still needs refinement.
- **Open** means no decision has been made yet.

This is a planning document, not an assertion that the repository already implements the design.

## Product goal

Build a reliable study service that an LLM can use as its operational backend while the LLM remains responsible for teaching.

The service should:

- prepare books for visual, page-by-page study;
- let an agent navigate a book without searching a general-purpose file store;
- persist progress independently for every book;
- let an agent progressively construct, inspect, revise, and export flashcard decks;
- work locally first through MCP;
- later be deployable as a hosted service for multiple LLM clients;
- keep its canonical study state locally inspectable and recoverable.

The agent is the primary user interface. The service supplies deterministic storage, navigation, and conversion operations. It does not attempt to reproduce the agent's reasoning or pedagogy.

## Settled architectural decisions

### 1. The MCP server will be written in Go

**Decided.**

Go is the practical fit for the main service:

- the service is mostly local storage operations, process execution, and a small tool API;
- Go has an official MCP SDK;
- it can expose streamable HTTP on a configured local address and port for the ChatGPT plugin;
- it may also expose STDIO later for clients that launch local MCP processes;
- the same implementation can later expose the appropriate remote MCP transport;
- deployment can remain a small compiled service.

There is no requirement to introduce TypeScript, Node.js, FastAPI, or a Python application server.

### 2. Python remains only PDF conversion machinery

**Decided.**

The existing Python preparation code will be copied into the new service project and adapted there. The historical prompt/plugin repository keeps its original version. In the new project, the converter remains a program that accepts a PDF and writes temporary prepared book artifacts to a destination. It does not become a server and does not own book management.

The Go MCP server runs the Python converter as a subprocess when a book is imported. Go does not reimplement PDF rendering or add a duplicate "import controller" around the same logic. Its role is to expose the operation, provide the destination, observe the result, and make the prepared book available through the rest of the MCP API.

In a later hosted deployment, Go and the Python converter can live in the same container or task. Python remains an internal conversion dependency, not an independent network service.

### 3. SQLite is the canonical local store

**Decided.**

Version one stores all canonical library data in one local SQLite database, including:

- book metadata;
- ordered table-of-contents entries;
- rendered page images as BLOBs with their MIME type and physical PDF page index;
- named study sessions with their cursor and current section;
- progress logs;
- flashcard decks, cards, and flashcard revision data.

The filesystem is used only for temporary converter input/output and explicit exports. The Python converter writes temporary preparation artifacts; after validation, Go ingests the complete prepared book into SQLite transactionally. A failed import must not expose a partial book.

Human inspectability does not require TOML to be canonical. The database can be inspected with ordinary SQLite viewers, and deterministic exports provide portable flashcard files. A later hosted pipeline may place rendered artifacts in S3 or another appropriate store; that future scale target must not complicate the local proof of concept.

### 4. Study agents will not receive PDFs as source material

**Decided.**

A prepared book exposes:

- compact ordered table-of-contents entries;
- rendered page images.

Each TOC entry contains its order, nesting depth, title, and physical PDF page index. No repeated page-label prose, Markdown formatting, or image-path mapping is required.

The agent retrieves page images progressively. Rendered page images are the source of truth for page content, layout, diagrams, tables, captions, and semantic boundaries.

Prepared split PDFs are removed from the new design. They consumed substantial storage and encouraged the model to privilege extracted PDF text over the actual page image. The original PDF is converter input only and is not retained in SQLite or copied into canonical service storage.

The service never deletes or modifies the caller's source PDF. It removes only temporary files that it created for conversion. Re-preparing a book therefore requires the caller to supply the source PDF again.

### 5. The API has three domains

**Decided.**

The MCP surface is divided conceptually into:

1. book management;
2. progressive reading;
3. flashcard storage and editing.

These are domains within the Go MCP server, not three separate services.

### 6. There is no globally selected book

**Decided.**

`select_book(book_id)` must not exist. A global selected-book state would allow study sessions for different books to interfere with one another.

Every book-scoped operation takes a `book_id`. Each book owns its page collection, flashcard collection, and zero or more named study sessions; each session has its own cursor and progress log.

A title is editable display metadata. A stable `book_id` identifies the book in storage and does not change when the title changes.

### 7. The service does not judge whether learning is complete

**Decided.**

`complete_page` must not exist. It would combine navigation, pedagogical judgment, progress semantics, and possibly flashcard creation into an operation the storage API cannot honestly perform.

The LLM decides:

- what question to ask;
- whether an answer demonstrates understanding;
- whether terminology needs explanation;
- whether the learner should continue;
- whether knowledge deserves a flashcard;
- what should be logged about the session.

The service records page position and explicit log messages. It does not claim that a page, section, or concept has been understood merely because the cursor moved.

### 8. Flashcard decks are canonical, mutable local records

**Decided.**

The new system replaces the old `card-inbox.md -> archive -> Quizlet` pipeline.

There is no required inbox, publication transition, archived batch, or `archive_batch` operation. A flashcard deck becomes canonical as soon as it is created. The agent can progressively add, revise, and remove cards as the learner works through a book.

Quizlet is no longer the canonical store. Export to Quizlet-compatible text or another review system can be added later as a deterministic derivative of local decks.

Confirmed version-one card fields are:

- `id`;
- `deck_id`;
- `front`;
- `back`;
- `source_pages`, stored as CSV physical PDF page indices;
- integer `revision`.

Each logical card ID can have multiple immutable card records distinguished by revision number. Creating a card creates revision 1. Updating a card checks `expected_revision` and, when it matches the latest revision, inserts the next revision with the same card ID. Existing revisions are never rewritten. The latest revision is the canonical current card, while all earlier revisions remain readable historical data.

Cards have no position or semantic ordering field. When stable output ordering is needed for listing or export, the service sorts by `card_id`; that ordering is not user-managed card state.

Cards do not have tags. Card type, difficulty, explanation, hints, creation time, review statistics, or similar enrichment fields are explicitly prohibited from this system and must not be added later without reversing this decision with the user.

### 9. The MCP service will live in a new project

**Decided.**

The existing `guided-study` repository is the historical ChatGPT prompt/plugin project. It should not become a mixed application repository containing the new Go service, SQLite database code, and Python conversion machinery.

The new implementation will live at `C:\Users\andgate\Projects\guided-study-desktop-mcp`. That project will own the Go MCP server, Fyne Windows application, SQLite schema and storage code, copied-and-adapted Python converter, and tests. The existing repository remains available as the source of the old workflow and the pedagogical behavior that will later be adapted to use the new service.

The descriptive `guided-study-desktop-mcp` name is intentionally scoped to the local Windows proof of concept. A later cloud implementation can fork or otherwise derive from it under the reserved `guided-study-service` project name. Until implementation is explicitly requested, the new project will not be created.

### 10. The Windows tray and MCP server are one application

**Decided.**

The local Windows 11 build is one long-running Go application and one executable. That process owns:

- the system-tray icon and menu;
- the streamable HTTP MCP server listening at its configured address and port;
- SQLite access;
- Python converter subprocesses;
- any later settings or management window.

A custom ChatGPT plugin stores the MCP server URL and connection details. ChatGPT connects to that endpoint and does not launch, own, or supervise the application.

The application starts the MCP server alongside the desktop event loop. Closing a settings window does not exit the application; choosing Quit from the tray shuts down the MCP endpoint and database cleanly. A separate tray executable, Windows Service, or STDIO bridge is not part of the planned local architecture.

The application will use Fyne v2 for its Windows desktop lifecycle, system-tray menu, and any initial settings/status window. Fyne's built-in tray API is stable, well documented, and keeps the application in Go without adding a web-frontend toolchain. Wails v2 would require separate tray integration, while Wails v3 is still new enough that AI-generated code may confuse its APIs with Wails v2.

## Local storage model

The local library is one SQLite database. Its exact schema remains to be designed, but the ownership boundaries are fixed:

- Go owns all canonical database writes and revision checks.
- Python writes only temporary conversion output.
- TOC entries are stored in a structured SQLite table and returned by `get_book` as a compact ordered array.
- Each rendered page is stored as an image BLOB with its MIME type and physical PDF page index.
- Flashcard history remains queryable even after later edits.
- Explicit exports are derivatives and never become the canonical source.

Large artifacts can move to S3 or another suitable store when a hosted architecture is actually designed. Local storage does not need to imitate that architecture now.

The version-one TOC table has the logical fields `book_id`, `position`, `depth`, `title`, and `page_index`. `page_index` is always the physical PDF page index. The Python converter may emit these entries as temporary CSV; Go validates and ingests them, and CSV is not canonical storage.

## MCP tool surface

The names below describe intended responsibilities. They are not yet a frozen protocol.

### Book management

```text
import_book(file_reference, title?)
list_books()
get_book(book_id)
rename_book(book_id, new_title)
remove_book(book_id)
```

Important constraints:

- `import_book` invokes the Python converter; it does not contain another PDF conversion implementation in Go.
- `list_books` should provide enough summary metadata to let an agent identify a book.
- `get_book(book_id)` returns book metadata and the ordered structured TOC entries. It must not return page images, extracted book text, or other page content.
- Renaming a book changes display metadata, not its stable ID.
- `remove_book` performs immediate hard deletion after exact target validation. It has no service-level trash or confirmation-token protocol and must be annotated as destructive.

The exact method by which a ChatGPT upload becomes `file_reference` is open and must be tested rather than assumed.

### Progressive reading

```text
create_session(book_id, name)
list_sessions(book_id)
rename_session(book_id, session_id, new_name)
delete_session(book_id, session_id)
current_page(book_id, session_id)
next_page(book_id, session_id, expected_page_index)
prev_page(book_id, session_id, expected_page_index)
goto_page(book_id, session_id, target_page_index, expected_page_index)
get_progress(book_id, session_id)
set_current_section(book_id, session_id, section)
log_progress(book_id, session_id, expected_last_log_id, log_message)
amend_last_log(book_id, session_id, expected_last_log_id, replacement_message)
delete_last_log(book_id, session_id, expected_last_log_id)
```

Likely navigation behavior:

- `current_page` returns the current rendered page plus its identity and useful TOC context without moving the cursor.
- `next_page` advances that book's cursor by one page and returns the new rendered page and metadata.
- `prev_page` moves backward by one page and returns the resulting page and metadata.
- `goto_page` moves to an explicitly numbered physical PDF page index and returns it.
- navigation validates bounds and never silently wraps between the first and last pages;
- `set_current_section` stores agent-supplied text. The service and Python converter do not attempt to infer the learner's current semantic section from the PDF or TOC;
- `log_progress` appends the supplied message only when the current tail ID matches `expected_last_log_id`. The service attaches the current page, current section, and timestamp context;
- `amend_last_log` and `delete_last_log` operate only on the current tail and reject an incorrect `expected_last_log_id`;
- a log-tail conflict returns the expected and actual tail IDs plus enough recent entries for the agent to recognize unexpected newer state;
- `get_progress` returns the current physical PDF page index, current section, and recent progress-log entries. It does not return a pending question, generated narrative summary, concepts-learned model, cursor revision, or cursor-updated timestamp.

Study sessions are durable, named reading threads scoped to one book. They are not tied to an MCP connection, ChatGPT conversation, or individual agent process. A new chat lists the sessions for the chosen book, resumes an unambiguous session, or asks the user which session to use. If none is appropriate, it creates a new named session.

`list_sessions` returns each session's stable ID, editable name, current page, current section, and latest log entry so the agent has enough context to choose. Session names are unique within a book under case-insensitive comparison. Renaming a session changes display metadata only. Deleting a session immediately hard-deletes its cursor and progress log but does not affect the book or its flashcards.

The important interaction style is command-oriented: `next_page` and `prev_page` control progress themselves. The caller does not calculate and submit a new page number for ordinary sequential reading.

`get_book` owns metadata and TOC access; the reading operations own rendered-page retrieval and session cursor state. There must not be a global selected-book state hiding behind these operations.

The current section is deliberately free text supplied by the teaching agent. The converter is not expected to extract reliable structured section identities, and nominal TOC boundaries may not match semantic boundaries. Repeating the same section value is an idempotent update.

The progress log is a plain chronological stream without categories or a regenerated narrative summary. Each committed entry has a monotonically increasing ID that is never reused. Only tail correction is supported in version one; removing several unwanted entries requires guarded last-entry deletions one at a time.

An amendment updates the current tail entry in place and keeps its existing log ID, page, section, and timestamp context. Log IDs identify entries, not revisions. After conversation branching removes newer tool calls from the agent's context, its next mutation carries the older tail ID it still expects. The resulting conflict exposes the actual newer tail and recent entries. The agent can then leave an already-correct entry alone, amend the actual tail in place, or delete unwanted tail entries from newest to oldest before appending corrected state.

`create_session()` returns a generated UUID. The agent explicitly supplies that `session_id` to every cursor and progress operation. Any later agent can continue the same durable study thread by selecting that session. The service does not store or manage transient agent identities.

### Flashcard storage

The API must manage decks and their card collections, not merely create anonymous batches.

Deck-level operations:

```text
list_decks(book_id)
create_deck(book_id, deck_id, title, description?, cards?)
read_deck(book_id, deck_id)
update_deck(book_id, deck_id, metadata_changes)
delete_deck(book_id, deck_id)
```

Individual card operations:

```text
add_card(book_id, deck_id, card)
update_card(book_id, deck_id, card_id, changes)
delete_card(book_id, deck_id, card_id, expected_revision)
read_card_revision(book_id, deck_id, card_id, revision)
list_card_revisions(book_id, deck_id, card_id)
```

Batch card operations:

```text
add_cards(book_id, deck_id, cards)
delete_cards(book_id, deck_id, card_ids)
```

Cross-deck moves, batch updates, reordering, and whole-deck replacement are outside version one. Routine edits use narrower operations.

Deletion is immediate and permanent in version one:

- `delete_last_log` removes only the guarded current tail entry;
- `delete_card` removes the logical card and all of its immutable revisions;
- `delete_deck` removes the deck and all contained cards and revisions;
- `remove_book` removes the complete prepared book and its study data.

There are no trash tables, soft-delete flags, restoration workflow, or confirmation tokens. Every destructive tool validates its exact target, applies relevant concurrency checks, and is accurately annotated as destructive. Revision immutability means historical card records are never edited; explicit deletion of the entire logical card still removes them.

The protocol and user-facing model use "deck" consistently. `deck_id` identifies the deck associated with every card revision.

Potential deterministic exports, after canonical storage works:

```text
export_deck(book_id, deck_id, format)
```

Possible formats include TSV, CSV, Anki-compatible output, and Quizlet import text. Export never archives, locks, or mutates the canonical deck.

## State and concurrency rules

### Different books

**Decided requirement:** sessions belonging to different books must not share or overwrite cursor state. Book-scoped rows and mandatory `book_id` parameters provide this isolation.

### Transactions and concurrency

Go performs each bounded mutation in a SQLite transaction. Navigation calls supply `expected_page_index`; the service moves the cursor only when the stored page matches that expectation. This prevents a regenerated tool call from silently advancing an already-moved cursor again without adding a meaningless revision to pedagogical progress.

Flashcard mutations continue to use `expected_revision`. A successful update inserts a new immutable card revision; it does not mutate or replace the expected revision.

Cross-deck card moves are outside version one.

### Durable sessions and multiple agents

**Decided:** a cursor and progress log belong to a named, book-scoped study session rather than to an agent execution. Reading and progress operations require both `book_id` and `session_id`. A user may create separate sessions for chapters or other study threads, and multiple successive agents can resume the same session.

Different sessions for the same book have independent cursor and log state. Multiple agents deliberately operating on one session share that state. Cursor mutations require `expected_page_index`, and log mutations require the expected tail ID, so stale or replayed calls fail instead of silently advancing or rewriting newer state.

The service stores no transient agent identity, connection ownership, session expiry, user permissions, or automatic session selection. At the start of a chat, the teaching agent lists sessions and either selects an obvious match or asks the user which one to resume.

## Book import and preparation flow

### Local target flow

```text
user supplies PDF to the LLM host
        ↓
agent calls import_book with an accessible file reference
        ↓
Go validates the request and chooses the book identity/destination
        ↓
Go runs the Python converter as a subprocess
        ↓
Python renders page images and produces a temporary TOC CSV
        ↓
Go validates and transactionally ingests the prepared book into SQLite
```

The converter must be updated so its temporary output contains no split PDFs. It should render each source page once and emit the compact TOC fields in CSV form for Go to validate and ingest.

Conversion failure must not leave a partially importable book masquerading as ready. Python writes to a temporary import directory; Go validates the result, ingests the complete book in one SQLite transaction, and then removes the temporary artifacts. The exact Go/Python handoff contract should be specified before coding.

### Upload/file-reference uncertainty

**Open and important.** We have not established that ChatGPT automatically forwards uploaded file bytes, a local path, or an opaque file handle to arbitrary MCP tools.

ChatGPT is the sole first-class LLM host for version one. Codex may be used as a development client, but its attachment or MCP behavior must not define or stand in for the ChatGPT integration contract. Claude and other hosts are out of scope.

The earliest integration spike should answer:

1. What argument can an MCP tool actually receive after a user uploads a PDF in ChatGPT?
2. Is that reference readable by a local MCP process?
3. What changes when the MCP server is remote?
4. Can the host transfer the file without putting PDF content into the model's study context?

If a host cannot pass uploads to MCP, a direct upload endpoint or minimal upload interface may eventually be necessary. That is a fallback to test, not an excuse to invent a Python server or a second conversion service now.

## Responsibilities by component

| Component | Owns | Does not own |
|---|---|---|
| Study agent/LLM | Teaching strategy, questions, evaluation, deciding what matters, card wording, calling tools | Durable storage implementation, PDF conversion, pretending it can infer completion from an API call |
| Go MCP server | Tool contracts, book lifecycle, per-book navigation, progress/log persistence, flashcard mutations, invoking conversion | PDF rendering logic, pedagogical judgment, global selected-book state |
| Python converter | Reading one input PDF and writing temporary rendered images plus TOC CSV to the requested destination | Canonical storage, MCP transport, book CRUD, progress, flashcards, a web server |
| Storage layer | Durable prepared artifacts and mutable study state | Teaching or conversion logic |

## Lessons from the prototype that this architecture must preserve

Not every observed issue requires its own feature. The replacement architecture should address the important structural failures:

- Page images, not PDF extraction or nominal section ranges, are the authority for what appears on a page.
- A semantic section may continue across a nominal TOC boundary or share a page with the next section.
- File-store search by duplicated names is not reliable navigation; MCP operations should resolve records deterministically from `book_id` and page identity.
- Persistence operations must be bounded. A save or log call performs one well-defined mutation and returns; it must not enter a recursive search/verification loop.
- Conversation state and durable state can diverge, so the agent should make intentional navigation and logging calls as it works.
- Conceptual understanding is more important than matching textbook wording.
- The user may read ahead or change scope; the teaching prompt must accommodate that rather than blindly replaying a linear script.
- External generative Quizlet creation cannot be treated as deterministic publication of canonical cards.

## Explicitly rejected designs

The following should not be reintroduced without revisiting the decision with the user:

- Google Drive as the durable backend for the serious implementation;
- a Python web service or FastAPI service;
- a Go layer that duplicates the Python converter's PDF logic;
- giving PDFs or split PDFs to the study agent;
- a global `select_book` operation;
- `complete_page`;
- `archive_batch`;
- JSON or TOML files as the canonical mutable flashcard store;
- cross-deck card moves in version one;
- batch card updates, explicit reordering, or whole-deck replacement in version one;
- an inbox/archive/publish lifecycle as the canonical flashcard model;
- treating Quizlet generation as an exact storage API.

## Local-first implementation sequence

### Phase 0: freeze contracts before implementation

- Review and correct this plan.
- Define the version-one SQLite schema.
- Decide stable ID rules for books, study sessions, decks, and cards.
- Define revision checks and the minimum flashcard history needed to inspect or restore earlier content.
- Specify the Python converter's input/output and failure contract.
- Define MCP result shapes, especially how a page image and page metadata are returned together.

### Phase 1: narrow technical spikes

- Create `C:\Users\andgate\Projects\guided-study-desktop-mcp`.
- Create the smallest possible Go MCP server with streamable HTTP on a configured local address and port.
- Register that server URL in ChatGPT developer mode and verify the connection.
- Return one local image through an MCP tool and confirm the target agent host can consume it correctly.
- Test uploaded-PDF handoff from ChatGPT to the MCP tool rather than assuming support.
- Run the existing Python converter from Go and capture exit status and diagnostics.
- Test optimistic concurrency and flashcard revision history in the chosen storage backend.

These spikes should be disposable or directly reusable. They exist to resolve protocol uncertainty before the full service is built.

### Phase 2: SQLite model and book preparation

- Implement the SQLite schema and stable IDs.
- Modify the Python converter to produce temporary images plus compact TOC CSV without split PDFs.
- Implement `import_book`, `list_books`, `rename_book`, and exact-target hard deletion through `remove_book`.
- Ingest prepared books transactionally so failed conversion or ingestion does not expose an incomplete book.

### Phase 3: progressive reader

- Implement named book-scoped sessions with independently resumable cursors and logs.
- Implement current/next/previous/goto operations that return the rendered image and page context.
- Implement progress inspection and bounded progress logging.
- Test bounds, missing pages, corrupt metadata, restart/resume behavior, separate sessions on one book, and multiple agents resuming the same session.

### Phase 4: flashcard collection

- Implement canonical flashcard storage in SQLite.
- Implement deck lifecycle operations.
- Implement individual card operations plus the retained narrow batch operations.
- Verify revision conflicts do not silently overwrite newer card content.

### Phase 5: rewrite the study behavior

- Replace Google Drive traversal with the MCP tools.
- Remove PDF-first and split-PDF instructions.
- Remove the card inbox, archive transaction, and Quizlet-as-canonical workflow.
- Preserve the useful pedagogical rules from the existing study prompt.
- Add clear bounded-tool-call behavior and recovery rules.
- Update `README.md`, `ARCHITECTURE.md`, the skills, and tests only after the MCP contract is stable enough to document accurately.

### Phase 6: deterministic export

- Add exports only after local card editing is trustworthy.
- Verify exported terms are exact derivatives of the canonical database records.
- Keep export history, if desired later, separate from canonical card content.

### Phase 7: hosted service

- Deploy the existing streamable HTTP MCP transport behind an appropriate hosted endpoint.
- Choose hosted persistence based on the needs demonstrated by the local proof of concept.
- Add authentication, tenant isolation, quotas, conflict handling, backup/versioning, and conversion resource limits.
- Package the Go server and Python conversion dependency together.
- Revisit upload flow using evidence from the earlier host integration spike.

## Questions still requiring discussion

All questions from the original discussion list have been resolved. Additional contract details should be introduced only when they are necessary to specify the version-one storage and MCP contract.

## Immediate next step

Review this document for incorrect assumptions and missing decisions. Once corrected, the next useful artifact is the version-one storage and MCP contract: the SQLite schema, temporary/export directory rules, tool arguments, tool results, validation rules, and error behavior. Implementation should follow that contract rather than continuing to grow the obsolete Drive/Quizlet workflow.
