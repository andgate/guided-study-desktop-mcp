package store

import (
	"context"
	"database/sql"
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/google/uuid"
)

var deckIDPattern = regexp.MustCompile(`^[a-z0-9][a-z0-9_-]{0,63}$`)

// normalizeCard validates a card and sorts its source page numbers.
func (s *Store) normalizeCard(ctx context.Context, q interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}, bookID string, card NewCard) (NewCard, error) {
	front, e := cleanRequired("front", card.Front)
	if e != nil {
		return card, e
	}
	back, e := cleanRequired("back", card.Back)
	if e != nil {
		return card, e
	}
	raw, e := cleanRequired("source_pages", card.SourcePages)
	if e != nil {
		return card, e
	}
	var count int
	if err := q.QueryRowContext(ctx, `SELECT page_count FROM books WHERE book_id=?`, bookID).
		Scan(&count); err == sql.ErrNoRows {
		return card, errf(
			CodeNotFound,
			map[string]any{"book_id": bookID},
			"Book %q was not found.",
			bookID,
		)
	} else if err != nil {
		return card, storageError(err)
	}

	seen := map[int]bool{}
	pages := []int{}
	for _, part := range strings.Split(raw, ",") {
		if part == "" || strings.TrimSpace(part) != part {
			return card, errf(
				CodeInvalidArgument,
				map[string]any{"source_pages": raw},
				"source_pages must be comma-separated integers without spaces.",
			)
		}
		n, err := strconv.Atoi(part)
		if err != nil || n < 1 || n > count {
			return card, errf(
				CodeInvalidArgument,
				map[string]any{"source_pages": raw, "page_count": count},
				"source_pages contains an invalid page index.",
			)
		}
		if !seen[n] {
			seen[n] = true
			pages = append(pages, n)
		}
	}
	sort.Ints(pages)
	parts := make([]string, len(pages))
	for i, n := range pages {
		parts[i] = strconv.Itoa(n)
	}
	return NewCard{Front: front, Back: back, SourcePages: strings.Join(parts, ",")}, nil
}

func (s *Store) ListDecks(ctx context.Context, bookID string) ([]DeckSummary, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT d.book_id,d.deck_id,d.title,d.description,d.revision,count(c.card_id) FROM decks d LEFT JOIN current_cards c ON c.book_id=d.book_id AND c.deck_id=d.deck_id WHERE d.book_id=? GROUP BY d.book_id,d.deck_id,d.title,d.description,d.revision ORDER BY d.deck_id`,
		bookID,
	)
	if err != nil {
		return nil, storageError(err)
	}
	defer rows.Close()
	out := []DeckSummary{}
	for rows.Next() {
		var x DeckSummary
		var desc sql.NullString
		if err := rows.Scan(
			&x.BookID,
			&x.DeckID,
			&x.Title,
			&desc,
			&x.Revision,
			&x.CardCount,
		); err != nil {
			return nil, storageError(err)
		}
		x.Description = scanNullableString(desc)
		out = append(out, x)
	}
	if len(out) == 0 {
		// Check whether the book exists when it has no decks.
		var n int
		err = s.db.QueryRowContext(ctx, `SELECT 1 FROM books WHERE book_id=?`, bookID).Scan(&n)
		if err == sql.ErrNoRows {
			return nil, errf(
				CodeNotFound,
				map[string]any{"book_id": bookID},
				"Book %q was not found.",
				bookID,
			)
		}
	}
	return out, rows.Err()
}

func (s *Store) CreateDeck(
	ctx context.Context,
	bookID, deckID, title string,
	description *string,
	cards []NewCard,
) (Deck, error) {
	// Validate the deck and its cards before writing.
	if !deckIDPattern.MatchString(deckID) {
		return Deck{}, errf(
			CodeInvalidArgument,
			map[string]any{"deck_id": deckID},
			"deck_id must match [a-z0-9][a-z0-9_-]{0,63}.",
		)
	}
	title, e := cleanRequired("title", title)
	if e != nil {
		return Deck{}, e
	}
	description = cleanOptional(description)
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return Deck{}, storageError(err)
	}
	defer tx.Rollback()

	normalized := make([]NewCard, len(cards))
	for i, c := range cards {
		normalized[i], err = s.normalizeCard(ctx, tx, bookID, c)
		if err != nil {
			return Deck{}, err
		}
	}
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO decks(book_id,deck_id,title,description,revision) VALUES(?,?,?,?,1)`,
		bookID,
		deckID,
		title,
		description,
	)
	if isConstraint(err) {
		var exists int
		if x := tx.QueryRowContext(ctx, `SELECT 1 FROM books WHERE book_id=?`, bookID).
			Scan(&exists); x == sql.ErrNoRows {
			return Deck{}, errf(
				CodeNotFound,
				map[string]any{"book_id": bookID},
				"Book %q was not found.",
				bookID,
			)
		}
		return Deck{}, errf(
			CodeAlreadyExists,
			map[string]any{"book_id": bookID, "deck_id": deckID},
			"Deck %q already exists for this book.",
			deckID,
		)
	}
	if err != nil {
		return Deck{}, storageError(err)
	}

	// Create the deck and its cards in one transaction.
	out := Deck{
		BookID:      bookID,
		DeckID:      deckID,
		Title:       title,
		Description: description,
		Revision:    1,
		Cards:       []CardRevision{},
	}
	for _, c := range normalized {
		id := uuid.NewString()
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO card_revisions(book_id,deck_id,card_id,revision,front,back,source_pages) VALUES(?,?,?,1,?,?,?)`,
			bookID,
			deckID,
			id,
			c.Front,
			c.Back,
			c.SourcePages,
		); err != nil {
			return Deck{}, storageError(err)
		}
		out.Cards = append(
			out.Cards,
			CardRevision{
				CardID:      id,
				DeckID:      deckID,
				Front:       c.Front,
				Back:        c.Back,
				SourcePages: c.SourcePages,
				Revision:    1,
			},
		)
	}
	sort.Slice(out.Cards, func(i, j int) bool { return out.Cards[i].CardID < out.Cards[j].CardID })
	if err = tx.Commit(); err != nil {
		return Deck{}, storageError(err)
	}

	return out, nil
}

func (s *Store) ReadDeck(ctx context.Context, bookID, deckID string) (Deck, error) {
	// Read the deck separately so an empty deck is not mistaken for a missing one.
	var d Deck
	var desc sql.NullString
	err := s.db.QueryRowContext(ctx, `SELECT book_id,deck_id,title,description,revision FROM decks WHERE book_id=? AND deck_id=?`, bookID, deckID).
		Scan(&d.BookID, &d.DeckID, &d.Title, &desc, &d.Revision)
	if err == sql.ErrNoRows {
		return d, errf(
			CodeNotFound,
			map[string]any{"book_id": bookID, "deck_id": deckID},
			"Deck %q was not found for this book.",
			deckID,
		)
	}
	if err != nil {
		return d, storageError(err)
	}

	d.Description = scanNullableString(desc)
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT card_id,deck_id,front,back,source_pages,revision FROM current_cards WHERE book_id=? AND deck_id=? ORDER BY card_id`,
		bookID,
		deckID,
	)
	if err != nil {
		return d, storageError(err)
	}
	defer rows.Close()
	d.Cards = []CardRevision{}
	for rows.Next() {
		var c CardRevision
		if err := rows.Scan(
			&c.CardID,
			&c.DeckID,
			&c.Front,
			&c.Back,
			&c.SourcePages,
			&c.Revision,
		); err != nil {
			return d, storageError(err)
		}
		d.Cards = append(d.Cards, c)
	}
	return d, rows.Err()
}

func (s *Store) UpdateDeckWithChanges(
	ctx context.Context,
	bookID, deckID string,
	expected int,
	title *string,
	description *string,
	descriptionSet bool,
) (DeckSummary, error) {
	// descriptionSet is true when the request includes the description field.
	if title == nil && !descriptionSet {
		return DeckSummary{}, errf(
			CodeInvalidArgument,
			nil,
			"At least one metadata change is required.",
		)
	}
	var cleanTitle *string
	if title != nil {
		v, e := cleanRequired("title", *title)
		if e != nil {
			return DeckSummary{}, e
		}
		cleanTitle = &v
	}
	if descriptionSet {
		description = cleanOptional(description)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return DeckSummary{}, storageError(err)
	}
	defer tx.Rollback()
	var actual int
	var ct string
	var cd sql.NullString
	err = tx.QueryRowContext(ctx, `SELECT title,description,revision FROM decks WHERE book_id=? AND deck_id=?`, bookID, deckID).
		Scan(&ct, &cd, &actual)
	if err == sql.ErrNoRows {
		return DeckSummary{}, errf(CodeNotFound, nil, "Deck was not found for this book.")
	}
	if err != nil {
		return DeckSummary{}, storageError(err)
	}
	if actual != expected {
		return DeckSummary{}, errf(
			CodeDeckRevisionConflict,
			map[string]any{"expected_revision": expected, "actual_revision": actual},
			"Expected deck revision %d, but the latest revision is %d.",
			expected,
			actual,
		)
	}

	// Apply the changes and increment the revision.
	if cleanTitle != nil {
		ct = *cleanTitle
	}
	var dv any = scanNullableString(cd)
	if descriptionSet {
		dv = description
	}
	_, err = tx.ExecContext(
		ctx,
		`UPDATE decks SET title=?,description=?,revision=revision+1 WHERE book_id=? AND deck_id=?`,
		ct,
		dv,
		bookID,
		deckID,
	)
	if err != nil {
		return DeckSummary{}, storageError(err)
	}
	if err = tx.Commit(); err != nil {
		return DeckSummary{}, storageError(err)
	}

	// Return the updated deck summary.
	all, err := s.ListDecks(ctx, bookID)
	if err != nil {
		return DeckSummary{}, err
	}
	for _, x := range all {
		if x.DeckID == deckID {
			return x, nil
		}
	}
	return DeckSummary{}, errf(CodeNotFound, nil, "Deck was not found for this book.")
}

func (s *Store) DeleteDeck(ctx context.Context, bookID, deckID string, expected int) error {
	// Check the revision and delete the deck in one transaction.
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return storageError(err)
	}
	defer tx.Rollback()
	var actual int
	err = tx.QueryRowContext(ctx, `SELECT revision FROM decks WHERE book_id=? AND deck_id=?`, bookID, deckID).
		Scan(&actual)
	if err == sql.ErrNoRows {
		return errf(CodeNotFound, nil, "Deck was not found for this book.")
	}
	if err != nil {
		return storageError(err)
	}
	if actual != expected {
		return errf(
			CodeDeckRevisionConflict,
			map[string]any{"expected_revision": expected, "actual_revision": actual},
			"Expected deck revision %d, but the latest revision is %d.",
			expected,
			actual,
		)
	}
	if _, err = tx.ExecContext(
		ctx,
		`DELETE FROM decks WHERE book_id=? AND deck_id=?`,
		bookID,
		deckID,
	); err != nil {
		return storageError(err)
	}
	if err = tx.Commit(); err != nil {
		return storageError(err)
	}
	return nil
}

func (s *Store) AddCards(
	ctx context.Context,
	bookID, deckID string,
	cards []NewCard,
) ([]CardRevision, error) {
	if len(cards) == 0 {
		return nil, errf(CodeInvalidArgument, nil, "At least one card is required.")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, storageError(err)
	}
	defer tx.Rollback()
	var exists int
	if err = tx.QueryRowContext(ctx, `SELECT 1 FROM decks WHERE book_id=? AND deck_id=?`, bookID, deckID).
		Scan(&exists); err == sql.ErrNoRows {
		return nil, errf(CodeNotFound, nil, "Deck was not found for this book.")
	} else if err != nil {
		return nil, storageError(err)
	}

	// Validate every card before writing the batch.
	out := make([]CardRevision, len(cards))
	for i, c := range cards {
		n, e := s.normalizeCard(ctx, tx, bookID, c)
		if e != nil {
			return nil, e
		}
		id := uuid.NewString()
		if _, err = tx.ExecContext(
			ctx,
			`INSERT INTO card_revisions(book_id,deck_id,card_id,revision,front,back,source_pages) VALUES(?,?,?,1,?,?,?)`,
			bookID,
			deckID,
			id,
			n.Front,
			n.Back,
			n.SourcePages,
		); err != nil {
			return nil, storageError(err)
		}
		out[i] = CardRevision{
			CardID:      id,
			DeckID:      deckID,
			Front:       n.Front,
			Back:        n.Back,
			SourcePages: n.SourcePages,
			Revision:    1,
		}
	}
	sort.Slice(out, func(i, j int) bool { return out[i].CardID < out[j].CardID })
	if err = tx.Commit(); err != nil {
		return nil, storageError(err)
	}
	return out, nil
}

func (s *Store) UpdateCard(
	ctx context.Context,
	bookID, deckID, cardID string,
	expected int,
	front, back, sourcePages *string,
) (CardRevision, error) {
	if front == nil && back == nil && sourcePages == nil {
		return CardRevision{}, errf(
			CodeInvalidArgument,
			nil,
			"At least one card change is required.",
		)
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return CardRevision{}, storageError(err)
	}
	defer tx.Rollback()
	var current CardRevision
	err = tx.QueryRowContext(ctx, `SELECT card_id,deck_id,front,back,source_pages,revision FROM current_cards WHERE book_id=? AND deck_id=? AND card_id=?`, bookID, deckID, cardID).
		Scan(&current.CardID, &current.DeckID, &current.Front, &current.Back, &current.SourcePages, &current.Revision)
	if err == sql.ErrNoRows {
		return current, errf(CodeNotFound, nil, "Card was not found in this deck.")
	}
	if err != nil {
		return current, storageError(err)
	}
	if current.Revision != expected {
		return current, errf(
			CodeCardRevisionConflict,
			map[string]any{"expected_revision": expected, "actual_revision": current.Revision},
			"Expected card revision %d, but the latest revision is %d.",
			expected,
			current.Revision,
		)
	}

	// Apply changes to the current card revision.
	candidate := NewCard{Front: current.Front, Back: current.Back, SourcePages: current.SourcePages}
	if front != nil {
		candidate.Front = *front
	}
	if back != nil {
		candidate.Back = *back
	}
	if sourcePages != nil {
		candidate.SourcePages = *sourcePages
	}
	candidate, e := s.normalizeCard(ctx, tx, bookID, candidate)
	if e != nil {
		return current, e
	}

	// Save the update as a new revision.
	current.Revision++
	current.Front = candidate.Front
	current.Back = candidate.Back
	current.SourcePages = candidate.SourcePages
	_, err = tx.ExecContext(
		ctx,
		`INSERT INTO card_revisions(book_id,deck_id,card_id,revision,front,back,source_pages) VALUES(?,?,?,?,?,?,?)`,
		bookID,
		deckID,
		cardID,
		current.Revision,
		current.Front,
		current.Back,
		current.SourcePages,
	)
	if err != nil {
		return current, storageError(err)
	}
	if err = tx.Commit(); err != nil {
		return current, storageError(err)
	}
	return current, nil
}

func (s *Store) CardRevisions(
	ctx context.Context,
	bookID, deckID, cardID string,
) ([]CardRevision, error) {
	rows, err := s.db.QueryContext(
		ctx,
		`SELECT card_id,deck_id,front,back,source_pages,revision FROM card_revisions WHERE book_id=? AND deck_id=? AND card_id=? ORDER BY revision`,
		bookID,
		deckID,
		cardID,
	)
	if err != nil {
		return nil, storageError(err)
	}
	defer rows.Close()
	out := []CardRevision{}
	for rows.Next() {
		var c CardRevision
		if err := rows.Scan(
			&c.CardID,
			&c.DeckID,
			&c.Front,
			&c.Back,
			&c.SourcePages,
			&c.Revision,
		); err != nil {
			return nil, storageError(err)
		}
		out = append(out, c)
	}
	if len(out) == 0 {
		return nil, errf(CodeNotFound, nil, "Card was not found in this deck.")
	}
	return out, rows.Err()
}

func (s *Store) ReadCardRevision(
	ctx context.Context,
	bookID, deckID, cardID string,
	revision int,
) (CardRevision, error) {
	var c CardRevision
	err := s.db.QueryRowContext(ctx, `SELECT card_id,deck_id,front,back,source_pages,revision FROM card_revisions WHERE book_id=? AND deck_id=? AND card_id=? AND revision=?`, bookID, deckID, cardID, revision).
		Scan(&c.CardID, &c.DeckID, &c.Front, &c.Back, &c.SourcePages, &c.Revision)
	if err == sql.ErrNoRows {
		return c, errf(CodeNotFound, nil, "Card revision was not found.")
	}
	if err != nil {
		return c, storageError(err)
	}
	return c, nil
}

type CardDelete struct {
	CardID           string `json:"card_id"`
	ExpectedRevision int    `json:"expected_revision"`
}

func (s *Store) DeleteCards(
	ctx context.Context,
	bookID, deckID string,
	cards []CardDelete,
) ([]string, error) {
	if len(cards) == 0 {
		return nil, errf(CodeInvalidArgument, nil, "At least one card is required.")
	}

	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, storageError(err)
	}
	defer tx.Rollback()

	// Validate the entire batch before deleting any cards.
	seen := map[string]bool{}
	ids := make([]string, len(cards))
	for i, target := range cards {
		if seen[target.CardID] {
			return nil, errf(CodeInvalidArgument, nil, "Duplicate card_id in delete request.")
		}
		seen[target.CardID] = true
		var actual int
		err = tx.QueryRowContext(ctx, `SELECT revision FROM current_cards WHERE book_id=? AND deck_id=? AND card_id=?`, bookID, deckID, target.CardID).
			Scan(&actual)
		if err == sql.ErrNoRows {
			return nil, errf(
				CodeNotFound,
				map[string]any{"card_id": target.CardID},
				"Card was not found in this deck.",
			)
		}
		if err != nil {
			return nil, storageError(err)
		}
		if actual != target.ExpectedRevision {
			return nil, errf(
				CodeCardRevisionConflict,
				map[string]any{
					"card_id":           target.CardID,
					"expected_revision": target.ExpectedRevision,
					"actual_revision":   actual,
				},
				"Expected card revision %d, but the latest revision is %d.",
				target.ExpectedRevision,
				actual,
			)
		}
		ids[i] = target.CardID
	}

	for _, id := range ids {
		if _, err = tx.ExecContext(
			ctx,
			`DELETE FROM card_revisions WHERE book_id=? AND deck_id=? AND card_id=?`,
			bookID,
			deckID,
			id,
		); err != nil {
			return nil, storageError(err)
		}
	}
	sort.Strings(ids)
	if err = tx.Commit(); err != nil {
		return nil, storageError(err)
	}
	return ids, nil
}
