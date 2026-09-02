package importer

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"os/exec"
	"strings"

	"github.com/andgate/guided-study-desktop-mcp/internal/store"
)

const (
	defaultDPI         = 200
	defaultJPEGQuality = 90
	converterLimit     = 1
)

type Config struct {
	ConverterPath string
	DatabasePath  string
	DPI           int
	JPEGQuality   int
}

type renderConfig struct {
	DPI         int `json:"dpi"`
	JPEGQuality int `json:"jpeg_quality"`
}

type convertRequest struct {
	DatabasePath  string       `json:"database_path"`
	FileReference string       `json:"file_reference"`
	Title         string       `json:"title"`
	Render        renderConfig `json:"render"`
}

type convertFailure struct {
	Code string `json:"code"`
}

// Importer runs the bundled converter.
type Importer struct {
	config Config
	gate   chan struct{}
}

func New(config Config) *Importer {
	if config.DPI == 0 {
		config.DPI = defaultDPI
	}
	if config.JPEGQuality == 0 {
		config.JPEGQuality = defaultJPEGQuality
	}
	return &Importer{
		config: config,
		gate:   make(chan struct{}, converterLimit),
	}
}

func (i *Importer) Import(
	ctx context.Context,
	fileReference, title string,
) (store.BookSummary, error) {
	// Serialize database writers.
	select {
	case i.gate <- struct{}{}:
		defer func() { <-i.gate }()
	case <-ctx.Done():
		return store.BookSummary{}, conversionError(
			"wait for converter",
			ctx.Err(),
			"",
		)
	}

	request := convertRequest{
		DatabasePath:  i.config.DatabasePath,
		FileReference: fileReference,
		Title:         title,
		Render: renderConfig{
			DPI:         i.config.DPI,
			JPEGQuality: i.config.JPEGQuality,
		},
	}

	var input bytes.Buffer
	if err := json.NewEncoder(&input).Encode(request); err != nil {
		return store.BookSummary{}, conversionError("encode request", err, "")
	}

	var output bytes.Buffer
	var diagnostics bytes.Buffer
	cmd := exec.CommandContext(ctx, i.config.ConverterPath)
	cmd.Stdin = &input
	cmd.Stdout = &output
	cmd.Stderr = &diagnostics
	hideConsole(cmd)

	if err := cmd.Run(); err != nil {
		if failure, ok := knownFailure(diagnostics.String()); ok {
			failure.Cause = err
			return store.BookSummary{}, failure
		}
		return store.BookSummary{}, conversionError(
			"run converter",
			err,
			diagnostics.String(),
		)
	}

	var book store.BookSummary
	decoder := json.NewDecoder(&output)
	if err := decoder.Decode(&book); err != nil {
		return book, conversionError("decode result", err, diagnostics.String())
	}
	if err := requireEOF(decoder); err != nil {
		return book, conversionError("decode result", err, diagnostics.String())
	}
	return book, nil
}

func knownFailure(diagnostics string) (*store.Error, bool) {
	var failure convertFailure
	if err := json.Unmarshal([]byte(strings.TrimSpace(diagnostics)), &failure); err != nil {
		return nil, false
	}
	message := ""
	switch failure.Code {
	case store.CodeOutlineRequired:
		message = "Book outline is required."
	case store.CodeOutlineUnusable:
		message = "Book outline cannot be stored."
	default:
		return nil, false
	}
	return &store.Error{
		Code:    failure.Code,
		Message: message,
	}, true
}

func requireEOF(decoder *json.Decoder) error {
	var extra any
	err := decoder.Decode(&extra)
	if err == io.EOF {
		return nil
	}
	if err == nil {
		return fmt.Errorf("unexpected trailing JSON")
	}
	return err
}

func conversionError(reason string, cause error, diagnostics string) *store.Error {
	details := map[string]any{"reason": reason}
	if text := strings.TrimSpace(diagnostics); text != "" {
		details["diagnostics"] = text
	}
	return &store.Error{
		Code:    store.CodeConversionFailed,
		Message: "Book import failed.",
		Details: details,
		Cause:   cause,
	}
}
