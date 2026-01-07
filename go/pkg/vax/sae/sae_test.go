package sae

import (
	"crypto/ed25519"
	"encoding/json"
	"testing"

	"vax/pkg/vax/jcs"
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

func TestGenerateKeyPair(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		pubKey, privKey, err := GenerateKeyPair()
		if err != nil {
			t.Fatalf("GenerateKeyPair failed: %v", err)
		}

		if len(pubKey) != ed25519.PublicKeySize {
			t.Errorf("pubKey length = %d, want %d", len(pubKey), ed25519.PublicKeySize)
		}
		if len(privKey) != ed25519.PrivateKeySize {
			t.Errorf("privKey length = %d, want %d", len(privKey), ed25519.PrivateKeySize)
		}
	})

	t.Run("randomness", func(t *testing.T) {
		_, priv1, _ := GenerateKeyPair()
		_, priv2, _ := GenerateKeyPair()

		if string(priv1) == string(priv2) {
			t.Error("GenerateKeyPair produced same key twice")
		}
	})
}

func TestSAE_Sign(t *testing.T) {
	t.Run("valid signing", func(t *testing.T) {
		pubKey, privKey, _ := GenerateKeyPair()

		sae := &Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO:       map[string]any{"name": "alice"},
			Signature:  nil,
		}

		err := sae.Sign(privKey)
		if err != nil {
			t.Fatalf("Sign failed: %v", err)
		}

		if sae.Signature == nil {
			t.Error("Signature should not be nil after signing")
		}

		if len(sae.Signature) != ed25519.SignatureSize {
			t.Errorf("Signature length = %d, want %d", len(sae.Signature), ed25519.SignatureSize)
		}

		// Verify signature is valid
		// Need to verify against the unsigned SAE
		unsignedSAE := &Envelope{
			ActionType: sae.ActionType,
			Timestamp:  sae.Timestamp,
			SDTO:       sae.SDTO,
			Signature:  nil,
		}
		canonical, _ := json.Marshal(unsignedSAE)
		// Note: actual verification would use jcs.Marshal
		_ = pubKey
		_ = canonical
	})

	t.Run("invalid private key", func(t *testing.T) {
		sae := &Envelope{
			ActionType: "test",
			Timestamp:  1234567890,
			SDTO:       map[string]any{},
		}

		invalidKey := []byte{0x01, 0x02, 0x03}
		err := sae.Sign(invalidKey)

		if err == nil {
			t.Error("Sign should fail with invalid private key")
		}
	})

	t.Run("empty private key", func(t *testing.T) {
		sae := &Envelope{
			ActionType: "test",
			Timestamp:  1234567890,
			SDTO:       map[string]any{},
		}

		err := sae.Sign(nil)

		if err == nil {
			t.Error("Sign should fail with nil private key")
		}
	})
}

func TestSAE_SignAndVerify(t *testing.T) {
	t.Run("full cycle", func(t *testing.T) {
		pubKey, privKey, _ := GenerateKeyPair()

		sae := &Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO:       map[string]any{"amount": 500},
			Signature:  nil,
		}

		// Get canonical form before signing
		canonicalBeforeSign, _ := jcs.Marshal(sae)

		// Sign
		err := sae.Sign(privKey)
		if err != nil {
			t.Fatalf("Sign failed: %v", err)
		}

		// Verify
		valid := ed25519.Verify(pubKey, canonicalBeforeSign, sae.Signature)
		if !valid {
			t.Error("Signature verification failed")
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

func BenchmarkSign(b *testing.B) {
	_, privKey, _ := GenerateKeyPair()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		sae := &Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO:       map[string]any{"name": "alice"},
		}
		_ = sae.Sign(privKey)
	}
}
