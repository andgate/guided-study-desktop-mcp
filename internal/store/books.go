package store

import (
	"context"
	"database/sql"
)

func (s *Store) ListBooks(ctx context.Context) ([]BookSummary, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT book_id,title,page_count FROM books ORDER BY title COLLATE NOCASE,book_id`,
	)
	if err != nil {
		return nil, storageError(err)
	}
	defer rows.Close()

	books := []BookSummary{}
	for rows.Next() {
		var book BookSummary
		if err := rows.Scan(&book.BookID, &book.Title, &book.PageCount); err != nil {
			return nil, storageError(err)
		}
		books = append(books, book)
	}
	if err := rows.Err(); err != nil {
		return nil, storageError(err)
	}
	return books, nil
}

func (s *Store) GetBook(ctx context.Context, bookID string) (Book, error) {
	var book Book
	err := s.db.QueryRowContext(
		ctx,
		`SELECT book_id,title,page_count FROM books WHERE book_id=?`,
		bookID,
	).Scan(&book.BookID, &book.Title, &book.PageCount)
	if err == sql.ErrNoRows {
		return book, bookMissing(bookID)
	}
	if err != nil {
		return book, storageError(err)
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT book_id,outline_index,title,page_index
		 FROM book_outline WHERE book_id=? ORDER BY outline_index`,
		bookID,
	)
	if err != nil {
		return book, storageError(err)
	}
	defer rows.Close()

	book.Outline = []OutlineEntry{}
	for rows.Next() {
		var entry OutlineEntry
		if err := rows.Scan(
			&entry.BookID,
			&entry.OutlineIndex,
			&entry.Title,
			&entry.PageIndex,
		); err != nil {
			return book, storageError(err)
		}
		book.Outline = append(book.Outline, entry)
	}
	if err := rows.Err(); err != nil {
		return book, storageError(err)
	}
	return book, nil
}

func (s *Store) RenameBook(ctx context.Context, bookID, newTitle string) (BookSummary, error) {
	title, serviceErr := cleanRequired("new_title", newTitle)
	if serviceErr != nil {
		return BookSummary{}, serviceErr
	}

	result, err := s.db.ExecContext(
		ctx,
		`UPDATE books SET title=? WHERE book_id=?`,
		title,
		bookID,
	)
	if err != nil {
		return BookSummary{}, storageError(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return BookSummary{}, bookMissing(bookID)
	}

	var book BookSummary
	err = s.db.QueryRowContext(
		ctx,
		`SELECT book_id,title,page_count FROM books WHERE book_id=?`,
		bookID,
	).Scan(&book.BookID, &book.Title, &book.PageCount)
	if err != nil {
		return book, storageError(err)
	}
	return book, nil
}

func (s *Store) RemoveBook(ctx context.Context, bookID string) error {
	result, err := s.db.ExecContext(ctx, `DELETE FROM books WHERE book_id=?`, bookID)
	if err != nil {
		return storageError(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return bookMissing(bookID)
	}
	return nil
}

func (s *Store) ReadPages(
	ctx context.Context,
	bookID string,
	start, end int,
) (RenderedPages, error) {
	pages := RenderedPages{
		BookID:         bookID,
		StartPageIndex: start,
		EndPageIndex:   end,
		Pages:          []RenderedPage{},
	}
	if start > end {
		return pages, errf(
			CodeInvalidArgument,
			map[string]any{"start_page_index": start, "end_page_index": end},
			"Start page exceeds end page.",
		)
	}

	err := s.db.QueryRowContext(
		ctx,
		`SELECT page_count FROM books WHERE book_id=?`,
		bookID,
	).Scan(&pages.PageCount)
	if err == sql.ErrNoRows {
		return pages, bookMissing(bookID)
	}
	if err != nil {
		return pages, storageError(err)
	}
	if start < 1 || end > pages.PageCount {
		return pages, errf(
			CodeOutOfBounds,
			map[string]any{
				"start_page_index": start,
				"end_page_index":   end,
				"page_count":       pages.PageCount,
			},
			"Page range is outside this book.",
		)
	}

	rows, err := s.db.QueryContext(
		ctx,
		`SELECT page_index,mime_type,image_data FROM book_pages
		 WHERE book_id=? AND page_index BETWEEN ? AND ? ORDER BY page_index`,
		bookID,
		start,
		end,
	)
	if err != nil {
		return pages, storageError(err)
	}
	defer rows.Close()

	for rows.Next() {
		var page RenderedPage
		if err := rows.Scan(&page.PageIndex, &page.MIMEType, &page.ImageData); err != nil {
			return pages, storageError(err)
		}
		pages.Pages = append(pages.Pages, page)
	}
	if err := rows.Err(); err != nil {
		return pages, storageError(err)
	}

	if len(pages.Pages) != end-start+1 {
		return pages, errf(CodeNotFound, nil, "Requested page was not found.")
	}
	return pages, nil
}

func bookMissing(bookID string) *Error {
	return errf(
		CodeNotFound,
		map[string]any{"book_id": bookID},
		"Book %q was not found.",
		bookID,
	)
}
