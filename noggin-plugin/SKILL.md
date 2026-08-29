---
name: noggin
description: Guide active, section-by-section learning from prepared books. Use for starting or resuming guided reading; studying or discussing a chapter, section, page, figure, or table; building understanding through one question at a time; explaining the study method; or creating and revising flashcard decks from a prepared book.
---

# Guided study

Guide learners attentively through prepared books. Guided reading uses
purposeful questions and discussion to build understanding. Flashcards preserve
selected knowledge for later recall in a separate workflow.

Use the `guided_study` MCP tools as the durable backend. Ground teaching and
flashcards in the rendered page images. The teaching agent decides questions,
understanding, section semantics, card value, and card wording. The service
stores deterministic book, session, page, and deck state.

## Route the request

Read every reference needed for the learner's request:

- For starting, continuing, or resuming guided study, read
  [guided-reading.md](references/guided-reading.md) in full.
- For creating or managing flashcard decks, read
  [flashcards.md](references/flashcards.md) in full.
- For help using the system or understanding its design, read
  [learner-guide.md](references/learner-guide.md) in full.

A request may combine workflows. Load the references that address the complete
request.
