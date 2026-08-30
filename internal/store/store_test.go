package store

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
)

func testStore(t *testing.T) *Store {
	t.Helper()
	st, err := Open(context.Background(), filepath.Join(t.TempDir(), "test.db"))
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = st.Close() })
	return st
}

func testBook(t *testing.T, st *Store) BookSummary {
	t.Helper()
	ctx := context.Background()
	tx, err := st.db.BeginTx(ctx, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer tx.Rollback()

	book := BookSummary{BookID: "test-book", Title: "Test Book", PageCount: 17}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO books(book_id,title,page_count) VALUES(?,?,?)`,
		book.BookID,
		book.Title,
		book.PageCount,
	)
	if err != nil {
		t.Fatal(err)
	}
	for page := 1; page <= book.PageCount; page++ {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO book_pages(book_id,page_index,mime_type,image_data)
			 VALUES(?,?,?,?)`,
			book.BookID,
			page,
			"image/jpeg",
			[]byte{byte(page)},
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	outline := []OutlineEntry{
		{
			BookID:       book.BookID,
			OutlineIndex: 0,
			Title:        "Opening",
			PageIndex:    1,
		},
		{
			BookID:       book.BookID,
			OutlineIndex: 1,
			Title:        "Closing",
			PageIndex:    13,
		},
	}
	for _, entry := range outline {
		_, err = tx.ExecContext(
			ctx,
			`INSERT INTO book_outline(book_id,outline_index,title,page_index)
			 VALUES(?,?,?,?)`,
			entry.BookID,
			entry.OutlineIndex,
			entry.Title,
			entry.PageIndex,
		)
		if err != nil {
			t.Fatal(err)
		}
	}
	if err := tx.Commit(); err != nil {
		t.Fatal(err)
	}
	return book
}

func requireCode(t *testing.T, err error, code string) {
	t.Helper()
	var serviceErr *Error
	if !errors.As(err, &serviceErr) {
		t.Fatalf("expected service error, got %v", err)
	}
	if serviceErr.Code != code {
		t.Fatalf("expected code %s, got %s", code, serviceErr.Code)
	}
}

func TestPageReading(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	book := testBook(t, st)
	created, err := st.CreateSession(ctx, book.BookID, "Main", 8)
	if err != nil {
		t.Fatal(err)
	}
	session := created.Session
	if session.OriginPageIndex != 8 ||
		session.CheckpointPageIndex != 8 ||
		session.CheckpointHeading != nil {
		t.Fatalf("unexpected checkpoint: %+v", session)
	}
	if created.Selection.Batch.StartPageIndex != 8 ||
		created.Selection.Batch.EndPageIndex != 12 ||
		len(created.Selection.Pages) != 5 {
		t.Fatalf("unexpected initial batch: %+v", created.Selection)
	}

	second, err := st.ContinueReading(
		ctx,
		book.BookID,
		session.SessionID,
		Checkpoint{PageIndex: 10, Heading: "Opening"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if second.Batch.StartPageIndex != 13 ||
		second.Batch.EndPageIndex != 17 ||
		len(second.Pages) != 5 {
		t.Fatalf("unexpected next batch: %+v", second)
	}
	progress, err := st.sessionSummary(ctx, book.BookID, session.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if progress.CheckpointPageIndex != 10 ||
		progress.CheckpointHeading == nil ||
		*progress.CheckpointHeading != "Opening" {
		t.Fatalf("continue did not save checkpoint: %+v", progress)
	}

	saved, err := st.SaveCheckpoint(
		ctx,
		book.BookID,
		session.SessionID,
		Checkpoint{PageIndex: 17, Heading: "Closing"},
	)
	if err != nil {
		t.Fatal(err)
	}
	if saved.CheckpointPageIndex != 17 ||
		saved.CheckpointHeading == nil ||
		*saved.CheckpointHeading != "Closing" {
		t.Fatalf("checkpoint was not saved: %+v", saved)
	}

	_, err = st.ContinueReading(
		ctx,
		book.BookID,
		session.SessionID,
		Checkpoint{PageIndex: 15, Heading: "Closing"},
	)
	requireCode(t, err, CodeNoNextBatch)
	unchanged, err := st.sessionSummary(ctx, book.BookID, session.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if unchanged.CheckpointPageIndex != 17 {
		t.Fatalf("failed continue changed checkpoint: %+v", unchanged)
	}

	selection, err := st.GotoPage(ctx, book.BookID, session.SessionID, 6)
	if err != nil {
		t.Fatal(err)
	}
	if selection.Batch.StartPageIndex != 6 ||
		selection.Batch.EndPageIndex != 10 ||
		len(selection.Pages) != 5 {
		t.Fatalf("goto did not reset progress: %+v", selection)
	}
	reset, err := st.sessionSummary(ctx, book.BookID, session.SessionID)
	if err != nil {
		t.Fatal(err)
	}
	if reset.OriginPageIndex != 6 ||
		reset.CheckpointPageIndex != 6 ||
		reset.CheckpointHeading != nil {
		t.Fatalf("goto did not reset checkpoint: %+v", reset)
	}

	_, err = st.SaveCheckpoint(
		ctx,
		book.BookID,
		session.SessionID,
		Checkpoint{PageIndex: 5, Heading: "Earlier"},
	)
	requireCode(t, err, CodeInvalidArgument)
	_, err = st.SaveCheckpoint(
		ctx,
		book.BookID,
		session.SessionID,
		Checkpoint{PageIndex: 6, Heading: ""},
	)
	requireCode(t, err, CodeInvalidArgument)
}

func TestSessionNames(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	book := testBook(t, st)
	stored, err := st.GetBook(ctx, book.BookID)
	if err != nil {
		t.Fatal(err)
	}
	if len(stored.Outline) != 2 || stored.Outline[1].PageIndex != 13 {
		t.Fatalf("outline was not loaded: %+v", stored.Outline)
	}

	_, err = st.CreateSession(ctx, book.BookID, "Missing", 0)
	requireCode(t, err, CodeOutOfBounds)

	created, err := st.CreateSession(ctx, book.BookID, "Main", 1)
	if err != nil {
		t.Fatal(err)
	}
	session := created.Session
	_, err = st.CreateSession(ctx, book.BookID, "main", 1)
	requireCode(t, err, CodeAlreadyExists)
	updated, err := st.RenameSession(ctx, book.BookID, session.SessionID, "Primary")
	if err != nil {
		t.Fatal(err)
	}
	if updated.Name != "Primary" {
		t.Fatalf("session was not renamed: %+v", updated)
	}
}

func TestCardRevisionsAndAtomicBatchDelete(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	book := testBook(t, st)
	desc := "Initial"
	deck, err := st.CreateDeck(
		ctx,
		book.BookID,
		"chapter-1",
		"Chapter 1",
		&desc,
		[]NewCard{{Front: "Question?", Back: "Answer", SourcePages: "2,1,2"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	card := deck.Cards[0]
	if card.SourcePages != "1,2" || card.Revision != 1 {
		t.Fatalf("card was not canonicalized: %+v", card)
	}

	// Add a revision and reject stale updates.
	newBack := "Better answer"
	updated, err := st.UpdateCard(ctx, book.BookID, deck.DeckID, card.CardID, 1, nil, &newBack, nil)
	if err != nil {
		t.Fatal(err)
	}
	if updated.Revision != 2 {
		t.Fatalf("expected revision 2, got %+v", updated)
	}
	_, err = st.UpdateCard(ctx, book.BookID, deck.DeckID, card.CardID, 1, nil, &newBack, nil)
	requireCode(t, err, CodeCardRevisionConflict)
	history, err := st.CardRevisions(ctx, book.BookID, deck.DeckID, card.CardID)
	if err != nil || len(history) != 2 || history[0].Back != "Answer" {
		t.Fatalf("immutable history missing: %+v %v", history, err)
	}

	// Reject the entire batch when one revision is stale.
	second, err := st.AddCards(
		ctx,
		book.BookID,
		deck.DeckID,
		[]NewCard{{Front: "Second?", Back: "Second", SourcePages: "3"}},
	)
	if err != nil {
		t.Fatal(err)
	}
	_, err = st.DeleteCards(
		ctx,
		book.BookID,
		deck.DeckID,
		[]CardDelete{
			{CardID: card.CardID, ExpectedRevision: 1},
			{CardID: second[0].CardID, ExpectedRevision: 1},
		},
	)
	requireCode(t, err, CodeCardRevisionConflict)
	read, err := st.ReadDeck(ctx, book.BookID, deck.DeckID)
	if err != nil || len(read.Cards) != 2 {
		t.Fatalf("batch delete was not atomic: %+v %v", read, err)
	}
	ids, err := st.DeleteCards(
		ctx,
		book.BookID,
		deck.DeckID,
		[]CardDelete{
			{CardID: card.CardID, ExpectedRevision: 2},
			{CardID: second[0].CardID, ExpectedRevision: 1},
		},
	)
	if err != nil || len(ids) != 2 {
		t.Fatalf("valid batch delete failed: %v %v", ids, err)
	}
}

func TestDeckMetadataRevisionAndExplicitDescriptionClear(t *testing.T) {
	ctx := context.Background()
	st := testStore(t)
	book := testBook(t, st)
	desc := "Description"
	deck, err := st.CreateDeck(ctx, book.BookID, "deck", "Deck", &desc, nil)
	if err != nil {
		t.Fatal(err)
	}
	renamed := "Renamed"
	summary, err := st.UpdateDeckWithChanges(ctx, book.BookID, deck.DeckID, 1, &renamed, nil, true)
	if err != nil {
		t.Fatal(err)
	}
	if summary.Revision != 2 || summary.Description != nil {
		t.Fatalf("unexpected update: %+v", summary)
	}
	err = st.DeleteDeck(ctx, book.BookID, deck.DeckID, 1)
	requireCode(t, err, CodeDeckRevisionConflict)
	if err = st.DeleteDeck(ctx, book.BookID, deck.DeckID, 2); err != nil {
		t.Fatal(err)
	}
}
