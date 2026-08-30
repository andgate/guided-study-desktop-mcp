package store

import "fmt"

const (
	// Argument and lookup failures.
	CodeInvalidArgument = "invalid_argument"
	CodeNotFound        = "not_found"
	CodeAlreadyExists   = "already_exists"
	CodeOutOfBounds     = "out_of_bounds"

	// Reading failures.
	CodeNoNextBatch = "no_next_batch"

	// Stale-write failures.
	CodeDeckRevisionConflict = "deck_revision_conflict"
	CodeCardRevisionConflict = "card_revision_conflict"

	// Import and storage failures.
	CodeConversionFailed = "conversion_failed"
	CodeOutlineRequired  = "outline_required"
	CodeOutlineUnusable  = "outline_unusable"
	CodeStorageError     = "storage_error"
)

// Error is returned by MCP tools.
type Error struct {
	Code    string         `json:"code"`
	Message string         `json:"message"`
	Details map[string]any `json:"details,omitempty"`
	Cause   error          `json:"-"`
}

func (e *Error) Error() string { return e.Message }
func (e *Error) Unwrap() error { return e.Cause }

// errf creates a service error.
func errf(code string, details map[string]any, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...), Details: details}
}

// storageError hides database error details from callers.
func storageError(err error) *Error {
	return &Error{Code: CodeStorageError, Message: "Local storage failed.", Cause: err}
}
