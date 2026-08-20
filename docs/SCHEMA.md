# Guided Study SQLite Schema

## Status

This document is the version-one storage contract for the local proof of concept. It refines `PLAN.md`; it does not describe an implemented database yet.

SQLite is the canonical store. The source PDF and converter output are temporary inputs only. Explicit exports are derivatives and are not read back as canonical state.

## Conventions

- The application enables `PRAGMA foreign_keys = ON` for every database connection.
- All writes that affect more than one row are performed in a transaction.
- `page_index` is always the **1-based physical PDF page index**. Page 1 is the first physical page in the source PDF, regardless of a printed page label.
- `book_id`, `session_id`, and `card_id` are service-generated UUIDs in canonical lowercase text form.
- `deck_id` is supplied by the caller and must match `[a-z0-9][a-z0-9_-]{0,63}`. It is unique within a book.
- IDs are stable and are never derived from editable titles.
- Blank titles, card fronts, card backs, sections, and log messages are rejected after trimming, except that an unset current section is represented by `NULL`.
- No creation or update timestamps are stored for books, cursors, decks, or cards. A progress-log timestamp is stored because it is part of the entry context established in `PLAN.md`.
- The schema intentionally has no tags, card order, card type, difficulty, explanation, hints, review statistics, or similar enrichment fields.

## Initial DDL

```sql
PRAGMA foreign_keys = ON;

CREATE TABLE books (
    book_id    TEXT PRIMARY KEY,
    title      TEXT NOT NULL CHECK (length(trim(title)) > 0),
    page_count INTEGER NOT NULL CHECK (page_count >= 1)
);

CREATE TABLE book_pages (
    book_id    TEXT NOT NULL,
    page_index INTEGER NOT NULL CHECK (page_index >= 1),
    mime_type  TEXT NOT NULL CHECK (length(trim(mime_type)) > 0),
    image_data BLOB NOT NULL CHECK (length(image_data) > 0),
    PRIMARY KEY (book_id, page_index),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE toc_entries (
    book_id    TEXT NOT NULL,
    position   INTEGER NOT NULL CHECK (position >= 1),
    depth      INTEGER NOT NULL CHECK (depth >= 0),
    title      TEXT NOT NULL CHECK (length(trim(title)) > 0),
    page_index INTEGER NOT NULL CHECK (page_index >= 1),
    PRIMARY KEY (book_id, position),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, page_index)
        REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE INDEX toc_entries_by_page
    ON toc_entries (book_id, page_index, position);

CREATE TABLE study_sessions (
    book_id         TEXT NOT NULL,
    session_id      TEXT NOT NULL,
    name            TEXT NOT NULL COLLATE NOCASE
        CHECK (length(trim(name)) > 0),
    page_index      INTEGER NOT NULL DEFAULT 1 CHECK (page_index >= 1),
    current_section TEXT NULL
        CHECK (current_section IS NULL OR length(trim(current_section)) > 0),
    PRIMARY KEY (book_id, session_id),
    UNIQUE (book_id, name),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, page_index)
        REFERENCES book_pages (book_id, page_index)
) WITHOUT ROWID;

CREATE TABLE progress_logs (
    log_id      INTEGER PRIMARY KEY AUTOINCREMENT,
    book_id     TEXT NOT NULL,
    session_id  TEXT NOT NULL,
    page_index  INTEGER NOT NULL CHECK (page_index >= 1),
    section     TEXT NULL,
    message     TEXT NOT NULL CHECK (length(trim(message)) > 0),
    logged_at   TEXT NOT NULL CHECK (length(trim(logged_at)) > 0),
    FOREIGN KEY (book_id, session_id)
        REFERENCES study_sessions (book_id, session_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, page_index)
        REFERENCES book_pages (book_id, page_index)
);

CREATE INDEX progress_logs_by_stream
    ON progress_logs (book_id, session_id, log_id);

CREATE TABLE decks (
    book_id     TEXT NOT NULL,
    deck_id     TEXT NOT NULL,
    title       TEXT NOT NULL CHECK (length(trim(title)) > 0),
    description TEXT NULL,
    revision    INTEGER NOT NULL DEFAULT 1 CHECK (revision >= 1),
    PRIMARY KEY (book_id, deck_id),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE card_revisions (
    book_id      TEXT NOT NULL,
    deck_id      TEXT NOT NULL,
    card_id      TEXT NOT NULL,
    revision     INTEGER NOT NULL CHECK (revision >= 1),
    front        TEXT NOT NULL CHECK (length(trim(front)) > 0),
    back         TEXT NOT NULL CHECK (length(trim(back)) > 0),
    source_pages TEXT NOT NULL CHECK (length(trim(source_pages)) > 0),
    PRIMARY KEY (book_id, deck_id, card_id, revision),
    FOREIGN KEY (book_id, deck_id)
        REFERENCES decks (book_id, deck_id) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE INDEX card_revisions_latest
    ON card_revisions (book_id, deck_id, card_id, revision DESC);

CREATE VIEW current_cards AS
SELECT candidate.book_id,
       candidate.deck_id,
       candidate.card_id,
       candidate.revision,
       candidate.front,
       candidate.back,
       candidate.source_pages
FROM card_revisions AS candidate
WHERE NOT EXISTS (
    SELECT 1
    FROM card_revisions AS newer
    WHERE newer.book_id = candidate.book_id
      AND newer.deck_id = candidate.deck_id
      AND newer.card_id = candidate.card_id
      AND newer.revision > candidate.revision
);
```

## Table behavior and invariants

### Books, pages, and TOC

A successful import inserts the book, every rendered page, and every TOC entry in one transaction. Before committing, Go validates that:

- pages are numbered contiguously from 1 through `books.page_count`;
- every image MIME type is supported by the MCP page-result implementation;
- every TOC `position` is contiguous from 1;
- every TOC `page_index` names an ingested page;
- depth values are non-negative.

An empty TOC is valid. `get_book` returns the book metadata and TOC entries ordered by `position`. It never returns `image_data`.

`remove_book` deletes the book row. Foreign-key cascades hard-delete its pages, TOC, cursor state, progress logs, decks, and card revisions. It does not delete the caller's source PDF.

### Study sessions and reading progress

A study session is a durable, named reading thread for one book. It is not an LLM process, chat connection, user account, or permission boundary. Any later agent or chat can resume it by using its `session_id`.

Creating a session generates its UUID and stores page 1 with no current section. Session names are trimmed and unique within a book under case-insensitive comparison. They are editable display metadata; renaming a session never changes its ID.

`list_sessions` exposes each session's name, current page, current section, and latest progress-log entry so an agent can recognize which thread to resume. Deleting a session hard-deletes only that session's cursor and progress log. It does not delete the book, decks, or cards.

Navigation compares the caller's `expected_page_index` to the stored session page before writing. A mismatch changes nothing. Different sessions for the same book have independent cursors and logs; multiple agents deliberately using the same session share its state and are protected by the existing cursor and log-tail checks.

`current_section` is agent-supplied free text. The service does not infer it from the TOC. Setting the same normalized value again is idempotent. Passing JSON `null` clears the section.

### Progress log

`log_id` is globally monotonic because the table uses `AUTOINCREMENT`; a deleted ID is never reused. Ordering within a session stream is by `log_id`.

An append captures the cursor's current `page_index`, current section, and a service-generated UTC RFC 3339 timestamp. It succeeds only if `expected_last_log_id` equals the stream's actual tail ID, where both are `null` when the stream is empty.

An amendment is allowed only for the guarded tail entry. It changes only `message` and preserves the entry's `log_id`, page, section, and timestamp. A deletion is also tail-only and removes exactly one entry. Correcting several entries therefore requires repeated guarded deletions from newest to oldest.

### Decks and cards

`decks.revision` is a lightweight optimistic-concurrency token for deck metadata only. Creating a deck sets it to 1. A successful `update_deck` requires the latest revision and increments it by one. Adding, updating, or deleting cards does not change the deck metadata revision.

Each row in `card_revisions` is an immutable version of one logical card. Creation inserts revision 1. Update compares `expected_revision` with the latest revision, then inserts revision `N + 1` while copying any unchanged fields. Existing rows are never updated.

`source_pages` is stored in canonical CSV form, for example `"3,4,9"`: decimal 1-based page indices, ascending, unique, and without spaces. Go validates every index against the owning book. It is deliberately not expanded into tags, sources, or explanatory metadata.

Current cards are obtained through `current_cards`. Listing and export sort by `card_id`; the database stores no user-managed position.

`delete_card` hard-deletes every revision with the exact `(book_id, deck_id, card_id)` after checking the latest revision. `delete_deck` cascades through all card revisions in that deck.

## Python converter handoff

The converter receives an input PDF path or readable reference resolved by Go and a newly created temporary output directory. On success it writes:

- one image per physical page, named `page-0001.<ext>`, `page-0002.<ext>`, and so on without gaps;
- `toc.csv` with the exact header `position,depth,title,page_index`.

The converter exits nonzero on failure and writes diagnostics to standard error. It must not write canonical SQLite state, retain the source PDF, or emit split PDFs.

Go treats converter output as untrusted staging data. It validates file names, image formats, page continuity, CSV shape, TOC references, and configured size limits before beginning the ingest transaction. After success or failure, Go removes only the temporary directory it created. The original PDF is never copied into the database and never deleted or modified.

## Deliberately absent

Version one has no soft-delete columns, trash tables, agent identity table, chat/connection tracking, card revision timestamps, cursor revision, cursor timestamp, pending question, concepts-learned field, card ordering, tags, review scheduling, or migration framework. Those omissions are intentional, not unfinished columns to add opportunistically.
