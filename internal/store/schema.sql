PRAGMA foreign_keys = ON;

-- Books and page images.
CREATE TABLE IF NOT EXISTS books (
    book_id TEXT PRIMARY KEY,
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    page_count INTEGER NOT NULL CHECK (page_count >= 1)
);

CREATE TABLE IF NOT EXISTS book_pages (
    book_id TEXT NOT NULL,
    page_index INTEGER NOT NULL CHECK (page_index >= 1),
    mime_type TEXT NOT NULL CHECK (length(trim(mime_type)) > 0),
    image_data BLOB NOT NULL CHECK (length(image_data) > 0),
    PRIMARY KEY (book_id, page_index),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS book_outline (
    book_id TEXT NOT NULL,
    outline_index INTEGER NOT NULL CHECK (outline_index >= 0),
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    page_index INTEGER NOT NULL CHECK (page_index >= 1),
    PRIMARY KEY (book_id, outline_index),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, page_index) REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS book_outline_by_page
    ON book_outline (book_id, page_index, outline_index);

-- Reading sessions.
CREATE TABLE IF NOT EXISTS study_sessions (
    book_id TEXT NOT NULL,
    session_id TEXT NOT NULL,
    name TEXT NOT NULL COLLATE NOCASE CHECK (length(trim(name)) > 0),
    origin_page_index INTEGER NOT NULL CHECK (origin_page_index >= 1),
    checkpoint_page_index INTEGER NOT NULL CHECK (checkpoint_page_index >= 1),
    checkpoint_heading TEXT NULL CHECK (
        checkpoint_heading IS NULL OR length(trim(checkpoint_heading)) > 0
    ),
    PRIMARY KEY (book_id, session_id),
    UNIQUE (book_id, name),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE,
    FOREIGN KEY (book_id, origin_page_index) REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE,
    FOREIGN KEY (book_id, checkpoint_page_index) REFERENCES book_pages (book_id, page_index) ON DELETE CASCADE
) WITHOUT ROWID;

-- Flashcard decks and card revisions.
CREATE TABLE IF NOT EXISTS decks (
    book_id TEXT NOT NULL,
    deck_id TEXT NOT NULL,
    title TEXT NOT NULL CHECK (length(trim(title)) > 0),
    description TEXT NULL,
    revision INTEGER NOT NULL DEFAULT 1 CHECK (revision >= 1),
    PRIMARY KEY (book_id, deck_id),
    FOREIGN KEY (book_id) REFERENCES books (book_id) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE TABLE IF NOT EXISTS card_revisions (
    book_id TEXT NOT NULL,
    deck_id TEXT NOT NULL,
    card_id TEXT NOT NULL,
    revision INTEGER NOT NULL CHECK (revision >= 1),
    front TEXT NOT NULL CHECK (length(trim(front)) > 0),
    back TEXT NOT NULL CHECK (length(trim(back)) > 0),
    source_pages TEXT NOT NULL CHECK (length(trim(source_pages)) > 0),
    PRIMARY KEY (book_id, deck_id, card_id, revision),
    FOREIGN KEY (book_id, deck_id) REFERENCES decks (book_id, deck_id) ON DELETE CASCADE
) WITHOUT ROWID;

CREATE INDEX IF NOT EXISTS card_revisions_latest ON card_revisions (book_id, deck_id, card_id, revision DESC);

-- Latest revision of each card.
CREATE VIEW IF NOT EXISTS current_cards AS
SELECT candidate.book_id, candidate.deck_id, candidate.card_id, candidate.revision,
       candidate.front, candidate.back, candidate.source_pages
FROM card_revisions AS candidate
WHERE NOT EXISTS (
    SELECT 1 FROM card_revisions AS newer
    WHERE newer.book_id = candidate.book_id
      AND newer.deck_id = candidate.deck_id
      AND newer.card_id = candidate.card_id
      AND newer.revision > candidate.revision
);
