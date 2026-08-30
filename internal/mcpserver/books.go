package mcpserver

import (
	"context"
	"encoding/csv"
	"fmt"
	"strconv"
	"strings"

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

type bookOutput struct {
	store.BookSummary
	OutlineCSV string `json:"outline_csv"`
}

var outlineHeader = []string{
	"outline_index",
	"title",
	"page_index",
}

func (s *Service) registerBookTools() {
	// Register book import and listing tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "import_book",
			Title:       "Import book",
			Description: "Import a PDF and extract its outline.",
			Annotations: toolAnnotations("Import book", false, false),
		},
		func(ctx context.Context, in importBookInput) (store.BookSummary, string, error) {
			book, err := s.importer.Import(ctx, in.FileReference, in.Title)
			return book, fmt.Sprintf("Imported %q with %d pages.", book.Title, book.PageCount), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "list_books",
			Title:       "List books",
			Description: "List books with stable IDs.",
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
			Description: "Read book metadata and its outline.",
			Annotations: toolAnnotations("Get book", true, false),
		},
		func(ctx context.Context, in bookIDInput) (bookOutput, string, error) {
			book, err := s.store.GetBook(ctx, in.BookID)
			if err != nil {
				return bookOutput{}, "", err
			}
			outline, err := outlineCSV(book.Outline)
			return bookOutput{
				BookSummary: book.BookSummary,
				OutlineCSV:  outline,
			}, fmt.Sprintf("Loaded metadata for %q.", book.Title), err
		},
	)

	// Register book update and deletion tools.
	addTool(
		s.server,
		&mcp.Tool{
			Name:        "rename_book",
			Title:       "Rename book",
			Description: "Rename a book.",
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
			Description: "Delete a book and its study data.",
			Annotations: toolAnnotations("Remove book", false, true),
		},
		func(ctx context.Context, in bookIDInput) (deletedBook, string, error) {
			err := s.store.RemoveBook(ctx, in.BookID)
			return deletedBook{
				BookID:  in.BookID,
				Deleted: err == nil,
			}, "Deleted the book and its study data.", err
		},
	)
}

func outlineCSV(outline []store.OutlineEntry) (string, error) {
	var output strings.Builder
	writer := csv.NewWriter(&output)
	writer.UseCRLF = true

	if err := writer.Write(outlineHeader); err != nil {
		return "", err
	}
	for _, entry := range outline {
		row := []string{
			strconv.Itoa(entry.OutlineIndex),
			entry.Title,
			strconv.Itoa(entry.PageIndex),
		}
		if err := writer.Write(row); err != nil {
			return "", err
		}
	}
	writer.Flush()
	return output.String(), writer.Error()
}
