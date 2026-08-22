package importer

import (
	"context"
	"encoding/csv"
	"fmt"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"slices"
	"strconv"
	"strings"
	"syscall"

	"github.com/andgate/guided-study-desktop-mcp/internal/store"
)

const (
	defaultDPI           = 200
	defaultJPEGQuality   = 90
	defaultMaxPages      = 2500
	defaultMaxImageBytes = int64(25 << 20)
	defaultMaxTotalBytes = int64(4 << 30)
)

var tocHeader = [...]string{
	"position",
	"depth",
	"title",
	"page_index",
}

type Config struct {
	ConverterPath string
	TempRoot      string
	DPI           int
	JPEGQuality   int
	MaxPages      int
	MaxImageBytes int64
	MaxTotalBytes int64
}

// Importer prepares PDFs for storage.
type Importer struct{ config Config }

func New(config Config) *Importer {
	// Apply defaults for omitted settings.
	if config.DPI == 0 {
		config.DPI = defaultDPI
	}
	if config.JPEGQuality == 0 {
		config.JPEGQuality = defaultJPEGQuality
	}
	if config.MaxPages == 0 {
		config.MaxPages = defaultMaxPages
	}
	if config.MaxImageBytes == 0 {
		config.MaxImageBytes = defaultMaxImageBytes
	}
	if config.MaxTotalBytes == 0 {
		config.MaxTotalBytes = defaultMaxTotalBytes
	}

	return &Importer{config: config}
}

func (i *Importer) Prepare(
	ctx context.Context,
	fileReference, title string,
) (store.PreparedBook, error) {
	title = strings.TrimSpace(title)
	if title == "" {
		return store.PreparedBook{}, &store.Error{
			Code:    store.CodeInvalidArgument,
			Message: "title must not be blank.",
		}
	}

	// Create a temporary directory for converter output.
	staging, err := os.MkdirTemp(i.config.TempRoot, "guided-study-import-")
	if err != nil {
		return store.PreparedBook{}, &store.Error{
			Code:    store.CodeStorageError,
			Message: "Could not create conversion staging directory.",
			Cause:   err,
		}
	}

	// Delete files when the import finishes.
	defer os.RemoveAll(staging)

	// Configure the converter.
	args := []string{
		"--input", fileReference,
		"--output", staging,
		"--dpi", strconv.Itoa(i.config.DPI),
		"--jpeg-quality", strconv.Itoa(i.config.JPEGQuality),
	}

	// Stop canceled conversions.
	cmd := exec.CommandContext(ctx, i.config.ConverterPath, args...)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}

	// Capture converter errors.
	diagnostics, err := cmd.CombinedOutput()
	if err != nil {
		return store.PreparedBook{}, &store.Error{
			Code:    store.CodeConversionFailed,
			Message: "PDF conversion failed.",
			Details: map[string]any{
				"diagnostics": strings.TrimSpace(string(diagnostics)),
			},
			Cause: err,
		}
	}

	// Load the converted files.
	prepared, err := i.loadStaging(staging, title)
	if err != nil {
		if e, ok := err.(*store.Error); ok {
			return store.PreparedBook{}, e
		}

		return store.PreparedBook{}, &store.Error{
			Code:    store.CodeConversionFailed,
			Message: "Converter output was invalid.",
			Details: map[string]any{"reason": err.Error()},
			Cause:   err,
		}
	}

	return prepared, nil
}

// imageName matches converter page filenames.
var imageName = regexp.MustCompile(`^page-([0-9]{4,})\.(jpg|jpeg|png)$`)

func (i *Importer) loadStaging(dir, title string) (store.PreparedBook, error) {
	// Load the pages.
	pages, err := i.loadPages(dir)
	if err != nil {
		return store.PreparedBook{}, err
	}

	// Load the table of contents.
	toc, err := loadTOC(filepath.Join(dir, "toc.csv"), len(pages))
	if err != nil {
		return store.PreparedBook{}, err
	}

	// Return the book.
	return store.PreparedBook{
		Title: title,
		Pages: pages,
		TOC:   toc,
	}, nil
}

func (i *Importer) loadPages(dir string) ([]store.PreparedPage, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	// Collect page images by page number.
	pages := map[int]store.PreparedPage{}
	var total int64
	for _, entry := range entries {
		// Skip the table of contents.
		if entry.Name() == "toc.csv" {
			continue
		}

		// Load and validate one page.
		page, imageBytes, err := i.loadPage(dir, entry.Name())
		if err != nil {
			return nil, err
		}

		// Check the book size.
		total += imageBytes
		if total > i.config.MaxTotalBytes {
			return nil, fmt.Errorf("prepared book exceeds total size limit")
		}

		// Reject duplicate page numbers.
		if _, exists := pages[page.PageIndex]; exists {
			return nil, fmt.Errorf("duplicate page %d", page.PageIndex)
		}

		// Save the page image.
		pages[page.PageIndex] = page
	}

	// Check the page count.
	pageCount := len(pages)
	if pageCount == 0 || pageCount > i.config.MaxPages {
		return nil, fmt.Errorf(
			"page count %d is outside configured limits",
			pageCount,
		)
	}

	// Order pages and reject missing numbers.
	ordered := make([]store.PreparedPage, pageCount)
	for n := 1; n <= pageCount; n++ {
		page, ok := pages[n]
		if !ok {
			return nil, fmt.Errorf("missing page %d", n)
		}
		ordered[n-1] = page
	}

	return ordered, nil
}

func (i *Importer) loadPage(dir, name string) (store.PreparedPage, int64, error) {
	// Read the page number.
	match := imageName.FindStringSubmatch(strings.ToLower(name))
	if match == nil {
		return store.PreparedPage{}, 0, fmt.Errorf("unexpected staging file %s", name)
	}

	pageIndex, err := strconv.Atoi(match[1])
	if err != nil {
		return store.PreparedPage{}, 0, fmt.Errorf("invalid page number: %w", err)
	}

	// Read the page image.
	data, err := os.ReadFile(filepath.Join(dir, name))
	if err != nil {
		return store.PreparedPage{}, 0, err
	}

	// Check the page size.
	imageBytes := int64(len(data))
	if imageBytes > i.config.MaxImageBytes {
		return store.PreparedPage{}, 0, fmt.Errorf("page %d exceeds image size limit", pageIndex)
	}

	// Verify the image format.
	mime := http.DetectContentType(data)
	if mime != "image/jpeg" && mime != "image/png" {
		return store.PreparedPage{}, 0, fmt.Errorf(
			"page %d has unsupported type %s",
			pageIndex,
			mime,
		)
	}

	return store.PreparedPage{
		PageIndex: pageIndex,
		MIMEType:  mime,
		ImageData: data,
	}, imageBytes, nil
}

func loadTOC(path string, pageCount int) ([]store.TOCEntry, error) {
	// Open the table of contents.
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Read the entire CSV file.
	reader := csv.NewReader(file)

	// Require the expected column count.
	reader.FieldsPerRecord = len(tocHeader)
	rows, err := reader.ReadAll()
	if err != nil {
		return nil, err
	}

	// Require the expected CSV header.
	rowCount := len(rows)
	if rowCount == 0 {
		return nil, fmt.Errorf("toc.csv has zero rows")
	}

	if !slices.Equal(rows[0], tocHeader[:]) {
		return nil, fmt.Errorf("toc.csv has an invalid header")
	}

	// Parse each table of contents row.
	entries := make([]store.TOCEntry, rowCount-1)
	for idx, row := range rows[1:] {
		entry, err := parseTOCRow(row, idx, pageCount)
		if err != nil {
			return nil, err
		}

		entries[idx] = entry
	}

	return entries, nil
}

func parseTOCRow(row []string, idx, pageCount int) (store.TOCEntry, error) {
	// Parse the numeric columns.
	position, positionErr := strconv.Atoi(row[0])
	depth, depthErr := strconv.Atoi(row[1])
	page, pageErr := strconv.Atoi(row[3])

	// Validate every field.
	rowNumber := idx + 2
	expectedPosition := idx + 1
	title := strings.TrimSpace(row[2])

	invalidNumber := positionErr != nil || depthErr != nil || pageErr != nil
	invalidPosition := position != expectedPosition
	invalidDepth := depth < 0
	invalidPage := page < 1 || page > pageCount
	invalidTitle := title == ""

	invalidRow := invalidNumber || invalidPosition || invalidDepth || invalidPage || invalidTitle
	if invalidRow {
		return store.TOCEntry{}, fmt.Errorf("invalid TOC row %d", rowNumber)
	}

	return store.TOCEntry{
		Position:  position,
		Depth:     depth,
		Title:     title,
		PageIndex: page,
	}, nil
}
