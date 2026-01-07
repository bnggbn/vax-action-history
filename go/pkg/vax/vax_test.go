// vax_test.go

package vax

import (
	"bytes"
	"encoding/hex"
	"testing"

	"vax/pkg/vax/jcs"
	"vax/pkg/vax/sae"
	"vax/pkg/vax/sdto"
)

// Test vectors (matching C test suite)
var (
	testGenesisSalt = []byte{
		0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
		0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0,
	}
)

func TestComputeGenesisSAI(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		actorID := "user123:device456"
		sai, err := ComputeGenesisSAI(actorID, testGenesisSalt)
		if err != nil {
			t.Fatalf("ComputeGenesisSAI failed: %v", err)
		}

		if len(sai) != SAISize {
			t.Errorf("genesis SAI length = %d, want %d", len(sai), SAISize)
		}

		t.Logf("genesis SAI: %x", sai)
	})

	t.Run("known vector", func(t *testing.T) {
		actorID := "user123:device456"
		sai, err := ComputeGenesisSAI(actorID, testGenesisSalt)
		if err != nil {
			t.Fatalf("ComputeGenesisSAI failed: %v", err)
		}

		// Expected from C test suite
		expected := "afc50728cd79e805a8ae06875a1ddf78ca11b0d56ec300b160fb71f50ce658c3"
		got := hex.EncodeToString(sai)

		if got != expected {
			t.Errorf("\ngot:  %s\nwant: %s", got, expected)
		}
	})

	t.Run("error: invalid genesis_salt length", func(t *testing.T) {
		_, err := ComputeGenesisSAI("test", []byte{0x01, 0x02})
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

func TestComputeSAI(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		prevSAI := make([]byte, SAISize)
		for i := range prevSAI {
			prevSAI[i] = 0x11
		}

		sae := []byte(`{"action":"test","value":42}`)

		sai, err := ComputeSAI(prevSAI, sae)
		if err != nil {
			t.Fatalf("ComputeSAI failed: %v", err)
		}

		if len(sai) != SAISize {
			t.Errorf("SAI length = %d, want %d", len(sai), SAISize)
		}

		t.Logf("SAI: %x", sai)
	})

	t.Run("deterministic", func(t *testing.T) {
		prevSAI := make([]byte, SAISize)
		saeData := []byte(`{"test":1}`)

		// Same inputs should produce same output
		sai1, _ := ComputeSAI(prevSAI, saeData)
		sai2, _ := ComputeSAI(prevSAI, saeData)

		if !bytes.Equal(sai1, sai2) {
			t.Error("ComputeSAI should be deterministic")
		}
	})

	t.Run("error: invalid prevSAI length", func(t *testing.T) {
		_, err := ComputeSAI([]byte{0x01}, []byte("test"))
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: empty SAE", func(t *testing.T) {
		_, err := ComputeSAI(make([]byte, SAISize), []byte{})
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

func TestVerifyAction(t *testing.T) {
	// Setup schema
	builder := sdto.NewSchemaBuilder()
	builder.SetActionStringLength("name", "1", "50")
	builder.SetActionNumberRange("amount", "0", "1000")
	schema := builder.BuildSchema()

	// Generate key pair for signing
	pubKey, privKey, err := sae.GenerateKeyPair()
	if err != nil {
		t.Fatalf("GenerateKeyPair failed: %v", err)
	}
	_ = pubKey // for verification later

	// Helper to build SAE bytes
	buildSAEBytes := func(s *sae.Envelope) []byte {
		b, _ := jcs.Marshal(s)
		return b
	}

	t.Run("valid action", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)
		for i := range expectedPrevSAI {
			expectedPrevSAI[i] = 0xAA
		}

		prevSAI := make([]byte, SAISize)
		copy(prevSAI, expectedPrevSAI)

		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO: map[string]any{
				"name":   "alice",
				"amount": 500.0,
			},
			Signature: nil,
		}
		saeBytes := buildSAEBytes(testSAE)

		// Client computes SAI
		clientSAI, _ := ComputeSAI(prevSAI, saeBytes)

		signedSAE, err := VerifyAction(expectedPrevSAI, prevSAI, saeBytes, clientSAI, schema, privKey)

		if err != nil {
			t.Errorf("VerifyAction failed: %v", err)
		}

		// Check that returned SAE is signed
		if signedSAE == nil || signedSAE.Signature == nil {
			t.Error("SAE should be signed after VerifyAction")
		}
	})

	t.Run("invalid prevSAI", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)
		for i := range expectedPrevSAI {
			expectedPrevSAI[i] = 0xAA
		}

		wrongPrevSAI := make([]byte, SAISize)
		for i := range wrongPrevSAI {
			wrongPrevSAI[i] = 0xBB
		}

		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO: map[string]any{
				"name":   "alice",
				"amount": 500.0,
			},
			Signature: nil,
		}
		saeBytes := buildSAEBytes(testSAE)
		clientSAI, _ := ComputeSAI(wrongPrevSAI, saeBytes)

		_, err := VerifyAction(expectedPrevSAI, wrongPrevSAI, saeBytes, clientSAI, schema, privKey)

		if err != ErrInvalidPrevSAI {
			t.Errorf("expected ErrInvalidPrevSAI, got %v", err)
		}
	})

	t.Run("invalid schema - missing field", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)

		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO: map[string]any{
				"name": "alice",
				// missing "amount"
			},
			Signature: nil,
		}
		saeBytes := buildSAEBytes(testSAE)
		clientSAI, _ := ComputeSAI(expectedPrevSAI, saeBytes)

		_, err := VerifyAction(expectedPrevSAI, expectedPrevSAI, saeBytes, clientSAI, schema, privKey)

		if err == nil {
			t.Error("expected error for missing field")
		}
	})

	t.Run("invalid schema - value out of range", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)

		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO: map[string]any{
				"name":   "alice",
				"amount": 9999.0, // > max 1000
			},
			Signature: nil,
		}
		saeBytes := buildSAEBytes(testSAE)
		clientSAI, _ := ComputeSAI(expectedPrevSAI, saeBytes)

		_, err := VerifyAction(expectedPrevSAI, expectedPrevSAI, saeBytes, clientSAI, schema, privKey)

		if err == nil {
			t.Error("expected error for value out of range")
		}
	})

	t.Run("error: already signed SAE", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)

		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO: map[string]any{
				"name":   "alice",
				"amount": 500.0,
			},
			Signature: []byte{0x01, 0x02}, // already signed
		}
		saeBytes := buildSAEBytes(testSAE)
		clientSAI, _ := ComputeSAI(expectedPrevSAI, saeBytes)

		_, err := VerifyAction(expectedPrevSAI, expectedPrevSAI, saeBytes, clientSAI, schema, privKey)

		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: SAI mismatch", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)

		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO: map[string]any{
				"name":   "alice",
				"amount": 500.0,
			},
			Signature: nil,
		}
		saeBytes := buildSAEBytes(testSAE)

		// Wrong SAI
		wrongSAI := make([]byte, SAISize)
		for i := range wrongSAI {
			wrongSAI[i] = 0xFF
		}

		_, err := VerifyAction(expectedPrevSAI, expectedPrevSAI, saeBytes, wrongSAI, schema, privKey)

		if err != ErrSAIMismatch {
			t.Errorf("expected ErrSAIMismatch, got %v", err)
		}
	})

	t.Run("error: invalid expectedPrevSAI length", func(t *testing.T) {
		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO:       map[string]any{"name": "alice", "amount": 500.0},
			Signature:  nil,
		}
		saeBytes := buildSAEBytes(testSAE)
		clientSAI := make([]byte, SAISize)

		_, err := VerifyAction([]byte{0x01}, make([]byte, SAISize), saeBytes, clientSAI, schema, privKey)
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: invalid prevSAI length", func(t *testing.T) {
		testSAE := &sae.Envelope{
			ActionType: "transfer",
			Timestamp:  1234567890,
			SDTO:       map[string]any{"name": "alice", "amount": 500.0},
			Signature:  nil,
		}
		saeBytes := buildSAEBytes(testSAE)
		clientSAI := make([]byte, SAISize)

		_, err := VerifyAction(make([]byte, SAISize), []byte{0x01}, saeBytes, clientSAI, schema, privKey)
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: empty saeBytes", func(t *testing.T) {
		clientSAI := make([]byte, SAISize)
		_, err := VerifyAction(make([]byte, SAISize), make([]byte, SAISize), []byte{}, clientSAI, schema, privKey)
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: invalid JSON", func(t *testing.T) {
		clientSAI := make([]byte, SAISize)
		_, err := VerifyAction(make([]byte, SAISize), make([]byte, SAISize), []byte("not json"), clientSAI, schema, privKey)
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

func TestChainSimulation(t *testing.T) {
	// Setup
	actorID := "alice:laptop"

	genesisSalt := make([]byte, GenesisSaltSize)
	for i := range genesisSalt {
		genesisSalt[i] = 0xAB
	}

	// Genesis
	prevSAI, err := ComputeGenesisSAI(actorID, genesisSalt)
	if err != nil {
		t.Fatalf("ComputeGenesisSAI failed: %v", err)
	}
	t.Logf("Genesis SAI: %x", prevSAI)

	// Action 1
	sae1 := []byte(`{"action":"create","id":1}`)

	sai1, err := ComputeSAI(prevSAI, sae1)
	if err != nil {
		t.Fatalf("ComputeSAI(1) failed: %v", err)
	}
	t.Logf("SAI_1: %x", sai1)

	// Action 2
	sae2 := []byte(`{"action":"update","id":1}`)

	sai2, err := ComputeSAI(sai1, sae2)
	if err != nil {
		t.Fatalf("ComputeSAI(2) failed: %v", err)
	}
	t.Logf("SAI_2: %x", sai2)

	// Verify chain properties - SAIs should be different
	if bytes.Equal(sai1, sai2) {
		t.Error("sai1 and sai2 should be different")
	}
}

// Benchmark tests
func BenchmarkComputeSAI(b *testing.B) {
	prevSAI := make([]byte, SAISize)
	sae := []byte(`{"action":"test","value":42}`)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeSAI(prevSAI, sae)
	}
}

func BenchmarkVerifyAction(b *testing.B) {
	prevSAI := make([]byte, SAISize)
	builder := sdto.NewSchemaBuilder()
	builder.SetActionStringLength("name", "1", "50")
	builder.SetActionNumberRange("amount", "0", "1000")
	schema := builder.BuildSchema()
	_, privKey, _ := sae.GenerateKeyPair()

	testSAE := &sae.Envelope{
		ActionType: "transfer",
		Timestamp:  1234567890,
		SDTO:       map[string]any{"name": "alice", "amount": 500.0},
		Signature:  nil,
	}
	saeBytes, _ := jcs.Marshal(testSAE)
	clientSAI, _ := ComputeSAI(prevSAI, saeBytes)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = VerifyAction(prevSAI, prevSAI, saeBytes, clientSAI, schema, privKey)
	}
}
