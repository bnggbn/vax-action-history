package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"testing"
)

func TestErrorKindsAndCausesSurviveWrapping(t *testing.T) {
	cause := &fs.PathError{Op: "open", Path: "private/object", Err: fs.ErrNotExist}
	typed := readError(cause)
	err := fmt.Errorf("load snapshot: %w", typed)
	var objectError *Error
	var pathError *fs.PathError
	if !errors.As(err, &objectError) || objectError.Kind != Missing ||
		!errors.Is(err, fs.ErrNotExist) || !errors.As(err, &pathError) || pathError != cause {
		t.Fatal("storage error lost its kind or underlying cause")
	}
	if readError(err) != err || writeError(err) != err {
		t.Fatal("typed failure was reclassified")
	}
	for _, test := range []struct {
		err  error
		kind Kind
	}{
		{readError(io.ErrUnexpectedEOF), Corrupt},
		{readError(fs.ErrPermission), IO},
		{writeError(fs.ErrNotExist), IO},
	} {
		if kind, _ := kindOf(test.err); kind != test.kind {
			t.Fatal("incorrect storage failure kind", kind)
		}
	}
	if _, ok := kindOf(errors.New("missing")); ok {
		t.Fatal("classified failure by diagnostic text")
	}
	if readError(nil) != nil || writeError(nil) != nil || wrapError(IO, "unused", nil) != nil {
		t.Fatal("nil cause became a failure")
	}
}
