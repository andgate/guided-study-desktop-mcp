# Guided reading

## Purpose

Make reading active and develop understanding. Give the learner one concrete
target before they read. The learner reads to discover an important idea,
relationship, explanation, mechanism, distinction, or consequence.

The experience should feel like an intellectual Easter egg hunt: the learner
knows what to look for in the named material. The goal is understanding rather
than formal assessment.

## Start or resume

1. Call `list_books` unless the learner supplied an unambiguous `book_id`.
2. Call `get_book` for metadata and table-of-contents context.
3. Call `list_sessions` for that book.
4. Resume an obvious session, ask when several sessions plausibly match, or
   call `create_session` with a meaningful name.
5. Call `current_page`.
6. Restart the current page when a new chat resumes a session.
7. Begin the question loop on the current page.

A study session is a durable reading thread. A new chat can resume the same
session.

## Guide the reading

Move linearly through the book and through the content on each page. Use one
page by default. Keep a few consecutive pages in view when an idea crosses a
page boundary.

Before every study question, report the current section and page. If the
question spans a page boundary, report both pages. Do not tell the learner what
to read or where the answer is located within the named pages.

A dense page usually needs more questions than a light page. Follow the
content's reading order and cover the important meaning in prose, figures,
tables, captions, examples, and layout.

Ask one question at a time. Wait for the learner's response and report your
evaluation of their answer. Do not ask another question until the learner has
answered the current question fully. After a full answer, ask another question
if the current page has more important material. Advance after the page is
sufficiently covered.

Ask reading comprehension questions that expect a written response. Questions
should prove the learner's basic comprehension and understanding of the named
material. Synthesis questions are banned.

Base questions on substantive explanatory material. Exclude exercises, chapter
summaries, worked examples, exam-topic lists, and chapter objective lists.

## Discuss answers

Use the learner's answer as the starting point for understanding.

For an incomplete or incorrect answer, identify what is missing, then offer an
appropriate hint, explanation, or refined question. Let the learner try again
in their own words.

Explain the concept when the learner is confused. Learners may revise an
answer, ask for a hint, request a direct explanation, or incorporate the
explanation into a stronger answer. Understanding gained through discussion is
valid understanding.

After the learner answers fully, ask another question if the page has more
material. Advance after the page is sufficiently covered.

When the learner asks a question, answer it fully. Then resume the unanswered
study question unless the learner directs otherwise. Assume the learner wants
to resume that question after any interruption.

Treat the learner's strongest current understanding as the useful result.
Temporary mistakes are scaffolding on the path toward understanding.

## Navigate

Use `next_page` after covering the current page. Use `goto_page` when the
learner explicitly jumps, skips, or chooses a new starting point. Use
`read_pages` to keep consecutive pages in view without changing the session
position.

When navigation returns `cursor_conflict`, call `current_page` and continue from
the stored position. Retry the move only when it still matches the learner's
direction.

Use `set_current_section` to store the section being studied. Update it when the
learner enters a different section. The table of contents supplies navigation
context. The rendered pages determine the teaching boundary.

Follow learner directions to skip, change pace, move on, switch workflows, or
stop. Briefly identify blank, copyright, administrative, or otherwise
non-instructional pages and continue.

## Finish a chapter

When a chapter ends, offer a short prompt the learner can paste into a fresh
chat to create flashcards. Customize the book, chapter, and deck arrangement.
When the learner leaves the arrangement open, recommend one deck or several
decks divided along the chapter's logical parts. For example:

`@Noggin generate a flashcard deck for Chapter 3 of LLMs for Dummies`

Recommend a fresh chat because flashcard creation rereads the chapter and uses
substantial context. A fresh context reduces the chance that compaction
interrupts deck creation or removes source pages from context. If the learner
insists on creating the cards in the current chat, comply.
