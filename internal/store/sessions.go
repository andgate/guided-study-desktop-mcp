package store

import (
	"context"
	"database/sql"

	"github.com/google/uuid"
)

const readingBatchSize = 5

type rowScanner interface {
	Scan(...any) error
}

func (s *Store) CreateSession(
	ctx context.Context,
	bookID, name string,
	startPage int,
) (CreatedSession, error) {
	name, serviceErr := cleanRequired("name", name)
	if serviceErr != nil {
		return CreatedSession{}, serviceErr
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CreatedSession{}, storageError(err)
	}
	defer tx.Rollback()

	pageCount, err := txPageCount(ctx, tx, bookID)
	if err != nil {
		return CreatedSession{}, err
	}
	if err := validPage(startPage, pageCount); err != nil {
		return CreatedSession{}, err
	}

	selection, err := selectPages(ctx, tx, bookID, "", startPage, pageCount)
	if err != nil {
		return CreatedSession{}, err
	}

	session := SessionSummary{
		BookID:              bookID,
		SessionID:           uuid.NewString(),
		Name:                name,
		OriginPageIndex:     startPage,
		CheckpointPageIndex: startPage,
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO study_sessions(
		 book_id,session_id,name,origin_page_index,
		 checkpoint_page_index,checkpoint_heading
		) VALUES(?,?,?,?,?,NULL)`,
		session.BookID,
		session.SessionID,
		session.Name,
		session.OriginPageIndex,
		session.CheckpointPageIndex,
	)
	if isConstraint(err) {
		return CreatedSession{}, errf(
			CodeAlreadyExists,
			map[string]any{"book_id": bookID, "name": name},
			"Session %q already exists.",
			name,
		)
	}
	if err != nil {
		return CreatedSession{}, storageError(err)
	}
	if err := tx.Commit(); err != nil {
		return CreatedSession{}, storageError(err)
	}

	selection.SessionID = session.SessionID
	return CreatedSession{Session: session, Selection: selection}, nil
}

func (s *Store) ListSessions(
	ctx context.Context,
	bookID string,
) ([]SessionSummary, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT book_id,session_id,name,origin_page_index,
		 checkpoint_page_index,checkpoint_heading
		 FROM study_sessions WHERE book_id=?
		 ORDER BY name COLLATE NOCASE,session_id`,
		bookID,
	)
	if err != nil {
		return nil, storageError(err)
	}
	defer rows.Close()

	sessions := []SessionSummary{}
	for rows.Next() {
		session, err := scanSession(rows)
		if err != nil {
			return nil, storageError(err)
		}
		sessions = append(sessions, session)
	}
	if err := rows.Err(); err != nil {
		return nil, storageError(err)
	}
	if len(sessions) > 0 {
		return sessions, nil
	}

	var exists int
	err = s.db.QueryRowContext(ctx, `SELECT 1 FROM books WHERE book_id=?`, bookID).
		Scan(&exists)
	if err == sql.ErrNoRows {
		return nil, bookMissing(bookID)
	}
	if err != nil {
		return nil, storageError(err)
	}
	return sessions, nil
}

func (s *Store) RenameSession(
	ctx context.Context,
	bookID, sessionID, newName string,
) (SessionSummary, error) {
	name, serviceErr := cleanRequired("new_name", newName)
	if serviceErr != nil {
		return SessionSummary{}, serviceErr
	}

	result, err := s.db.ExecContext(
		ctx,
		`UPDATE study_sessions SET name=? WHERE book_id=? AND session_id=?`,
		name,
		bookID,
		sessionID,
	)
	if isConstraint(err) {
		return SessionSummary{}, errf(
			CodeAlreadyExists,
			map[string]any{"book_id": bookID, "name": name},
			"Session %q already exists.",
			name,
		)
	}
	if err != nil {
		return SessionSummary{}, storageError(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return SessionSummary{}, sessionMissing(bookID, sessionID)
	}
	return s.sessionSummary(ctx, bookID, sessionID)
}

func (s *Store) DeleteSession(
	ctx context.Context,
	bookID, sessionID string,
) error {
	result, err := s.db.ExecContext(
		ctx,
		`DELETE FROM study_sessions WHERE book_id=? AND session_id=?`,
		bookID,
		sessionID,
	)
	if err != nil {
		return storageError(err)
	}
	count, _ := result.RowsAffected()
	if count == 0 {
		return sessionMissing(bookID, sessionID)
	}
	return nil
}

func (s *Store) GotoPage(
	ctx context.Context,
	bookID, sessionID string,
	pageIndex int,
) (PageSelection, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PageSelection{}, storageError(err)
	}
	defer tx.Rollback()

	if _, err := txSession(ctx, tx, bookID, sessionID); err != nil {
		return PageSelection{}, err
	}
	pageCount, err := txPageCount(ctx, tx, bookID)
	if err != nil {
		return PageSelection{}, err
	}
	if err := validPage(pageIndex, pageCount); err != nil {
		return PageSelection{}, err
	}

	selection, err := selectPages(
		ctx,
		tx,
		bookID,
		sessionID,
		pageIndex,
		pageCount,
	)
	if err != nil {
		return PageSelection{}, err
	}

	_, err = tx.ExecContext(
		ctx,
		`UPDATE study_sessions
		 SET origin_page_index=?,checkpoint_page_index=?,checkpoint_heading=NULL
		 WHERE book_id=? AND session_id=?`,
		pageIndex,
		pageIndex,
		bookID,
		sessionID,
	)
	if err != nil {
		return PageSelection{}, storageError(err)
	}
	if err := tx.Commit(); err != nil {
		return PageSelection{}, storageError(err)
	}
	return selection, nil
}

func (s *Store) ContinueReading(
	ctx context.Context,
	bookID, sessionID string,
	checkpoint Checkpoint,
) (ReadingBatch, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return ReadingBatch{}, storageError(err)
	}
	defer tx.Rollback()

	session, pageCount, checkpoint, err := validCheckpoint(
		ctx,
		tx,
		bookID,
		sessionID,
		checkpoint,
	)
	if err != nil {
		return ReadingBatch{}, err
	}

	current := batchRange(
		session.OriginPageIndex,
		checkpoint.PageIndex,
		pageCount,
	)
	nextStart := current.StartPageIndex + readingBatchSize
	if nextStart > pageCount {
		return ReadingBatch{}, errf(
			CodeNoNextBatch,
			map[string]any{
				"book_id":    bookID,
				"session_id": sessionID,
				"page_index": checkpoint.PageIndex,
			},
			"No later page batch exists.",
		)
	}

	next := PageRange{
		StartPageIndex: nextStart,
		EndPageIndex:   min(nextStart+readingBatchSize-1, pageCount),
	}
	pages, err := txPages(
		ctx,
		tx,
		bookID,
		next.StartPageIndex,
		next.EndPageIndex,
	)
	if err != nil {
		return ReadingBatch{}, err
	}
	if _, err := saveTx(ctx, tx, session, checkpoint); err != nil {
		return ReadingBatch{}, err
	}
	if err := tx.Commit(); err != nil {
		return ReadingBatch{}, storageError(err)
	}

	return ReadingBatch{
		BookID:    bookID,
		SessionID: sessionID,
		Batch:     next,
		Pages:     pages,
	}, nil
}

func (s *Store) SaveCheckpoint(
	ctx context.Context,
	bookID, sessionID string,
	checkpoint Checkpoint,
) (SessionSummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return SessionSummary{}, storageError(err)
	}
	defer tx.Rollback()

	session, _, checkpoint, err := validCheckpoint(
		ctx,
		tx,
		bookID,
		sessionID,
		checkpoint,
	)
	if err != nil {
		return SessionSummary{}, err
	}
	session, err = saveTx(ctx, tx, session, checkpoint)
	if err != nil {
		return SessionSummary{}, err
	}
	if err := tx.Commit(); err != nil {
		return SessionSummary{}, storageError(err)
	}
	return session, nil
}

func selectPages(
	ctx context.Context,
	tx *sql.Tx,
	bookID, sessionID string,
	startPage, pageCount int,
) (PageSelection, error) {
	pageRange := PageRange{
		StartPageIndex: startPage,
		EndPageIndex:   min(startPage+readingBatchSize-1, pageCount),
	}
	pages, err := txPages(
		ctx,
		tx,
		bookID,
		pageRange.StartPageIndex,
		pageRange.EndPageIndex,
	)
	if err != nil {
		return PageSelection{}, err
	}
	return PageSelection{
		BookID:    bookID,
		SessionID: sessionID,
		Batch:     pageRange,
		Pages:     pages,
	}, nil
}

func batchRange(origin, checkpoint, pageCount int) PageRange {
	offset := (checkpoint - origin) / readingBatchSize
	start := origin + offset*readingBatchSize
	return PageRange{
		StartPageIndex: start,
		EndPageIndex:   min(start+readingBatchSize-1, pageCount),
	}
}

func saveTx(
	ctx context.Context,
	tx *sql.Tx,
	session SessionSummary,
	checkpoint Checkpoint,
) (SessionSummary, error) {
	_, err := tx.ExecContext(
		ctx,
		`UPDATE study_sessions
		 SET checkpoint_page_index=?,checkpoint_heading=?
		 WHERE book_id=? AND session_id=?`,
		checkpoint.PageIndex,
		checkpoint.Heading,
		session.BookID,
		session.SessionID,
	)
	if err != nil {
		return SessionSummary{}, storageError(err)
	}

	session.CheckpointPageIndex = checkpoint.PageIndex
	session.CheckpointHeading = &checkpoint.Heading
	return session, nil
}

func validCheckpoint(
	ctx context.Context,
	tx *sql.Tx,
	bookID, sessionID string,
	checkpoint Checkpoint,
) (SessionSummary, int, Checkpoint, error) {
	heading, serviceErr := cleanRequired("heading", checkpoint.Heading)
	if serviceErr != nil {
		return SessionSummary{}, 0, Checkpoint{}, serviceErr
	}
	checkpoint.Heading = heading

	session, err := txSession(ctx, tx, bookID, sessionID)
	if err != nil {
		return SessionSummary{}, 0, Checkpoint{}, err
	}
	pageCount, err := txPageCount(ctx, tx, bookID)
	if err != nil {
		return SessionSummary{}, 0, Checkpoint{}, err
	}
	if err := validPage(checkpoint.PageIndex, pageCount); err != nil {
		return SessionSummary{}, 0, Checkpoint{}, err
	}
	if checkpoint.PageIndex < session.OriginPageIndex {
		return SessionSummary{}, 0, Checkpoint{}, errf(
			CodeInvalidArgument,
			map[string]any{
				"origin_page_index": session.OriginPageIndex,
				"page_index":        checkpoint.PageIndex,
			},
			"Checkpoint precedes the session origin.",
		)
	}
	return session, pageCount, checkpoint, nil
}

func validPage(pageIndex, pageCount int) error {
	if pageIndex >= 1 && pageIndex <= pageCount {
		return nil
	}
	return errf(
		CodeOutOfBounds,
		map[string]any{
			"page_index": pageIndex,
			"page_count": pageCount,
		},
		"Page is outside this book.",
	)
}

func txPageCount(
	ctx context.Context,
	tx *sql.Tx,
	bookID string,
) (int, error) {
	var pageCount int
	err := tx.QueryRowContext(
		ctx,
		`SELECT page_count FROM books WHERE book_id=?`,
		bookID,
	).Scan(&pageCount)
	if err == sql.ErrNoRows {
		return 0, bookMissing(bookID)
	}
	if err != nil {
		return 0, storageError(err)
	}
	return pageCount, nil
}

func txSession(
	ctx context.Context,
	tx *sql.Tx,
	bookID, sessionID string,
) (SessionSummary, error) {
	row := tx.QueryRowContext(
		ctx,
		`SELECT book_id,session_id,name,origin_page_index,
		 checkpoint_page_index,checkpoint_heading
		 FROM study_sessions WHERE book_id=? AND session_id=?`,
		bookID,
		sessionID,
	)
	session, err := scanSession(row)
	if err == sql.ErrNoRows {
		return session, sessionMissing(bookID, sessionID)
	}
	if err != nil {
		return session, storageError(err)
	}
	return session, nil
}

func scanSession(row rowScanner) (SessionSummary, error) {
	var session SessionSummary
	var heading sql.NullString
	err := row.Scan(
		&session.BookID,
		&session.SessionID,
		&session.Name,
		&session.OriginPageIndex,
		&session.CheckpointPageIndex,
		&heading,
	)
	session.CheckpointHeading = scanNullableString(heading)
	return session, err
}

func txPages(
	ctx context.Context,
	tx *sql.Tx,
	bookID string,
	start, end int,
) ([]RenderedPage, error) {
	rows, err := tx.QueryContext(
		ctx,
		`SELECT page_index,mime_type,image_data FROM book_pages
		 WHERE book_id=? AND page_index BETWEEN ? AND ? ORDER BY page_index`,
		bookID,
		start,
		end,
	)
	if err != nil {
		return nil, storageError(err)
	}
	defer rows.Close()

	pages := []RenderedPage{}
	for rows.Next() {
		var page RenderedPage
		if err := rows.Scan(&page.PageIndex, &page.MIMEType, &page.ImageData); err != nil {
			return nil, storageError(err)
		}
		pages = append(pages, page)
	}
	if err := rows.Err(); err != nil {
		return nil, storageError(err)
	}
	if len(pages) != end-start+1 {
		return nil, errf(CodeNotFound, nil, "Requested page was not found.")
	}
	return pages, nil
}

func (s *Store) sessionSummary(
	ctx context.Context,
	bookID, sessionID string,
) (SessionSummary, error) {
	row := s.db.QueryRowContext(
		ctx,
		`SELECT book_id,session_id,name,origin_page_index,
		 checkpoint_page_index,checkpoint_heading
		 FROM study_sessions WHERE book_id=? AND session_id=?`,
		bookID,
		sessionID,
	)
	session, err := scanSession(row)
	if err == sql.ErrNoRows {
		return session, sessionMissing(bookID, sessionID)
	}
	if err != nil {
		return session, storageError(err)
	}
	return session, nil
}

func sessionMissing(bookID, sessionID string) *Error {
	return errf(
		CodeNotFound,
		map[string]any{"book_id": bookID, "session_id": sessionID},
		"Study session was not found.",
	)
}
