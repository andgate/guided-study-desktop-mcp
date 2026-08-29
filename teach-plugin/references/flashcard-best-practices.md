# Flashcard Best Practices

## Objective

Create Q&A retrieval cards from the source scope requested by the learner.
The scope may be a section, chapter, page range, whole book, or any other
selection the learner requests.

The learner already understands the source. The cards preserve knowledge
worth recalling later.

Maximize coverage of operands. Do not reduce the deck to key facts, main
ideas, or an editorial sample.

## Knowledge selection

An **operand** is knowledge worth keeping readily recallable because it
supports reasoning, understanding, or performing tasks. Operands include
facts, definitions, principles, relationships, mechanisms, conditions,
distinctions, mappings, constraints, exceptions, and other reusable
knowledge.

An **operation** is something performed using knowledge, such as solving,
deriving, proving, calculating, programming, diagnosing, debugging,
designing, evaluating, or interpreting. Cards can preserve operands needed
for an operation. Recalling information about an operation does not preserve
the ability to perform it.

**External data** is exhaustive, volatile, rarely used, or cheaply retrieved
information better obtained from a reference, such as every value in a large
property table. External data can still contain operands, including its
meaning, organization, patterns, and useful landmarks.

When uncertain whether knowledge is an operand or external data, create the
card.

One operand may produce multiple cards when each card activates a separate,
meaningful retrieval target. Do not mechanically reverse every card.

## Source selection

Use the source as the authority for card meaning. Rewrite, reorganize, and
clarify its knowledge to create effective retrieval cards.

Inspect all material that teaches or develops the subject, including:

- explanatory prose;
- figures and diagrams;
- equations and tables;
- captions and callouts;
- labels and annotations;
- examples embedded in explanatory content;
- any other source material that conveys new subject knowledge.

Skip material whose primary purpose is navigation, repetition, assessment,
practice, reference, or publication apparatus, including:

- summaries and learning objectives;
- dedicated worked examples and exercises;
- glossaries and review questions;
- answer keys and bibliographies;
- any other source material that does not develop new subject knowledge.

An embedded example remains useful when it teaches an operand rather than
merely providing practice.

## Dynamic atomization

Atomize knowledge as far as possible while preserving fundamental couplings.
A **fundamental coupling** exists when separating the parts would lose the
useful relationship or produce underdetermined retrievals.

Word limits constrain atomization. Split an operand further when a clear,
independently useful retrieval can be made smaller. Keep its parts together
when their relationship is the knowledge being retrieved.

For lists and taxonomies, keep the complete membership together when the
whole structure is the retrieval target. Otherwise, create clearly cued cards
for independently useful members. Do not ask underdetermined questions such
as "What is one component of X?"

For ordered sequences, keep the sequence together when its order is the
retrieval target. Create separate cards for independently meaningful stages.

For mappings, create separate cards when each key clearly identifies its
value. Keep the mapping together when its organization or pattern is the
retrieval target.

For contrasts, keep both sides together when the distinction between them is
the retrieval target. Separate properties that remain meaningful on their
own.

For causal mechanisms, split a chain into independently meaningful links.
Given "A raises B, and B activates C," useful cards include:

- What does A raise? → B
- What does B activate? → C
- How does A activate C? → By raising B

The third card is useful only when retrieving the connecting mechanism is a
separate target.

A list, taxonomy, sequence, mapping, or other structure may support both a
whole-structure card and smaller cards when they test different retrievals.

## Card construction

- Each front asks for one clear retrieval target.
- Each front has one intended answer.
- Make each front understandable outside the source.
- Include only enough context to identify that target.
- Require production rather than recognition.
- Avoid yes/no questions and answer-bearing cues.
- Rewrite source prose aggressively for compact retrieval.
- Preserve technical meaning rather than textbook phrasing.
- Keep every condition, exception, negation, unit, threshold, and scope needed
  for the answer to remain true.
- Keep answers exact and minimal.

Questions: ≤10 words.

Answers: ≤5 words.

An answer may exceed five words when a fundamentally coupled structure cannot
be divided without losing the intended retrieval. Common cases include a
complete list or taxonomy, an ordered sequence, and a paired contrast.
Definitions and ordinary prose receive no special exception.

## Provenance

Each card contains:

- `front`
- `back`
- `source_pages`

Record every source page index that directly supports the card. Use one page
when it contains the complete support and multiple pages when support spans
them. Store page indices as ascending, comma-separated values without spaces.
