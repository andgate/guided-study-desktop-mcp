package mcpserver

import (
	"context"
	"fmt"

	"github.com/andgate/guided-study-desktop-mcp/internal/store"
	"github.com/modelcontextprotocol/go-sdk/mcp"
)

const pageBlockCount = 2

type createSessionInput struct {
	BookID         string `json:"book_id"`
	Name           string `json:"name"`
	StartPageIndex int    `json:"start_page_index"`
}

type sessionInput struct {
	BookID    string `json:"book_id"`
	SessionID string `json:"session_id"`
}

type checkpointInput struct {
	BookID    string `json:"book_id"`
	SessionID string `json:"session_id"`
	PageIndex int    `json:"page_index"`
	Heading   string `json:"heading"`
}

func (in checkpointInput) checkpoint() store.Checkpoint {
	return store.Checkpoint{
		PageIndex: in.PageIndex,
		Heading:   in.Heading,
	}
}

type gotoPageInput struct {
	BookID    string `json:"book_id"`
	SessionID string `json:"session_id"`
	PageIndex int    `json:"page_index"`
}

type readPagesInput struct {
	BookID         string `json:"book_id"`
	StartPageIndex int    `json:"start_page_index"`
	EndPageIndex   int    `json:"end_page_index"`
}

type renameSessionInput struct {
	BookID    string `json:"book_id"`
	SessionID string `json:"session_id"`
	NewName   string `json:"new_name"`
}

type sessionsOutput struct {
	Sessions []store.SessionSummary `json:"sessions"`
}

type pageBatchEntry struct {
	PageIndex int `json:"page_index"`
}

type pageBatch struct {
	BookID         string           `json:"book_id"`
	PageCount      int              `json:"page_count"`
	StartPageIndex int              `json:"start_page_index"`
	EndPageIndex   int              `json:"end_page_index"`
	Pages          []pageBatchEntry `json:"pages"`
}

func batchMetadata(pages store.RenderedPages) pageBatch {
	entries := make([]pageBatchEntry, 0, len(pages.Pages))
	for _, page := range pages.Pages {
		entries = append(entries, pageBatchEntry{PageIndex: page.PageIndex})
	}
	return pageBatch{
		BookID:         pages.BookID,
		PageCount:      pages.PageCount,
		StartPageIndex: pages.StartPageIndex,
		EndPageIndex:   pages.EndPageIndex,
		Pages:          entries,
	}
}

func pageContent(pages []store.RenderedPage) []mcp.Content {
	content := make([]mcp.Content, 0, len(pages)*pageBlockCount)
	for _, page := range pages {
		content = append(
			content,
			&mcp.TextContent{Text: fmt.Sprintf("Page %d.", page.PageIndex)},
			&mcp.ImageContent{Data: page.ImageData, MIMEType: page.MIMEType},
		)
	}
	return content
}

type pageHandler[In, Out any] func(
	context.Context,
	In,
) (Out, []store.RenderedPage, string, error)

func addPageTool[In, Out any](
	server *mcp.Server,
	tool *mcp.Tool,
	h pageHandler[In, Out],
) {
	// Register the response schema.
	addTool(server, tool, func(ctx context.Context, in In) (Out, string, error) {
		out, _, fallback, err := h(ctx, in)
		return out, fallback, err
	})

	// Include the rendered images.
	mcp.AddTool[In, any](
		server,
		tool,
		func(
			ctx context.Context,
			_ *mcp.CallToolRequest,
			in In,
		) (*mcp.CallToolResult, any, error) {
			out, pages, _, err := h(ctx, in)
			if err != nil {
				return toolErrorResult(err), nil, nil
			}
			return &mcp.CallToolResult{Content: pageContent(pages)}, out, nil
		},
	)
}

func (s *Service) registerReadingTools() {
	addPageTool(
		s.server,
		&mcp.Tool{
			Name:        "create_session",
			Title:       "Create study session",
			Description: "Create a session at one page.",
			Annotations: toolAnnotations("Create study session", false, false),
		},
		func(ctx context.Context, in createSessionInput) (
			store.CreatedSession,
			[]store.RenderedPage,
			string,
			error,
		) {
			created, err := s.store.CreateSession(
				ctx,
				in.BookID,
				in.Name,
				in.StartPageIndex,
			)
			return created,
				created.Selection.Pages,
				fmt.Sprintf("Created study session %q.", created.Session.Name),
				err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "list_sessions",
			Title:       "List study sessions",
			Description: "List durable study progress.",
			Annotations: toolAnnotations("List study sessions", true, false),
		},
		func(ctx context.Context, in bookIDInput) (sessionsOutput, string, error) {
			sessions, err := s.store.ListSessions(ctx, in.BookID)
			return sessionsOutput{Sessions: sessions}, fmt.Sprintf("Found %d study sessions.", len(sessions)), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "rename_session",
			Title:       "Rename study session",
			Description: "Rename a study session.",
			Annotations: toolAnnotations("Rename study session", false, false),
		},
		func(ctx context.Context, in renameSessionInput) (store.SessionSummary, string, error) {
			session, err := s.store.RenameSession(ctx, in.BookID, in.SessionID, in.NewName)
			return session, fmt.Sprintf("Renamed session to %q.", session.Name), err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "delete_session",
			Title:       "Delete study session",
			Description: "Delete one study session.",
			Annotations: toolAnnotations("Delete study session", false, true),
		},
		func(ctx context.Context, in sessionInput) (deletedSession, string, error) {
			err := s.store.DeleteSession(ctx, in.BookID, in.SessionID)
			return deletedSession{
				BookID:    in.BookID,
				SessionID: in.SessionID,
				Deleted:   err == nil,
			}, "Deleted the study session.", err
		},
	)

	addPageTool(
		s.server,
		&mcp.Tool{
			Name:        "read_pages",
			Title:       "Read book pages",
			Description: "Read any valid inclusive page range.",
			Annotations: toolAnnotations("Read book pages", true, false),
		},
		func(ctx context.Context, in readPagesInput) (
			pageBatch,
			[]store.RenderedPage,
			string,
			error,
		) {
			pages, err := s.store.ReadPages(
				ctx,
				in.BookID,
				in.StartPageIndex,
				in.EndPageIndex,
			)
			return batchMetadata(pages),
				pages.Pages,
				fmt.Sprintf("Read %d pages.", len(pages.Pages)),
				err
		},
	)

	addPageTool(
		s.server,
		&mcp.Tool{
			Name:        "goto_page",
			Title:       "Go to page",
			Description: "Reset the reading window at one page.",
			Annotations: toolAnnotations("Go to page", false, false),
		},
		func(ctx context.Context, in gotoPageInput) (
			store.PageSelection,
			[]store.RenderedPage,
			string,
			error,
		) {
			selection, err := s.store.GotoPage(
				ctx,
				in.BookID,
				in.SessionID,
				in.PageIndex,
			)
			return selection,
				selection.Pages,
				fmt.Sprintf("Started at page %d.", in.PageIndex),
				err
		},
	)

	addPageTool(
		s.server,
		&mcp.Tool{
			Name:        "continue_reading",
			Title:       "Continue reading",
			Description: "Save progress and preload the next batch.",
			Annotations: toolAnnotations("Continue reading", false, false),
		},
		func(ctx context.Context, in checkpointInput) (
			store.ReadingBatch,
			[]store.RenderedPage,
			string,
			error,
		) {
			batch, err := s.store.ContinueReading(
				ctx,
				in.BookID,
				in.SessionID,
				in.checkpoint(),
			)
			return batch,
				batch.Pages,
				fmt.Sprintf("Read %d pages.", len(batch.Pages)),
				err
		},
	)

	addTool(
		s.server,
		&mcp.Tool{
			Name:        "save_checkpoint",
			Title:       "Save reading checkpoint",
			Description: "Save the learner checkpoint.",
			Annotations: toolAnnotations("Save reading checkpoint", false, false),
		},
		func(ctx context.Context, in checkpointInput) (store.SessionSummary, string, error) {
			session, err := s.store.SaveCheckpoint(
				ctx,
				in.BookID,
				in.SessionID,
				in.checkpoint(),
			)
			return session,
				fmt.Sprintf("Saved checkpoint at page %d.", session.CheckpointPageIndex),
				err
		},
	)
}
