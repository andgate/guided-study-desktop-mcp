# Book import

## Import the book

Call `import_book` with the local book reference supplied by the host and the
book title. PDF and EPUB sources are accepted:

```text
file_reference: string
title: string
```

The service renders every page and extracts the outline in
one atomic import. It stores the outline as a flat ordered list containing the
outline index, title, and 1-based physical page index.

Do not create, repair, validate, enrich, or reconcile an outline. The extracted
outline is the only accepted navigation outline.

If `import_book` returns `outline_required`, report that the book has no
extractable outline and stop. If it returns `outline_unusable`, report that the
outline cannot be stored and stop. Do not offer an agent-generated outline
or another import fallback.

After a successful import, use `get_book` when the outline is needed for
navigation. Parse `outline_csv` as RFC 4180 CSV with this header:

```text
outline_index,title,page_index
```

Outline entries help locate requested headings and pages. They do not define
teaching boundaries or checkpoint headings.

## Reflowable books

EPUB has no fixed pages. The service lays each reflowable book out at one page
size, so its page indices are canonical here and match no printed edition.
Locate material by heading rather than by a page number the learner supplies.
