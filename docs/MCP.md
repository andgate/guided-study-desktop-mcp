# Guided Study MCP Contract

## Status and scope

This document is the draft version-one MCP contract for `guided-study-desktop-mcp`. It refines the provisional tool list in `PLAN.md`; no implementation is implied.

The local Windows application exposes one Streamable HTTP MCP endpoint, initially `/mcp`, from the same Go process that owns the tray application and SQLite connection. ChatGPT connects to the configured address and port. STDIO and non-ChatGPT hosts are outside the version-one contract.

The host supplies `import_book.file_reference` as an absolute path to the local PDF.

## Result and error conventions

- Ordinary successful tools return JSON in `structuredContent` and a short text fallback in `content`.
- Page-reading tools return structured metadata plus rendered MCP image blocks. Image bytes are not duplicated inside `structuredContent`.
- Inputs and structured results receive JSON Schemas in the server registration.
- Missing optional fields mean “not supplied.” JSON `null` is used only where this contract explicitly permits it.
- Mutations are atomic. A validation or conflict error leaves state unchanged.
- Destructive tools are registered with MCP's destructive annotation. Read-only tools are marked read-only.
- Times are UTC RFC 3339 strings.
- Page indices are 1-based and assigned during import.

An unsuccessful tool call sets MCP `isError` and returns this structured error shape with a matching short text message:

```json
{
  "code": "cursor_conflict",
  "message": "Expected page 12, but this cursor is on page 13.",
  "details": {
    "book_id": "…",
    "session_id": "…",
    "expected_page_index": 12,
    "actual_page_index": 13
  }
}
```

Stable version-one error codes are:

| Code | Meaning |
|---|---|
| `invalid_argument` | An input is missing, malformed, blank, or internally inconsistent. |
| `not_found` | The exact book, session, deck, card, revision, or page does not exist. |
| `already_exists` | A caller-supplied deck ID or case-insensitive session name is already in use within its book. |
| `out_of_bounds` | A requested page falls below 1 or above the book's page count. |
| `cursor_conflict` | `expected_page_index` does not equal the effective stored cursor. |
| `log_tail_conflict` | `expected_last_log_id` does not equal the current log tail. |
| `deck_revision_conflict` | `expected_revision` does not equal the deck metadata revision. |
| `card_revision_conflict` | `expected_revision` does not equal the logical card's latest revision. |
| `conversion_failed` | The converter could not run, returned failure, or produced invalid staging output. |
| `storage_error` | SQLite or local storage failed unexpectedly. |

Conflict errors include expected and actual values. `log_tail_conflict` additionally includes up to the latest 10 entries in ascending order so an agent can recognize and repair state created by a discarded conversation branch.

## Shared result types

```text
BookSummary {
  book_id: string
  title: string
  page_count: integer
}

TocEntry {
  position: integer
  depth: integer
  title: string
  page_index: integer
}

Book {
  book_id: string
  title: string
  page_count: integer
  toc: TocEntry[]
}

PageMetadata {
  book_id: string
  session_id: string
  session_name: string
  page_index: integer
  page_count: integer
  current_section: string | null
  toc_context: TocEntry[]
}

PageBatchEntry {
  page_index: integer
  toc_context: TocEntry[]
}

PageBatch {
  book_id: string
  page_count: integer
  start_page_index: integer
  end_page_index: integer
  pages: PageBatchEntry[]
}

ProgressLogEntry {
  log_id: integer
  page_index: integer
  section: string | null
  message: string
  logged_at: string
}

SessionSummary {
  book_id: string
  session_id: string
  name: string
  page_index: integer
  current_section: string | null
  last_log: ProgressLogEntry | null
}

Progress {
  book_id: string
  session_id: string
  session_name: string
  page_index: integer
  current_section: string | null
  last_log_id: integer | null
  recent_logs: ProgressLogEntry[]
}

DeckSummary {
  book_id: string
  deck_id: string
  title: string
  description: string | null
  revision: integer
  card_count: integer
}

CardRevision {
  card_id: string
  deck_id: string
  front: string
  back: string
  source_pages: string
  revision: integer
}

Deck {
  book_id: string
  deck_id: string
  title: string
  description: string | null
  revision: integer
  cards: CardRevision[]
}
```

`toc_context` contains the closest TOC lineage available for the returned page: the latest preceding entry at each applicable depth, followed by entries beginning on that exact page. It is navigation context, not inferred semantic section data.

`source_pages` is canonical CSV such as `"3,4,9"`, not a JSON array. Values are ascending, unique, 1-based, and contain no spaces.

## Book management tools

### `import_book`

Input:

```text
file_reference: string
title: string
```

The caller supplies the book's absolute local path and nonblank display title. The server passes the path to the PDF converter, validates all staged pages and TOC rows, and commits the entire book in one SQLite transaction. The service generates `book_id` and does not derive or replace the supplied title.

Result: `BookSummary`.

This tool does not retain, modify, or delete the caller's source PDF. It is non-destructive but not read-only.

### `list_books`

Input: none.

Result:

```text
{ books: BookSummary[] }
```

Books are sorted case-insensitively by title and then by `book_id`. Read-only.

### `get_book`

Input: `book_id: string`.

Result: `Book`, with TOC entries ordered by `position`.

This tool never returns page images, source PDF data, or extracted page text. Read-only.

### `rename_book`

Input:

```text
book_id: string
new_title: string
```

Result: updated `BookSummary`. The stable ID is unchanged. Non-destructive mutation.

### `remove_book`

Input: `book_id: string`.

Result:

```text
{ book_id: string, deleted: true }
```

This immediately hard-deletes the exact book and all canonical study data owned by it. It never touches the source PDF. Destructive.

## Page batch tools

### `read_pages`

Input:

```text
book_id: string
start_page_index: integer
end_page_index: integer
```

The range is inclusive and may have any valid size. The start must not exceed
the end, and both indices must fall within the book.

Result: `PageBatch` in `structuredContent`. Each page entry has matching TOC
context. The MCP content contains `Page N.` before each rendered image, with
entries, labels, and images in ascending order.

This tool is read-only. It does not require or access a study session, and it
does not move any session cursor.

## Progressive reading tools

A study session is a durable, named reading thread for exactly one book. It is not tied to the MCP connection or to one ChatGPT conversation. Any later agent can resume it by supplying its stable `session_id`.

There is no implicit or globally active session. After choosing a book, a new chat calls `list_sessions`, resumes an obvious session, asks the user when the choice is ambiguous, or creates a new one. Every reading and progress tool requires an existing session belonging to the supplied book.

Session page tools produce `PageMetadata` in `structuredContent` and one rendered MCP image block.

### `create_session`

Input:

```text
book_id: string
name: string
```

The server trims the name, requires it to be unique within the book under case-insensitive comparison, generates `session_id`, and initializes the cursor to page 1 with no current section or log entries.

Result: `SessionSummary`.

### `list_sessions`

Input: `book_id: string`.

Result: `{ sessions: SessionSummary[] }`, sorted case-insensitively by name and then by `session_id`. `last_log` contains the complete latest entry or `null`, giving the agent enough context to recognize a session. Read-only.

### `rename_session`

Input:

```text
book_id: string
session_id: string
new_name: string
```

The new name must remain unique within the book. The stable ID, cursor, section, and log are unchanged.

Result: updated `SessionSummary`.

### `delete_session`

Input:

```text
book_id: string
session_id: string
```

This immediately hard-deletes the exact session, its cursor, and its progress log. It does not delete the book, decks, or cards.

Result: `{ book_id: string, session_id: string, deleted: true }`. Destructive.

### `current_page`

Input:

```text
book_id: string
session_id: string
```

Result: current page image plus `PageMetadata`. Does not move the cursor. Read-only.

### `next_page`

Input:

```text
book_id: string
session_id: string
expected_page_index: integer
```

The server checks the stored session page, rejects a mismatch, rejects advancing past `page_count`, moves exactly one page, and returns the new page image plus `PageMetadata`.

### `prev_page`

Input is the same as `next_page`. It rejects moving before page 1, moves exactly one page, and returns the resulting page image plus `PageMetadata`.

### `goto_page`

Input:

```text
book_id: string
session_id: string
target_page_index: integer
expected_page_index: integer
```

The server checks the current cursor and target bounds, then stores the target and returns its image plus `PageMetadata`. Going to the already-current page succeeds without creating a second kind of cursor revision.

### `get_progress`

Input:

```text
book_id: string
session_id: string
```

Result: `Progress`. `recent_logs` contains up to the latest 10 entries in ascending order. This tool does not generate a narrative study summary, pending question, concepts-learned state, cursor revision, or update timestamp. Read-only.

### `set_current_section`

Input:

```text
book_id: string
session_id: string
section: string | null
```

The text is supplied by the agent and is not inferred or validated against the TOC. `null` clears it. Repeating the same value is idempotent.

Result: `Progress`.

### `log_progress`

Input:

```text
book_id: string
session_id: string
expected_last_log_id: integer | null
log_message: string
```

The expected ID is required and must be `null` for an empty stream. On success the server snapshots the current page and section, attaches its UTC timestamp, and appends one entry.

Result:

```text
{
  entry: ProgressLogEntry
  last_log_id: integer
}
```

### `amend_last_log`

Input:

```text
book_id: string
session_id: string
expected_last_log_id: integer
replacement_message: string
```

Only the actual tail may be amended. The entry keeps the same ID, page, section, and timestamp; only its message changes.

Result: `{ entry: ProgressLogEntry }`. Destructive because it replaces stored text.

### `delete_last_log`

Input:

```text
book_id: string
session_id: string
expected_last_log_id: integer
```

Only the actual tail may be deleted. Result:

```text
{
  deleted_log_id: integer
  last_log_id: integer | null
}
```

Deleting several unwanted entries requires guarded calls from newest to oldest. Destructive.

## Deck tools

### `list_decks`

Input: `book_id: string`.

Result: `{ decks: DeckSummary[] }`, sorted by `deck_id`. Read-only.

### `create_deck`

Input:

```text
book_id: string
deck_id: string
title: string
description?: string | null
cards?: NewCard[]

NewCard {
  front: string
  back: string
  source_pages: string
}
```

The caller supplies the validated `deck_id`; the service generates a UUID for each card. The deck revision and every initial card revision start at 1. Creation of the deck and all optional cards is one transaction.

Result: `Deck` with cards sorted by `card_id`.

### `read_deck`

Input:

```text
book_id: string
deck_id: string
```

Result: `Deck` containing only the latest revision of each logical card, sorted by `card_id`. Read-only.

### `update_deck`

Input:

```text
book_id: string
deck_id: string
expected_revision: integer
changes: {
  title?: string
  description?: string | null
}
```

At least one change is required. On a revision match, the server updates only deck metadata and increments the deck revision. `null` clears the description. Card content and history are unaffected.

Result: updated `DeckSummary`.

### `delete_deck`

Input:

```text
book_id: string
deck_id: string
expected_revision: integer
```

The revision check prevents deleting a deck whose metadata changed unexpectedly. A successful call immediately deletes the deck and all card revisions.

Result: `{ book_id: string, deck_id: string, deleted: true }`. Destructive.

## Card tools

### `add_card`

Input:

```text
book_id: string
deck_id: string
card: NewCard
```

The service generates `card_id` and inserts immutable revision 1.

Result: `CardRevision`.

### `update_card`

Input:

```text
book_id: string
deck_id: string
card_id: string
expected_revision: integer
changes: {
  front?: string
  back?: string
  source_pages?: string
}
```

At least one change is required. The server compares `expected_revision` with the current latest revision, then inserts revision `N + 1`, copying unchanged values. Existing revisions remain immutable.

Result: the new `CardRevision`.

### `delete_card`

Input:

```text
book_id: string
deck_id: string
card_id: string
expected_revision: integer
```

After checking the latest revision, the server hard-deletes the logical card and all its revisions.

Result: `{ card_id: string, deleted: true }`. Destructive.

### `read_card_revision`

Input:

```text
book_id: string
deck_id: string
card_id: string
revision: integer
```

Result: the exact immutable `CardRevision`. Read-only.

### `list_card_revisions`

Input:

```text
book_id: string
deck_id: string
card_id: string
```

Result: `{ revisions: CardRevision[] }`, including full content in ascending revision order. Read-only.

### `add_cards`

Input:

```text
book_id: string
deck_id: string
cards: NewCard[]
```

All cards are validated first, assigned service-generated UUIDs, and inserted at revision 1 in one transaction. One invalid card rejects the entire call.

Result: `{ cards: CardRevision[] }`, sorted by `card_id`.

### `delete_cards`

Input:

```text
book_id: string
deck_id: string
cards: {
  card_id: string
  expected_revision: integer
}[]
```

Every target and latest revision is validated before deletion. A missing card or any revision conflict rejects the entire call. On success, all named logical cards and their complete histories are hard-deleted in one transaction.

Result: `{ deleted_card_ids: string[] }`, sorted by `card_id`. Destructive.

## Explicitly outside version one

The MCP surface has no global selected book, implicit active session, `complete_page`, extracted-text reader, split-PDF reader, agent identity or connection tracking, cross-deck move, card reorder, batch card update, whole-deck replacement, soft delete, restore, publication/archive workflow, review scheduling, or export tool. Deterministic export can be added after canonical editing is proven.

The service also does not silently accept unrecognized input fields. Rejecting them helps prevent clients from accidentally introducing prohibited card metadata.
