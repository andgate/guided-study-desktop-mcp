package store

// Book data.
type BookSummary struct {
	BookID    string `json:"book_id"`
	Title     string `json:"title"`
	PageCount int    `json:"page_count"`
}

type OutlineEntry struct {
	BookID       string `json:"-"`
	OutlineIndex int    `json:"outline_index"`
	Title        string `json:"title"`
	PageIndex    int    `json:"page_index"`
}

type Book struct {
	BookSummary
	Outline []OutlineEntry `json:"outline"`
}

// Rendered pages need no session.
type RenderedPage struct {
	PageIndex int    `json:"page_index"`
	MIMEType  string `json:"-"`
	ImageData []byte `json:"-"`
}

type RenderedPages struct {
	BookID         string         `json:"book_id"`
	PageCount      int            `json:"page_count"`
	StartPageIndex int            `json:"start_page_index"`
	EndPageIndex   int            `json:"end_page_index"`
	Pages          []RenderedPage `json:"pages"`
}

// Reading session data.
type SessionSummary struct {
	BookID              string  `json:"book_id"`
	SessionID           string  `json:"session_id"`
	Name                string  `json:"name"`
	OriginPageIndex     int     `json:"origin_page_index"`
	CheckpointPageIndex int     `json:"checkpoint_page_index"`
	CheckpointHeading   *string `json:"checkpoint_heading"`
}

type Checkpoint struct {
	PageIndex int    `json:"page_index"`
	Heading   string `json:"heading"`
}

type PageRange struct {
	StartPageIndex int `json:"start_page_index"`
	EndPageIndex   int `json:"end_page_index"`
}

type ReadingBatch struct {
	BookID    string         `json:"book_id"`
	SessionID string         `json:"session_id"`
	Batch     PageRange      `json:"batch"`
	Pages     []RenderedPage `json:"-"`
}

type PageSelection struct {
	BookID    string         `json:"book_id"`
	SessionID string         `json:"session_id"`
	Batch     PageRange      `json:"batch"`
	Pages     []RenderedPage `json:"-"`
}

type CreatedSession struct {
	Session   SessionSummary `json:"session"`
	Selection PageSelection  `json:"selection"`
}

// Flashcard data.
type DeckSummary struct {
	BookID      string  `json:"book_id"`
	DeckID      string  `json:"deck_id"`
	Title       string  `json:"title"`
	Description *string `json:"description"`
	Revision    int     `json:"revision"`
	CardCount   int     `json:"card_count"`
}

type NewCard struct {
	Front       string `json:"front"`
	Back        string `json:"back"`
	SourcePages string `json:"source_pages"`
}

type CardRevision struct {
	CardID      string `json:"card_id"`
	DeckID      string `json:"deck_id"`
	Front       string `json:"front"`
	Back        string `json:"back"`
	SourcePages string `json:"source_pages"`
	Revision    int    `json:"revision"`
}

type Deck struct {
	BookID      string         `json:"book_id"`
	DeckID      string         `json:"deck_id"`
	Title       string         `json:"title"`
	Description *string        `json:"description"`
	Revision    int            `json:"revision"`
	Cards       []CardRevision `json:"cards"`
}
