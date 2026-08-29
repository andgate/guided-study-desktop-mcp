# Flashcard Implementation Plan

## Goal

Support autonomous flashcard generation from any source scope requested by
the learner. The agent reads rendered page images in batches and persists each
batch's cards directly to the requested deck or decks.

Card-generation policy belongs in
[`flashcard-best-practices.md`](flashcard-best-practices.md).

## Current MCP surface

The reading tools return one rendered page per call:

- `current_page` reads the study-session cursor.
- `next_page` advances the cursor once.
- `prev_page` moves the cursor back once.
- `goto_page` moves the cursor to one page.

These tools depend on a study session. Navigation calls mutate the learner's
cursor. They do not support flashcard generation across page batches.

The card tools already support the required incremental write shape:

- `create_deck` creates a deck with optional cards.
- `add_cards` validates and writes one card batch atomically.
- Each card contains `front`, `back`, and `source_pages`.

Decks have no generation-completion state.

## Batch page reader

Add a read-only, session-independent `read_pages` tool.

Input:

```text
book_id: string
start_page_index: integer
end_page_index: integer
```

The range is inclusive.

Requirements:

- Accept a stable `book_id`.
- Accept an explicit page range.
- Support arbitrary valid range sizes.
- Return rendered images for every requested page.
- Return each page's `page_index`.
- Return the book's `page_count`.
- Return TOC context for each page.
- Preserve page order.
- Reject invalid, reversed, or out-of-range bounds.
- Never create or require a study session.
- Never read or move a learner's cursor.

The flashcard workflow requests five pages per batch. The server does not
enforce that workflow-specific batch size.

## Batch result

Return structured metadata plus MCP image content.

```text
PageBatch {
  book_id: string
  page_count: integer
  start_page_index: integer
  end_page_index: integer
  pages: PageBatchEntry[]
}

PageBatchEntry {
  page_index: integer
  toc_context: TOCEntry[]
}
```

Return one image block for each `pages` entry. Image blocks and page entries
use the same ascending order. Include a `Page N.` text label before each image
so the page index remains explicit in model context.

## Agent workflow

1. Use the existing book and TOC tools to locate the requested source scope.
2. Create or select the requested deck or decks.
3. Read the next five pages within the requested scope.
4. Extract every complete retrieval target available from the source in
   context.
5. Add that batch's cards with one `add_cards` call per destination deck.
6. Repeat without user pauses until the requested scope is exhausted.

The generation prompt governs source filtering, operand selection,
atomization, wording, and provenance.

The MCP server does not track flashcard-generation state, completion, or
recovery. Existing book, TOC, deck, and card tools continue to provide their
current deterministic behavior.

## Code changes

### Store

- Add a session-independent rendered-page model.
- Add a range reader in `internal/store/books.go`.
- Query pages by inclusive bounds and ascending index.
- Load TOC context for each returned page.

### MCP server

- Add the range input and batch output types in
  `internal/mcpserver/reading.go`.
- Register `read_pages` with read-only tool annotations.
- Return ordered text labels and image blocks.
- Use the existing structured error format.

### Documentation and skill

- Document `read_pages` in `docs/MCP.md`.
- Add the five-page extraction loop to the teaching skill.
- Keep flashcard-writing policy grounded in
  `docs/flashcard-best-practices.md`.

## Verification

Ask the user to run the project checks and share the output. Do not run tests
or build the project from the agent environment.
