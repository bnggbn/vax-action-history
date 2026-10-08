package storage

import (
	"errors"
	"io"
	"io/fs"
)

// Kind classifies storage failures without depending on an application's HTTP codes.
type Kind string

const (
	InvalidInput      Kind = "invalid_input"
	InvalidOptions    Kind = "invalid_options"
	LimitExceeded     Kind = "limit_exceeded"
	Corrupt           Kind = "corrupt"
	Missing           Kind = "missing"
	IO                Kind = "io"
	UnsupportedFormat Kind = "unsupported_format"
	UnsafePath        Kind = "unsafe_path"
)

// Error preserves its underlying cause for errors.Is and errors.As.
// Message is diagnostic text, not a stable identifier or a public API response.
type Error struct {
	Kind    Kind
	Message string
	cause   error
}

func (e *Error) Error() string {
	if e.cause == nil {
		return e.Message
	}
	return e.Message + ": " + e.cause.Error()
}
func (e *Error) Unwrap() error                 { return e.cause }
func newError(kind Kind, message string) error { return &Error{Kind: kind, Message: message} }
func wrapError(kind Kind, message string, cause error) error {
	if cause == nil {
		return nil
	}
	return &Error{Kind: kind, Message: message, cause: cause}
}
func kindOf(err error) (Kind, bool) {
	var typed *Error
	if errors.As(err, &typed) {
		return typed.Kind, true
	}
	return "", false
}
func readError(err error) error {
	if err == nil {
		return nil
	}
	if _, found := kindOf(err); found {
		return err
	}
	if errors.Is(err, fs.ErrNotExist) {
		return wrapError(Missing, "required storage object is missing", err)
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		return wrapError(Corrupt, "storage data is truncated", err)
	}
	return wrapError(IO, "could not read storage data", err)
}
func writeError(err error) error {
	if err == nil {
		return nil
	}
	if _, found := kindOf(err); found {
		return err
	}
	return wrapError(IO, "could not write storage data", err)
}
