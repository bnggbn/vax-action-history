package jcs

import (
	"bytes"
	_ "embed"
	"encoding/json"
	"errors"
	"os"
	"path/filepath"
	"testing"
)

// The module ships its vectors so published-module tests cannot silently skip.
// Repository tests also require the shared Go/TS fixture to match this copy.
//
//go:embed testdata/cross-language.json
var crossLanguageVectors []byte

type TestVector struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected string          `json:"expected"`
}

func TestCrossLanguageCompatibility(t *testing.T) {
	root := filepath.Join("..", "..", "..", "..")
	if _, err := os.Stat(filepath.Join(root, "ts", "package.json")); err == nil {
		data, err := os.ReadFile(filepath.Join(root, "test-vectors.json"))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(data, crossLanguageVectors) {
			t.Fatal("shared and module-packaged test vectors differ")
		}
	} else if !errors.Is(err, os.ErrNotExist) {
		t.Fatal(err)
	}
	var vectors []TestVector
	if err := json.Unmarshal(crossLanguageVectors, &vectors); err != nil {
		t.Fatal(err)
	}
	for _, vector := range vectors {
		t.Run(vector.Name, func(t *testing.T) {
			got, err := CanonicalizeJSON(vector.Input)
			if err != nil {
				t.Fatal(err)
			}
			if string(got) != vector.Expected {
				t.Errorf("got: %q; want: %q", got, vector.Expected)
			}
		})
	}
}
