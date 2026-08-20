---
name: teach
description: Guide active, page-by-page learning from books prepared by the Guided Study MCP service. Use when a learner wants to start, continue, resume, or save a book study session; discuss rendered book pages; build understanding through one purposeful question at a time; or create and revise durable flashcards from material they understand.
---

# Teach

Use the `guided_study` MCP tools as the durable backend. Teach from rendered page images; never substitute extracted PDF text or invent page content.

## Start or resume

1. Call `list_books` unless the user already supplied an unambiguous `book_id`.
2. Call `get_book` for metadata and TOC context.
3. Call `list_sessions` for that book.
4. Resume an obvious session, ask the user when several sessions plausibly match, or call `create_session` with a meaningful name.
5. Call `get_progress`, then `current_page`.
6. Tell the learner the physical PDF page and ask one purposeful question grounded in that page image.

Do not create a new session merely because the chat is new. Sessions are durable study threads, not chat identities.

## Teach actively

- Keep exactly one active question.
- Direct attention toward meaning: mechanism, cause, distinction, relationship, assumption, consequence, figure, or application.
- Use one page by default and a few consecutive pages only when the idea crosses a boundary.
- Wait for the learner's answer. If incomplete, stay on the same question, explain or hint as needed, and let them refine it.
- Treat the learner's strongest current understanding as the useful result. Do not preserve temporary mistakes as durable knowledge.
- Answer learner questions fully, then return to the active question.
- Advance with `next_page`; use `goto_page` only for an explicit jump. Never calculate a hidden cursor outside the session.
- Use `set_current_section` when the semantic section becomes clear or changes. The TOC is navigation context, not a semantic authority.

Read [pedagogy.md](references/pedagogy.md) when choosing questions or deciding whether knowledge deserves a card.

## Persist bounded progress

Use `get_progress` before log mutations when the tail ID is not already certain.

At a checkpoint or when the learner stops:

1. Add or revise worthwhile cards first.
2. Call `log_progress` once with the exact continuation state and `expected_last_log_id`.
3. Report that the session is saved.

Log messages should say what was resolved and what to do next. The service already snapshots page, section, and time. Do not duplicate a full transcript or generate a separate narrative state model.

Use `amend_last_log` only to correct the actual tail. Use `delete_last_log` repeatedly from newest to oldest only when the user or conversation branch requires removing incorrect tail entries.

## Manage flashcards

- List existing decks before creating one when the appropriate deck is uncertain.
- Create focused cards only after the learner understands the material.
- Store physical page citations in canonical CSV such as `3,4,9`.
- Use the learner's best correct wording while keeping the page authoritative.
- Prefer important definitions, mechanisms, distinctions, relationships, sequences, rules, constraints, formulas, and diagram knowledge.
- Avoid anecdotes, duplicated prose, and details with little retrieval value.
- Use `expected_revision` for every deck/card update or delete.
- Read revision history when a conflict or wording question requires it; do not rewrite old revisions.

There is no inbox, archive, publication state, card order, tag model, or Quizlet canonical store.

## Resolve conflicts

- On `cursor_conflict`, inspect the actual page. If the intended move is still correct, retry from the actual cursor; otherwise keep the newer state.
- On `log_tail_conflict`, read the returned recent entries. Leave correct newer state alone, amend the actual tail, or delete unwanted tail entries newest-first before appending.
- On deck/card revision conflicts, read the latest deck or card history, reconcile wording, then retry with the latest revision.
- Never blindly repeat a stale mutation.

## Preserve service boundaries

The teaching agent decides questions, understanding, section semantics, card value, and card wording. The service stores deterministic state and images. Do not claim a page or concept is complete merely because the cursor moved.
