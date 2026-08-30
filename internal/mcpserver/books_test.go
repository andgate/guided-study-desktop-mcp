package mcpserver

import (
	"encoding/csv"
	"strings"
	"testing"

	"github.com/andgate/guided-study-desktop-mcp/internal/store"
)

func TestOutlineCSV(t *testing.T) {
	outline := []store.OutlineEntry{
		{
			OutlineIndex: 0,
			Title:        `Names, "Quotes"`,
			PageIndex:    2,
		},
	}

	encoded, err := outlineCSV(outline)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.HasPrefix(encoded, strings.Join(outlineHeader, ",")+"\r\n") {
		t.Fatalf("missing CSV header: %q", encoded)
	}

	rows, err := csv.NewReader(strings.NewReader(encoded)).ReadAll()
	if err != nil {
		t.Fatal(err)
	}
	if len(rows) != 2 || rows[1][1] != outline[0].Title {
		t.Fatalf("unexpected CSV rows: %#v", rows)
	}
}
