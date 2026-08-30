# Guided Study SQLite Schema

## Status

This document is the storage contract for the local service.

SQLite is the canonical store. The source PDF is an external import input.
Explicit exports are derivatives and are not read back as canonical state.

## Conventions

- The application enables `PRAGMA foreign_keys = ON` for every database connection.
- All writes that affect more than one row are performed in a transaction.
- `page_index` is always a **1-based index assigned during import**. Page 1 is the first rendered page, regardless of a printed page label.
- `book_id`, `session_id`, and `card_id` are service-generated UUIDs in canonical lowercase text form.
- `deck_id` is supplied by the caller and must match `[a-z0-9][a-z0-9_-]{0,63}`. It is unique within a book.
- IDs are stable and are never derived from editable titles.
- Blank titles, outline titles, checkpoint headings, card fronts, and card backs are rejected after trimming.
- No creation or update timestamps are stored for books, cursors, decks, or cards.
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

CREATE TABLE book_outline (
    book_id      TEXT NOT NULL,
    outline_index INTEGER NOT NULL CHECK (outline_index >= 0),
    title        TEXT NOT NULL CHECK (length(trim(title)) > 0),
    page_index   INTEGER NOT NULL CHECK (page_index >= 1),
    PRIMARY KEY (book_id, outline_index),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, page_index)
        REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE INDEX book_outline_by_page
    ON book_outline (book_id, page_index, outline_index);

CREATE TABLE study_sessions (
    book_id         TEXT NOT NULL,
    session_id      TEXT NOT NULL,
    name            TEXT NOT NULL COLLATE NOCASE
        CHECK (length(trim(name)) > 0),
    origin_page_index INTEGER NOT NULL
        CHECK (origin_page_index >= 1),
    checkpoint_page_index INTEGER NOT NULL
        CHECK (checkpoint_page_index >= 1),
    checkpoint_heading TEXT NULL
        CHECK (
            checkpoint_heading IS NULL
            OR length(trim(checkpoint_heading)) > 0
        ),
    PRIMARY KEY (book_id, session_id),
    UNIQUE (book_id, name),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, origin_page_index)
        REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE,
    FOREIGN KEY (book_id, checkpoint_page_index)
        REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE
) WITHOUT ROWID;

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

### Books, pages, and outlines

A successful import inserts the book, every rendered page, and every extracted
outline entry in one converter-owned SQLite transaction.

`outline_index` is zero-based and preserves PDF extraction order. Each outline
entry retains its extracted title and 1-based physical page. No hierarchy,
parent, depth, end page, or page range is stored.

`get_book` returns book metadata and an RFC 4180 outline CSV ordered by
`outline_index`. It never returns `image_data`.

`remove_book` deletes the book row. Foreign-key cascades hard-delete its pages,
outline entries, session state, decks, and card revisions. It does not delete the
caller's source PDF.

### Study sessions

A study session is a durable, named reading thread for one book. It is not an LLM process, chat connection, user account, or permission boundary. Any later agent or chat can resume it by using its `session_id`.

Creating a session requires a physical start page. It stores that page as both
the batch origin and initial checkpoint, with a null heading. Session names are
trimmed and unique within a book under case-insensitive comparison. Renaming a
session never changes its ID or checkpoint.

`list_sessions` exposes the stored origin, physical checkpoint page, and
nullable agent-supplied heading. Deleting a session hard-deletes only that state.
It does not delete the book, decks, or cards.

`goto_page` stores the selected page as the origin and checkpoint, clears the
heading, and returns the batch starting on that page. `save_checkpoint` updates
the physical page and heading without loading pages. `continue_reading`
validates and saves a supplied checkpoint while returning the deterministic
batch after the one containing that page. No delivered-page window or
completion state is stored.

Batch size is capped at five pages. The batch grid starts at the stored origin
and continues through the final physical page without overlap. A request beyond
the final available batch returns `no_next_batch` without changing the
checkpoint.

```text
batch_start = origin + floor((checkpoint - origin) / 5) * 5
batch_end = min(batch_start + 4, page_count)
next_batch_start = batch_start + 5
next_batch_end = min(next_batch_start + 4, page_count)
```

### Decks and cards

`decks.revision` is a lightweight optimistic-concurrency token for deck metadata only. Creating a deck sets it to 1. A successful `update_deck` requires the latest revision and increments it by one. Adding, updating, or deleting cards does not change the deck metadata revision.

Each row in `card_revisions` is an immutable version of one logical card. Creation inserts revision 1. Update compares `expected_revision` with the latest revision, then inserts revision `N + 1` while copying any unchanged fields. Existing rows are never updated.

`source_pages` is stored in canonical CSV form, for example `"3,4,9"`: decimal 1-based page indices, ascending, unique, and without spaces. Go validates every index against the owning book. It is deliberately not expanded into tags, sources, or explanatory metadata.

Current cards are obtained through `current_cards`. Listing and export sort by `card_id`; the database stores no user-managed position.

`delete_card` hard-deletes every revision with the exact `(book_id, deck_id, card_id)` after checking the latest revision. `delete_deck` cascades through all card revisions in that deck.

## Python converter handoff

Go sends one JSON request through standard input containing:

- the SQLite database path;
- the PDF file reference;
- the book title;
- render settings.

The converter opens the PDF, extracts its outline, then opens the database and
begins one transaction. It lazily feeds rendered page BLOBs into
`sqlite3.executemany()` and inserts the flat outline after all pages. Memory
remains bounded to roughly one rendered page.

After commit, the converter writes one JSON `BookSummary` to standard output.
Missing outlines return `outline_required`. Extracted outlines that cannot be
stored unchanged return `outline_unusable`. Failures roll back the transaction,
write structured diagnostics to standard error, and exit nonzero. No temporary
page-image or CSV files are created. The source PDF is never deleted or
modified.

## Deliberately absent

Version one has no soft-delete columns, trash tables, agent identity table, chat/connection tracking, card revision timestamps, cursor revision, cursor timestamp, pending question, concepts-learned field, card ordering, tags, review scheduling, or migration framework. Those omissions are intentional, not unfinished columns to add opportunistically.
