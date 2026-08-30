package mcpserver

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"

	"github.com/andgate/guided-study-desktop-mcp/internal/importer"
	"github.com/andgate/guided-study-desktop-mcp/internal/store"
	"github.com/google/jsonschema-go/jsonschema"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

type Service struct {
	store    *store.Store
	importer *importer.Importer
	server   *mcp.Server
}

// New creates the MCP server and registers its tools.
func New(st *store.Store, imp *importer.Importer, logger *slog.Logger) *Service {
	svc := &Service{store: st, importer: imp}
	svc.server = mcp.NewServer(
		&mcp.Implementation{Name: "guided-study", Version: "1.0.0"},
		&mcp.ServerOptions{
			Logger: logger,
		},
	)

	// Register tools by domain.
	svc.registerBookTools()
	svc.registerReadingTools()
	svc.registerCardTools()
	return svc
}

func (s *Service) Handler(logger *slog.Logger) http.Handler {
	return mcp.NewStreamableHTTPHandler(
		func(*http.Request) *mcp.Server { return s.server },
		&mcp.StreamableHTTPOptions{Logger: logger},
	)
}

type handler[In, Out any] func(context.Context, In) (Out, string, error)

// addTool registers a typed handler with structured and text responses.
func addTool[In, Out any](server *mcp.Server, tool *mcp.Tool, h handler[In, Out]) {
	schema, err := jsonschema.For[Out](nil)
	if err != nil {
		panic(fmt.Sprintf("infer output schema for %s: %v", tool.Name, err))
	}
	tool.OutputSchema = schema

	mcp.AddTool[In, any](
		server,
		tool,
		func(ctx context.Context, _ *mcp.CallToolRequest, in In) (*mcp.CallToolResult, any, error) {
			out, fallback, err := h(ctx, in)
			if err != nil {
				return toolErrorResult(err), nil, nil
			}
			return &mcp.CallToolResult{
				Content: []mcp.Content{&mcp.TextContent{Text: fallback}},
			}, out, nil
		},
	)
}

// toolErrorResult creates an MCP error response.
func toolErrorResult(err error) *mcp.CallToolResult {
	var serviceErr *store.Error
	if !errors.As(err, &serviceErr) {
		serviceErr = &store.Error{
			Code:    store.CodeStorageError,
			Message: "The operation failed unexpectedly.",
			Cause:   err,
		}
	}

	return &mcp.CallToolResult{
		IsError:           true,
		StructuredContent: serviceErr,
		Content:           []mcp.Content{&mcp.TextContent{Text: serviceErr.Message}},
	}
}

// toolAnnotations sets the common tool flags.
func toolAnnotations(title string, readOnly, destructive bool) *mcp.ToolAnnotations {
	closed := false
	return &mcp.ToolAnnotations{
		Title:           title,
		ReadOnlyHint:    readOnly,
		DestructiveHint: &destructive,
		OpenWorldHint:   &closed,
	}
}

// Inputs shared by several tools.
type emptyInput struct{}

type deletedBook struct {
	BookID  string `json:"book_id"`
	Deleted bool   `json:"deleted"`
}

type deletedSession struct {
	BookID    string `json:"book_id"`
	SessionID string `json:"session_id"`
	Deleted   bool   `json:"deleted"`
}

type deletedDeck struct {
	BookID  string `json:"book_id"`
	DeckID  string `json:"deck_id"`
	Deleted bool   `json:"deleted"`
}

type deletedCard struct {
	CardID  string `json:"card_id"`
	Deleted bool   `json:"deleted"`
}
