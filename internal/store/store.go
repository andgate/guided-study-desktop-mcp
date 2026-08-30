package store

import (
	"context"
	"database/sql"
	_ "embed"
	"fmt"
	"strings"

	_ "modernc.org/sqlite"
)

//go:embed schema.sql
var schemaSQL string

// Store reads and writes study data.
type Store struct{ db *sql.DB }

func Open(ctx context.Context, path string) (*Store, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, err
	}

	// Use one SQLite connection.
	db.SetMaxOpenConns(1)
	if _, err = db.ExecContext(
		ctx,
		"PRAGMA foreign_keys = ON; PRAGMA busy_timeout = 5000;",
	); err != nil {
		db.Close()
		return nil, err
	}
	if _, err = db.ExecContext(ctx, schemaSQL); err != nil {
		db.Close()
		return nil, fmt.Errorf("initialize schema: %w", err)
	}
	return &Store{db: db}, nil
}

func (s *Store) Close() error { return s.db.Close() }

// String helpers reject blank values.
func cleanRequired(name, value string) (string, *Error) {
	v := strings.TrimSpace(value)
	if v == "" {
		return "", errf(
			CodeInvalidArgument,
			map[string]any{"field": name},
			"%s must not be blank.",
			name,
		)
	}
	return v, nil
}

func cleanOptional(value *string) *string {
	if value == nil {
		return nil
	}
	v := strings.TrimSpace(*value)
	if v == "" {
		return nil
	}
	return &v
}

func scanNullableString(value sql.NullString) *string {
	if !value.Valid {
		return nil
	}
	v := value.String
	return &v
}

// isConstraint reports whether err is a SQLite constraint error.
func isConstraint(err error) bool {
	return err != nil && strings.Contains(strings.ToLower(err.Error()), "constraint")
}
