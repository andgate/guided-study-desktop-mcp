# Create flashcards

Create flashcards independently from guided-reading sessions. Do not require a
session or use its cursor.

```mermaid
flowchart TD
    A["Flashcard request"] --> B{"Is guided reading active in this chat?"}
    B -- "No" --> C["Begin flashcard creation"]
    B -- "Yes" --> D["Recommend a fresh chat and provide a customized prompt"]
    D --> E{"Does the learner prefer the fresh chat?"}
    E -- "Yes" --> F["Continue there with the provided prompt"]
    E -- "No" --> C
```

Name the actual book, source scope, and deck arrangement in the prompt. For
example:

`@Noggin generate a flashcard deck for Chapter 3 of LLMs for Dummies`

## Generate decks

```mermaid
flowchart TD
    A["Resolve the requested book and complete source scope with list_books and get_book"] --> B{"Is the destination deck arrangement clear?"}
    B -- "No" --> C["List existing decks"]
    B -- "Yes" --> D["Create or select every destination deck"]
    C --> D
    D --> E["Call read_pages for the next five consecutive source pages"]
    E --> F["Extract every complete retrieval target supported by the batch"]
    F --> G["Assign each target to its receiving deck"]
    G --> H["Call add_cards once per receiving deck"]
    H --> I{"Does the requested source scope have pages remaining?"}
    I -- "Yes" --> E
    I -- "No" --> J["Finish deck creation"]
```

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
