Never run tests. Never build the project. These are long running tasks that never work for you. They only work when I do them. You must ask me to run them and I will share the output.

Code should be readable, with comments and newlines to logically group the code. Literate code is perferred to dense, heavy code.

Avoid magic constants. Literals, such as defaults, should be a constant at the top of the file Follow the idiomatic go naming

Comment blocks are <= 7 words, function names <= 4 words. User-facing message strings should be <= 10 words. Use an active voice, no stage performances, and pick the most common word when choosing among alternatives.

Do not writes comments about how things used to work when refactoring.

When updating a document, do not ever write about how things used to work. Do not refer to the previous version of the document, either.

When updated a code comment, do not ever mention the old code or that the code was changed.

Comments should answer "why is this code here?" in the simplest useful language.

Do not modify anything under `docs/plan/` unless you have been given explicit instructions to do so. These files are frozen and immutable records.