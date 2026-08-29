# Guided Study Plan 2

## Goal

Restore a linear, thorough guided-reading experience while keeping one
`@Noggin` entry point for study, flashcard creation, and learner help.

Remove study progress logs and the redundant `get_progress` workflow from the
prototype.

## Noggin routing

Keep one `noggin` skill. Use its `SKILL.md` as a concise router for three modes:

1. Guided reading.
2. Chapter flashcard creation.
3. Learner help and design explanation.

Load only the reference needed for the current mode.

## Guided reading

The guided-reading reference must direct the study agent to:

- move linearly through the book and each page;
- name the current section and page before every study question;
- ask one question at a time;
- use each question as an intellectual Easter egg hunt that focuses reading;
- ask as many sequential questions as the page requires for thorough coverage;
- wait for the learner's answer;
- report the evaluation of the learner's answer in the visible response;
- remain on the current question until the learner answers it fully;
- refine the current question when needed;
- ask another question after the current question is fully answered;
- answer learner questions, then resume the unanswered study question unless
  the learner directs otherwise;
- advance only after the current page is sufficiently covered or the learner
  explicitly changes direction;
- follow learner requests to skip, change pace, move, stop, or switch modes;
- restart the current page when a new chat resumes an interrupted page.

Use rendered page images as the source authority.

At chapter completion, offer a short, customized prompt for generating the
chapter's flashcards in a fresh chat.

## Flashcard creation

Keep flashcard creation independent from study sessions.

- Recommend a fresh chat when flashcards are requested during guided reading.
- Provide a short prompt such as
  `@Noggin generate a flashcard deck for Chapter 3 of LLMs for Dummies`.
- Customize the prompt for the requested book, chapter, and deck arrangement.
- Continue in the current chat when the learner explicitly insists.
- Read the complete requested source scope.
- Create one or more decks according to the learner's requested arrangement.
- Save completed batches as the source is processed.

## Learner guide

Add an on-demand learner guide that explains:

- how to start and resume guided reading;
- what the question loop feels like;
- why questions guide attention before and during reading;
- how answer discussion supports understanding;
- how pages and sections progress;
- how to skip, stop, resume, or change pace;
- why chapter flashcards belong in a fresh chat;
- how to request one deck or several logically divided decks.

The guide must describe the intended experience without loading during ordinary
study or flashcard creation.

## Remove progress logs

Remove progress logs and `get_progress` from the prototype:

- unregister the MCP tools;
- remove MCP input and output types;
- remove store methods and models;
- remove log-tail conflict handling;
- remove the SQL table and index from the creation schema;
- simplify session results to cursor and section state;
- update MCP instructions, README content, and technical documentation;
- remove tests that cover the deleted feature while preserving remaining
  session coverage.

Keep `docs/testing/test_1.md` as historical testing evidence.

No application migration is required. Manually and destructively drop the
`progress_logs` table from:

`C:\Users\andgate\AppData\Local\GuidedStudy\guided-study.db`

## Verification

Review the final diff and inspect the local database schema. Do not run tests or
build the project. Ask the user to run the appropriate verification commands
and share the output.
