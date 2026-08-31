---
name: noggin
description: Prepare PDF books and guide active, page-by-page learning with Guided Study. Use for importing a book; starting or resuming guided reading; studying or discussing a chapter, section, page, figure, or table; building understanding through one question at a time; explaining the study method; or creating and revising flashcard decks from a prepared book.
---

# Guided study

Guide learners attentively through prepared books. Guided reading uses
purposeful questions and discussion to build understanding. Flashcards preserve
selected knowledge for later recall in a separate workflow.

Use the `noggin_mcp` MCP tools as the durable backend. Ground teaching and
flashcards in the rendered page images. The teaching agent decides questions,
understanding, heading semantics, card value, and card wording. The service
stores deterministic book, session, outline, and deck state.

## MCP server guidelines

The server keeps no active book or session. Supply the exact stored IDs each
tool requires; never substitute a title or name for an ID.

All page indices are 1-based physical PDF pages.

## Route the request

Read every reference needed for the learner's request:

- For starting, continuing, or resuming guided study, read
  [guided-reading.md](references/guided-reading.md) in full.
- For preparing or importing a PDF book, read
  [book-import.md](references/book-import.md) in full.
- For creating or managing flashcard decks, read
  [flashcards.md](references/flashcards.md) in full.
- For help using the system or understanding its design, read
  [learner-guide.md](references/learner-guide.md) in full.

A request may combine workflows. Load the references that address the complete
request.
