package jcs

import "testing"

func TestUTF8KeyOrderAcrossEntrypoints(t *testing.T) {
	value := map[string]any{"！": 1, "😀": 2}
	const expected = `{"\uff01":1,"\ud83d\ude00":2}`
	for _, entry := range []struct {
		name   string
		encode func(any) ([]byte, error)
	}{{"value", CanonicalizeValue}, {"marshal", Marshal}} {
		t.Run(entry.name, func(t *testing.T) {
			actual, err := entry.encode(value)
			if err != nil {
				t.Fatal(err)
			}
			if string(actual) != expected {
				t.Fatalf("got %q; want %q", actual, expected)
			}
		})
	}
}
