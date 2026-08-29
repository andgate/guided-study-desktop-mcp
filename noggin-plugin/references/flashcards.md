# Create flashcards

Create flashcards independently from guided-reading sessions. Do not require a
session or use its cursor.

If the request occurs during active guided reading, recommend a fresh chat and
give the learner a prompt they can paste directly. Name the actual book, source
scope, and deck arrangement. For example:

`@Noggin generate a flashcard deck for Chapter 3 of LLMs for Dummies`

Continue in the current chat when the learner insists.

## Generate decks

1. Use the learner's requested source scope and deck arrangement.
2. Call `list_books` unless the learner supplied an unambiguous `book_id`.
3. Call `get_book` and use its table of contents to locate the complete scope.
4. List existing decks when the destination is uncertain.
5. Create or select every requested destination deck.
6. Call `read_pages` for the next five consecutive source pages.
7. Extract every complete retrieval target supported by the batch.
8. Save the batch immediately through `add_cards`, once per receiving deck.
9. Repeat steps 6–8 with the next consecutive page batch until every page in
   the requested scope has been processed.

Use one deck unless the learner requests several logically divided decks.
Honor the requested divisions throughout the source scope.

## Manage decks

- Store source page indices in canonical CSV such as `3,4,9`.
- Use the clearest correct wording while keeping the rendered pages
  authoritative.
- Supply `expected_revision` when updating or deleting a deck or card.
- Deleting a card permanently removes all of its revisions.
- Review revision history when an edit makes a card worse, then use an earlier
  version to repair it.
- Do not rewrite old revisions.

If an update or delete returns a revision conflict, read the current deck state
before deciding whether to retry.

Read [flashcard-best-practices.md](flashcard-best-practices.md) in full before
creating or revising cards.
