package mcpserver

import (
	"context"
	"fmt"

	"github.com/andgate/guided-study-desktop-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type importBookInput struct {
	FileReference string `json:"file_reference" jsonschema:"Absolute path to the local PDF supplied by the host."`
	Title         string `json:"title"          jsonschema:"Required display title chosen by the caller."`
}

type bookIDInput struct {
	BookID string `json:"book_id"`
}

type renameBookInput struct {
	BookID   string `json:"book_id"`
	NewTitle string `json:"new_title"`
}

type booksOutput struct {
	Books []store.BookSummary `json:"books"`
}

func (s *Service) registerBookTools() {
	// Register book import and listing tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "import_book",
			Title:       "Import book",
			Description: "Prepare a PDF into rendered page images and an ordered table of contents, then store the complete book transactionally.",
			Annotations: toolAnnotations("Import book", false, false),
		},
		func(ctx context.Context, in importBookInput) (store.BookSummary, string, error) {
			prepared, err := s.importer.Prepare(ctx, in.FileReference, in.Title)
			if err != nil {
				return store.BookSummary{}, "", err
			}
			book, err := s.store.ImportPreparedBook(ctx, prepared)
			return book, fmt.Sprintf("Imported %q with %d pages.", book.Title, book.PageCount), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "list_books",
			Title:       "List books",
			Description: "List prepared books so the user can choose one by stable ID.",
			Annotations: toolAnnotations("List books", true, false),
		},
		func(ctx context.Context, _ emptyInput) (booksOutput, string, error) {
			books, err := s.store.ListBooks(ctx)
			return booksOutput{Books: books}, fmt.Sprintf("Found %d books.", len(books)), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "get_book",
			Title:       "Get book",
			Description: "Read book metadata and its ordered table of contents without page content.",
			Annotations: toolAnnotations("Get book", true, false),
		},
		func(ctx context.Context, in bookIDInput) (store.Book, string, error) {
			book, err := s.store.GetBook(ctx, in.BookID)
			return book, fmt.Sprintf("Loaded metadata for %q.", book.Title), err
		},
	)

	// Register book update and deletion tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "rename_book",
			Title:       "Rename book",
			Description: "Change a book's display title without changing its stable ID.",
			Annotations: toolAnnotations("Rename book", false, false),
		},
		func(ctx context.Context, in renameBookInput) (store.BookSummary, string, error) {
			book, err := s.store.RenameBook(ctx, in.BookID, in.NewTitle)
			return book, fmt.Sprintf("Renamed the book to %q.", book.Title), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "remove_book",
			Title:       "Remove book",
			Description: "Permanently delete the exact book and all sessions, logs, decks, cards, and rendered pages it owns. The source PDF is never touched.",
			Annotations: toolAnnotations("Remove book", false, true),
		},
		func(ctx context.Context, in bookIDInput) (deletedBook, string, error) {
			err := s.store.RemoveBook(ctx, in.BookID)
			return deletedBook{
				BookID:  in.BookID,
				Deleted: err == nil,
			}, "Deleted the book and its canonical study data.", err
		},
	)
}
