package jcs

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
)

type TestVector struct {
	Name     string          `json:"name"`
	Input    json.RawMessage `json:"input"`
	Expected string          `json:"expected"`
}

func TestCrossLanguageCompatibility(t *testing.T) {
	// Read test vectors from root directory
	vectorPath := filepath.Join("..", "..", "test-vectors.json")
	data, err := os.ReadFile(vectorPath)
	if err != nil {
		t.Skipf("Skipping cross-lang test: cannot read test-vectors.json: %v", err)
		return
	}

	var vectors []TestVector
	if err := json.Unmarshal(data, &vectors); err != nil {
		t.Fatalf("Failed to parse test vectors: %v", err)
	}

	for _, v := range vectors {
		t.Run(v.Name, func(t *testing.T) {
			// Parse input
			var input interface{}
			if err := json.Unmarshal(v.Input, &input); err != nil {
				t.Fatalf("Failed to parse input: %v", err)
			}

			// Marshal using our JCS implementation
			got, err := Marshal(input)
			if err != nil {
				t.Fatalf("Marshal failed: %v", err)
			}

			if string(got) != v.Expected {
				t.Errorf("\ngot:  %q\nwant: %q", string(got), v.Expected)
			}
		})
	}
}
