# Guided Study MCP Contract

## Status and scope

This document defines the MCP contract for `guided-study-desktop-mcp`.

The local desktop application exposes one Streamable HTTP MCP endpoint, initially `/mcp`, from the same Go process that owns the tray application and SQLite connection. ChatGPT connects to the configured address and port. STDIO and non-ChatGPT hosts are outside the version-one contract.

The host supplies `import_book.file_reference` as an absolute path to the local PDF or EPUB.

## Result and error conventions

- Ordinary successful tools return JSON in `structuredContent` and a short text fallback in `content`.
- Page-reading tools return structured metadata plus rendered MCP image blocks. Image bytes are not duplicated inside `structuredContent`.
- Inputs and structured results receive JSON Schemas in the server registration.
- Missing optional fields mean “not supplied.” JSON `null` is used only where this contract explicitly permits it.
- Mutations are atomic. A validation or conflict error leaves state unchanged.
- Destructive tools are registered with MCP's destructive annotation. Read-only tools are marked read-only.
- Page indices are 1-based and assigned during import.

An unsuccessful tool call sets MCP `isError` and returns this structured error shape with a matching short text message:

```json
{
  "code": "no_next_batch",
  "message": "No later page batch exists.",
  "details": {
    "book_id": "…",
    "session_id": "…",
    "page_index": 42
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
| `no_next_batch` | No page batch follows the supplied checkpoint. |
| `deck_revision_conflict` | `expected_revision` does not equal the deck metadata revision. |
| `card_revision_conflict` | `expected_revision` does not equal the logical card's latest revision. |
| `conversion_failed` | The converter could not run or return a committed book. |
| `outline_required` | The book has no extractable outline. |
| `outline_unusable` | The extracted outline cannot be stored unchanged. |
| `storage_error` | SQLite or local storage failed unexpectedly. |

Conflict errors include expected and actual values.

## Shared result types

```text
BookSummary {
  book_id: string
  title: string
  page_count: integer
}

Book {
  book_id: string
  title: string
  page_count: integer
  outline_csv: string
}

PageBatchEntry {
  page_index: integer
}

PageBatch {
  book_id: string
  page_count: integer
  start_page_index: integer
  end_page_index: integer
  pages: PageBatchEntry[]
}

SessionSummary {
  book_id: string
  session_id: string
  name: string
  origin_page_index: integer
  checkpoint_page_index: integer
  checkpoint_heading: string | null
}

PageSelection {
  book_id: string
  session_id: string
  batch: {
    start_page_index: integer
    end_page_index: integer
  }
}

CreatedSession {
  session: SessionSummary
  selection: PageSelection
}

ReadingBatch {
  book_id: string
  session_id: string
  batch: {
    start_page_index: integer
    end_page_index: integer
  }
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

`source_pages` is canonical CSV such as `"3,4,9"`, not a JSON array. Values are ascending, unique, 1-based, and contain no spaces.

## Book management tools

### `import_book`

Input:

```text
file_reference: string
title: string
```

The caller supplies the absolute local book path and nonblank display title. The
converter extracts the outline, renders every page, and inserts the complete
book in one SQLite transaction. A missing outline returns `outline_required`.
An extracted outline that cannot be stored unchanged returns
`outline_unusable`.

Result: `BookSummary`.

This tool does not retain, modify, or delete the caller's source book. It is non-destructive but not read-only.

Reflowable formats such as EPUB are laid out at a fixed page size before rendering. Their page indices are canonical inside this library and do not match a printed edition.

### `list_books`

Input: none.

Result:

```text
{ books: BookSummary[] }
```

Books are sorted case-insensitively by title and then by `book_id`. Read-only.

### `get_book`

Input: `book_id: string`.

Result: `Book`. `outline_csv` begins with:

```csv
outline_index,title,page_index
```

Remaining RFC 4180 rows appear in extraction order. The outline helps callers
locate headings and pages. It does not define page ownership, teaching bounds,
or checkpoint identity.

This tool never returns page images, source book data, or extracted page text. Read-only.

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

This immediately hard-deletes the exact book and all canonical study data owned by it. It never touches the source book. Destructive.

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

Result: `PageBatch` in `structuredContent`. The MCP content contains `Page N.`
before each rendered image, with entries, labels, and images in ascending order.

This tool is read-only. It does not require or access a study session, and it
does not move any session cursor.

## Progressive reading tools

A study session is a durable, named reading thread for exactly one book. Any
later agent can resume it with its stable `session_id`.

There is no globally active session. Except for `create_session`, every
progressive-reading tool requires an existing session belonging to the supplied
book.

### `create_session`

Input:

```text
book_id: string
name: string
start_page_index: integer
```

The server trims the name, requires case-insensitive uniqueness within the book,
and generates `session_id`. The start page must exist in the book. It becomes
the batch origin and physical checkpoint. The checkpoint heading starts as
`null`. The response includes the consecutive batch beginning at the chosen
page and rendered images.

Result: `CreatedSession` plus labeled MCP image blocks.

### `list_sessions`

Input: `book_id: string`.

Result: `{ sessions: SessionSummary[] }`, sorted case-insensitively by name and then by `session_id`. Read-only.

### `rename_session`

Input:

```text
book_id: string
session_id: string
new_name: string
```

The new name must remain unique within the book. Progress is unchanged.

Result: updated `SessionSummary`.

### `delete_session`

Input:

```text
book_id: string
session_id: string
```

This hard-deletes the exact session and its progress. It does not delete the
book, decks, or cards.

Result: `{ book_id: string, session_id: string, deleted: true }`. Destructive.

### `goto_page`

Input:

```text
book_id: string
session_id: string
page_index: integer
```

The page must belong to the supplied book. The call makes it the new batch
origin, resets the physical checkpoint to that page, and clears the checkpoint
heading. It returns the consecutive batch beginning at the chosen page.

Result: `PageSelection` plus labeled MCP image blocks.

### `continue_reading`

Input:

```text
book_id: string
session_id: string
page_index: integer
heading: string
```

The physical page and nonblank heading form the learner checkpoint. The heading
is supplied by the agent from the rendered page; the service does not derive or
validate it against the book outline.

The call finds the fixed batch containing the checkpoint, then returns the
following consecutive batch. Batches contain at most five pages and align to
the session's stored origin.

```text
batch_start = origin + floor((checkpoint - origin) / 5) * 5
batch_end = min(batch_start + 4, page_count)
next_batch_start = batch_start + 5
next_batch_end = min(next_batch_start + 4, page_count)
```

The teaching agent preloads before its first question sourced from the current
batch's midpoint page. The trigger is
`batch_start + floor(batch_size / 2)`, where
`batch_size = batch_end - batch_start + 1`.

The checkpoint is saved atomically with successful page loading. The service
stores no delivered-page window. When no later batch exists, `no_next_batch` is
returned and the supplied checkpoint is not saved.

Result: `ReadingBatch` plus labeled MCP image blocks.

### `save_checkpoint`

Input:

```text
book_id: string
session_id: string
page_index: integer
heading: string
```

The call saves the physical page and nonblank agent-supplied heading without
loading pages. The page must fall between the session origin and the final
physical page.

Result: updated `SessionSummary`.

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
