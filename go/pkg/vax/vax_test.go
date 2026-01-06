// vax_test.go

package vax

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Test vectors (matching C test suite)
var (
	testGenesisSalt = []byte{
		0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
		0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0,
	}
)

func TestComputeGI(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		gi, err := computeGI()
		if err != nil {
			t.Fatalf("computeGI failed: %v", err)
		}

		if len(gi) != GISize {
			t.Errorf("gi length = %d, want %d", len(gi), GISize)
		}

		t.Logf("gi: %x", gi)
	})

	t.Run("randomness", func(t *testing.T) {
		gi1, _ := computeGI()
		gi2, _ := computeGI()

		// Random values should be different (extremely high probability)
		if bytes.Equal(gi1, gi2) {
			t.Error("computeGI produced same output twice (should be random)")
		}
	})

	t.Run("non-zero", func(t *testing.T) {
		gi, _ := computeGI()
		zeros := make([]byte, GISize)

		if bytes.Equal(gi, zeros) {
			t.Error("computeGI produced all zeros")
		}
	})
}

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

	t.Run("randomness due to gi", func(t *testing.T) {
		prevSAI := make([]byte, SAISize)
		sae := []byte(`{"test":1}`)

		// Since gi is random, same inputs produce different outputs
		sai1, _ := ComputeSAI(prevSAI, sae)
		sai2, _ := ComputeSAI(prevSAI, sae)

		if bytes.Equal(sai1, sai2) {
			t.Error("ComputeSAI produced same output (gi should be random)")
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
	t.Run("valid action", func(t *testing.T) {
		expectedPrevSAI := make([]byte, SAISize)
		for i := range expectedPrevSAI {
			expectedPrevSAI[i] = 0xAA
		}

		prevSAI := make([]byte, SAISize)
		copy(prevSAI, expectedPrevSAI)

		// Verify
		err := VerifyAction(expectedPrevSAI, prevSAI)

		if err != nil {
			t.Errorf("VerifyAction failed: %v", err)
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

		err := VerifyAction(expectedPrevSAI, wrongPrevSAI)

		if err != ErrInvalidPrevSAI {
			t.Errorf("expected ErrInvalidPrevSAI, got %v", err)
		}
	})

	t.Run("error: invalid expectedPrevSAI length", func(t *testing.T) {
		err := VerifyAction([]byte{0x01}, make([]byte, SAISize))
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: invalid prevSAI length", func(t *testing.T) {
		err := VerifyAction(make([]byte, SAISize), []byte{0x01})
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
func BenchmarkComputeGI(b *testing.B) {
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = computeGI()
	}
}

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

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyAction(prevSAI, prevSAI)
	}
}
