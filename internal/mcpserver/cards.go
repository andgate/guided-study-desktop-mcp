package mcpserver

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"github.com/andgate/guided-study-desktop-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type createDeckInput struct {
	BookID      string          `json:"book_id"`
	DeckID      string          `json:"deck_id"`
	Title       string          `json:"title"`
	Description *string         `json:"description,omitempty"`
	Cards       []store.NewCard `json:"cards,omitempty"`
}

type deckInput struct {
	BookID string `json:"book_id"`
	DeckID string `json:"deck_id"`
}

type updateDeckInput struct {
	BookID           string          `json:"book_id"`
	DeckID           string          `json:"deck_id"`
	ExpectedRevision int             `json:"expected_revision"`
	Changes          json.RawMessage `json:"changes"`
}

type deleteDeckInput struct {
	BookID           string `json:"book_id"`
	DeckID           string `json:"deck_id"`
	ExpectedRevision int    `json:"expected_revision"`
}

// Card mutation inputs include the owning book and deck.
type cardInput struct {
	BookID string        `json:"book_id"`
	DeckID string        `json:"deck_id"`
	Card   store.NewCard `json:"card"`
}

type updateCardChanges struct {
	Front       *string `json:"front,omitempty"`
	Back        *string `json:"back,omitempty"`
	SourcePages *string `json:"source_pages,omitempty"`
}

type updateCardInput struct {
	BookID           string            `json:"book_id"`
	DeckID           string            `json:"deck_id"`
	CardID           string            `json:"card_id"`
	ExpectedRevision int               `json:"expected_revision"`
	Changes          updateCardChanges `json:"changes"`
}

type cardRevisionInput struct {
	BookID   string `json:"book_id"`
	DeckID   string `json:"deck_id"`
	CardID   string `json:"card_id"`
	Revision int    `json:"revision"`
}

type cardIDInput struct {
	BookID string `json:"book_id"`
	DeckID string `json:"deck_id"`
	CardID string `json:"card_id"`
}

type deleteCardInput struct {
	BookID           string `json:"book_id"`
	DeckID           string `json:"deck_id"`
	CardID           string `json:"card_id"`
	ExpectedRevision int    `json:"expected_revision"`
}

type addCardsInput struct {
	BookID string          `json:"book_id"`
	DeckID string          `json:"deck_id"`
	Cards  []store.NewCard `json:"cards"`
}

type deleteCardsInput struct {
	BookID string             `json:"book_id"`
	DeckID string             `json:"deck_id"`
	Cards  []store.CardDelete `json:"cards"`
}

// Collection wrappers name arrays in generated schemas.
type decksOutput struct {
	Decks []store.DeckSummary `json:"decks"`
}

type cardsOutput struct {
	Cards []store.CardRevision `json:"cards"`
}

type revisionsOutput struct {
	Revisions []store.CardRevision `json:"revisions"`
}

type deletedCardsOutput struct {
	DeletedCardIDs []string `json:"deleted_card_ids"`
}

// updateDeckSchema allows description to be omitted, set, or cleared.
func updateDeckSchema() map[string]any {
	return map[string]any{
		"type":                 "object",
		"additionalProperties": false,
		"required":             []string{"book_id", "deck_id", "expected_revision", "changes"},
		"properties": map[string]any{
			"book_id":           map[string]any{"type": "string"},
			"deck_id":           map[string]any{"type": "string"},
			"expected_revision": map[string]any{"type": "integer", "minimum": 1},
			"changes": map[string]any{
				"type":                 "object",
				"additionalProperties": false,
				"minProperties":        1,
				"properties": map[string]any{
					"title":       map[string]any{"type": "string"},
					"description": map[string]any{"type": []string{"string", "null"}},
				},
			},
		},
	}
}

func parseDeckChanges(
	raw json.RawMessage,
) (title *string, description *string, descriptionSet bool, err error) {
	// Preserve the difference between omitted and null fields.
	var fields map[string]json.RawMessage
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if err = decoder.Decode(&fields); err != nil {
		return
	}
	if len(fields) == 0 {
		err = &store.Error{
			Code:    store.CodeInvalidArgument,
			Message: "At least one metadata change is required.",
		}
		return
	}

	for key, value := range fields {
		switch key {
		case "title":
			var v string
			if e := json.Unmarshal(value, &v); e != nil {
				err = &store.Error{
					Code:    store.CodeInvalidArgument,
					Message: "changes.title must be a string.",
				}
				return
			}
			title = &v
		case "description":
			descriptionSet = true
			if bytes.Equal(bytes.TrimSpace(value), []byte("null")) {
				description = nil
			} else {
				var v string
				if e := json.Unmarshal(value, &v); e != nil {
					err = &store.Error{
						Code:    store.CodeInvalidArgument,
						Message: "changes.description must be a string or null.",
					}
					return
				}
				description = &v
			}
		default:
			err = &store.Error{
				Code:    store.CodeInvalidArgument,
				Message: fmt.Sprintf("Unrecognized deck metadata field %q.", key),
			}
			return
		}
	}
	return
}

func (s *Service) registerCardTools() {
	// Deck tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "list_decks",
			Title:       "List flashcard decks",
			Description: "List canonical flashcard decks for a book with metadata revision and current card count.",
			Annotations: toolAnnotations("List flashcard decks", true, false),
		},
		func(ctx context.Context, in bookIDInput) (decksOutput, string, error) {
			x, err := s.store.ListDecks(ctx, in.BookID)
			return decksOutput{Decks: x}, fmt.Sprintf("Found %d decks.", len(x)), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "create_deck",
			Title:       "Create flashcard deck",
			Description: "Create a canonical deck and optional initial cards atomically. The caller supplies the stable deck ID; card IDs are generated.",
			Annotations: toolAnnotations("Create flashcard deck", false, false),
		},
		func(ctx context.Context, in createDeckInput) (store.Deck, string, error) {
			x, err := s.store.CreateDeck(
				ctx,
				in.BookID,
				in.DeckID,
				in.Title,
				in.Description,
				in.Cards,
			)
			return x, fmt.Sprintf("Created deck %q with %d cards.", x.DeckID, len(x.Cards)), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "read_deck",
			Title:       "Read flashcard deck",
			Description: "Read deck metadata and only the latest immutable revision of every card, sorted by card ID.",
			Annotations: toolAnnotations("Read flashcard deck", true, false),
		},
		func(ctx context.Context, in deckInput) (store.Deck, string, error) {
			x, err := s.store.ReadDeck(ctx, in.BookID, in.DeckID)
			return x, fmt.Sprintf("Loaded deck %q with %d cards.", x.DeckID, len(x.Cards)), err
		},
	)

	updateTool := &mcp.Tool{
		Name:        "update_deck",
		Title:       "Update flashcard deck",
		Description: "Update title and/or description after validating the deck metadata revision. Card history is unaffected.",
		Annotations: toolAnnotations("Update flashcard deck", false, false),
		InputSchema: updateDeckSchema(),
	}
	addTool(
		s.server,
		updateTool,
		func(ctx context.Context, in updateDeckInput) (store.DeckSummary, string, error) {
			title, description, set, err := parseDeckChanges(in.Changes)
			if err != nil {
				return store.DeckSummary{}, "", err
			}
			x, err := s.store.UpdateDeckWithChanges(
				ctx,
				in.BookID,
				in.DeckID,
				in.ExpectedRevision,
				title,
				description,
				set,
			)
			return x, fmt.Sprintf(
				"Updated deck %q to metadata revision %d.",
				x.DeckID,
				x.Revision,
			), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "delete_deck",
			Title:       "Delete flashcard deck",
			Description: "Permanently delete the exact deck and every card revision after checking the deck metadata revision.",
			Annotations: toolAnnotations("Delete flashcard deck", false, true),
		},
		func(ctx context.Context, in deleteDeckInput) (deletedDeck, string, error) {
			err := s.store.DeleteDeck(ctx, in.BookID, in.DeckID, in.ExpectedRevision)
			return deletedDeck{
				BookID:  in.BookID,
				DeckID:  in.DeckID,
				Deleted: err == nil,
			}, "Deleted the deck and all of its card revisions.", err
		},
	)

	// Card tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "add_card",
			Title:       "Add flashcard",
			Description: "Add one card as immutable revision 1 with a service-generated card ID.",
			Annotations: toolAnnotations("Add flashcard", false, false),
		},
		func(ctx context.Context, in cardInput) (store.CardRevision, string, error) {
			x, err := s.store.AddCards(ctx, in.BookID, in.DeckID, []store.NewCard{in.Card})
			if err != nil || len(x) == 0 {
				return store.CardRevision{}, "", err
			}
			return x[0], fmt.Sprintf("Added card %s.", x[0].CardID), nil
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "update_card",
			Title:       "Update flashcard",
			Description: "Insert a new immutable card revision after checking the expected latest revision; unchanged fields are copied.",
			Annotations: toolAnnotations("Update flashcard", false, false),
		},
		func(ctx context.Context, in updateCardInput) (store.CardRevision, string, error) {
			x, err := s.store.UpdateCard(
				ctx,
				in.BookID,
				in.DeckID,
				in.CardID,
				in.ExpectedRevision,
				in.Changes.Front,
				in.Changes.Back,
				in.Changes.SourcePages,
			)
			return x, fmt.Sprintf("Created revision %d for card %s.", x.Revision, x.CardID), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "delete_card",
			Title:       "Delete flashcard",
			Description: "Permanently delete one logical card and all immutable revisions after checking its latest revision.",
			Annotations: toolAnnotations("Delete flashcard", false, true),
		},
		func(ctx context.Context, in deleteCardInput) (deletedCard, string, error) {
			_, err := s.store.DeleteCards(
				ctx,
				in.BookID,
				in.DeckID,
				[]store.CardDelete{{CardID: in.CardID, ExpectedRevision: in.ExpectedRevision}},
			)
			return deletedCard{
					CardID:  in.CardID,
					Deleted: err == nil,
				}, fmt.Sprintf(
					"Deleted card %s and all revisions.",
					in.CardID,
				), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "read_card_revision",
			Title:       "Read flashcard revision",
			Description: "Read one exact immutable revision of a logical card.",
			Annotations: toolAnnotations("Read flashcard revision", true, false),
		},
		func(ctx context.Context, in cardRevisionInput) (store.CardRevision, string, error) {
			x, err := s.store.ReadCardRevision(ctx, in.BookID, in.DeckID, in.CardID, in.Revision)
			return x, fmt.Sprintf("Loaded card %s revision %d.", x.CardID, x.Revision), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "list_card_revisions",
			Title:       "List flashcard revisions",
			Description: "Read the full immutable revision history for one logical card in ascending revision order.",
			Annotations: toolAnnotations("List flashcard revisions", true, false),
		},
		func(ctx context.Context, in cardIDInput) (revisionsOutput, string, error) {
			x, err := s.store.CardRevisions(ctx, in.BookID, in.DeckID, in.CardID)
			return revisionsOutput{
					Revisions: x,
				}, fmt.Sprintf(
					"Found %d revisions for card %s.",
					len(x),
					in.CardID,
				), err
		},
	)

	// Batch card tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "add_cards",
			Title:       "Add flashcards",
			Description: "Validate and add multiple cards atomically, each as immutable revision 1 with a generated ID.",
			Annotations: toolAnnotations("Add flashcards", false, false),
		},
		func(ctx context.Context, in addCardsInput) (cardsOutput, string, error) {
			x, err := s.store.AddCards(ctx, in.BookID, in.DeckID, in.Cards)
			return cardsOutput{Cards: x}, fmt.Sprintf("Added %d cards.", len(x)), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "delete_cards",
			Title:       "Delete flashcards",
			Description: "Validate every target and expected revision, then atomically delete all named logical cards and their histories.",
			Annotations: toolAnnotations("Delete flashcards", false, true),
		},
		func(ctx context.Context, in deleteCardsInput) (deletedCardsOutput, string, error) {
			x, err := s.store.DeleteCards(ctx, in.BookID, in.DeckID, in.Cards)
			return deletedCardsOutput{
					DeletedCardIDs: x,
				}, fmt.Sprintf(
					"Deleted %d cards and their revision histories.",
					len(x),
				), err
		},
	)
}
