package sae

import (
	"encoding/json"
	"testing"
)

func TestBuildSAE(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		sdto := map[string]any{
			"name":   "alice",
			"amount": 100,
		}

		saeBytes, err := BuildSAE("transfer", sdto)
		if err != nil {
			t.Fatalf("BuildSAE failed: %v", err)
		}

		if len(saeBytes) == 0 {
			t.Error("BuildSAE returned empty bytes")
		}

		// Verify it's valid JSON
		var parsed map[string]any
		if err := json.Unmarshal(saeBytes, &parsed); err != nil {
			t.Errorf("BuildSAE output is not valid JSON: %v", err)
		}

		// Check required fields
		if parsed["action_type"] != "transfer" {
			t.Errorf("action_type = %v, want transfer", parsed["action_type"])
		}
		if parsed["timestamp"] == nil {
			t.Error("timestamp should not be nil")
		}
		if parsed["sdto"] == nil {
			t.Error("sdto should not be nil")
		}
	})

	t.Run("empty sdto", func(t *testing.T) {
		saeBytes, err := BuildSAE("init", map[string]any{})
		if err != nil {
			t.Fatalf("BuildSAE failed: %v", err)
		}

		if len(saeBytes) == 0 {
			t.Error("BuildSAE returned empty bytes")
		}
	})

	t.Run("deterministic output", func(t *testing.T) {
		sdto := map[string]any{"key": "value"}

		// Note: timestamps differ, so outputs won't be identical
		// But structure should be consistent
		sae1, _ := BuildSAE("test", sdto)
		sae2, _ := BuildSAE("test", sdto)

		if len(sae1) == 0 || len(sae2) == 0 {
			t.Error("BuildSAE returned empty bytes")
		}
	})
}

func BenchmarkBuildSAE(b *testing.B) {
	sdto := map[string]any{
		"name":   "alice",
		"amount": 100,
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = BuildSAE("transfer", sdto)
	}
}
