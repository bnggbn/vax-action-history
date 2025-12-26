// vax_test.go

package vax

import (
	"bytes"
	"encoding/hex"
	"testing"
)

// Test vectors (matching C test suite)
var (
	testKChain = []byte{
		0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08,
		0x09, 0x0a, 0x0b, 0x0c, 0x0d, 0x0e, 0x0f, 0x10,
		0x11, 0x12, 0x13, 0x14, 0x15, 0x16, 0x17, 0x18,
		0x19, 0x1a, 0x1b, 0x1c, 0x1d, 0x1e, 0x1f, 0x20,
	}

	testGenesisSalt = []byte{
		0xa1, 0xa2, 0xa3, 0xa4, 0xa5, 0xa6, 0xa7, 0xa8,
		0xa9, 0xaa, 0xab, 0xac, 0xad, 0xae, 0xaf, 0xb0,
	}
)

func TestComputeGI(t *testing.T) {
	t.Run("basic", func(t *testing.T) {
		gi, err := ComputeGI(testKChain, 1)
		if err != nil {
			t.Fatalf("ComputeGI failed: %v", err)
		}

		if len(gi) != GISize {
			t.Errorf("gi length = %d, want %d", len(gi), GISize)
		}

		t.Logf("gi (counter=1): %x", gi)
	})

	t.Run("deterministic", func(t *testing.T) {
		gi1, _ := ComputeGI(testKChain, 42)
		gi2, _ := ComputeGI(testKChain, 42)

		if !bytes.Equal(gi1, gi2) {
			t.Error("ComputeGI is not deterministic")
		}
	})

	t.Run("counter changes", func(t *testing.T) {
		gi1, _ := ComputeGI(testKChain, 1)
		gi2, _ := ComputeGI(testKChain, 2)

		if bytes.Equal(gi1, gi2) {
			t.Error("different counters produced same gi")
		}
	})

	t.Run("error: invalid k_chain length", func(t *testing.T) {
		_, err := ComputeGI([]byte{0x01, 0x02}, 1)
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("known vector", func(t *testing.T) {
		// Test vector from C test suite
		kChain := make([]byte, KChainSize)
		gi, err := ComputeGI(kChain, 1)
		if err != nil {
			t.Fatalf("ComputeGI failed: %v", err)
		}

		// Expected from C test: OpenSSL HMAC-SHA256
		expected := "96b0dbcec77032023871b0df25214723e5b053da24d50b8f3338ea55f9966a69"
		got := hex.EncodeToString(gi)

		if got != expected {
			t.Errorf("\ngot:  %s\nwant: %s", got, expected)
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

		gi := make([]byte, GISize)
		for i := range gi {
			gi[i] = 0x22
		}

		sai, err := ComputeSAI(prevSAI, sae, gi)
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
		sae := []byte(`{"test":1}`)
		gi := make([]byte, GISize)

		sai1, _ := ComputeSAI(prevSAI, sae, gi)
		sai2, _ := ComputeSAI(prevSAI, sae, gi)

		if !bytes.Equal(sai1, sai2) {
			t.Error("ComputeSAI is not deterministic")
		}
	})

	t.Run("different SAE", func(t *testing.T) {
		prevSAI := make([]byte, SAISize)
		gi := make([]byte, GISize)

		sai1, _ := ComputeSAI(prevSAI, []byte(`{"action":"test1"}`), gi)
		sai2, _ := ComputeSAI(prevSAI, []byte(`{"action":"test2"}`), gi)

		if bytes.Equal(sai1, sai2) {
			t.Error("different SAE produced same SAI")
		}
	})

	t.Run("error: invalid prevSAI length", func(t *testing.T) {
		_, err := ComputeSAI([]byte{0x01}, []byte("test"), make([]byte, GISize))
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: invalid gi length", func(t *testing.T) {
		_, err := ComputeSAI(make([]byte, SAISize), []byte("test"), []byte{0x01})
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})

	t.Run("error: empty SAE", func(t *testing.T) {
		_, err := ComputeSAI(make([]byte, SAISize), []byte{}, make([]byte, GISize))
		if err != ErrInvalidInput {
			t.Errorf("expected ErrInvalidInput, got %v", err)
		}
	})
}

func TestVerifyAction(t *testing.T) {
	t.Run("valid action", func(t *testing.T) {
		kChain := make([]byte, KChainSize)
		for i := range kChain {
			kChain[i] = 0x42
		}

		expectedPrevSAI := make([]byte, SAISize)
		for i := range expectedPrevSAI {
			expectedPrevSAI[i] = 0xAA
		}

		// Create valid action
		counter := uint16(1)
		prevSAI := make([]byte, SAISize)
		copy(prevSAI, expectedPrevSAI)

		sae := []byte(`{"action":"test"}`)

		// Compute gi and sai
		gi, err := ComputeGI(kChain, counter)
		if err != nil {
			t.Fatalf("ComputeGI failed: %v", err)
		}

		sai, err := ComputeSAI(prevSAI, sae, gi)
		if err != nil {
			t.Fatalf("ComputeSAI failed: %v", err)
		}

		// Verify
		err = VerifyAction(
			kChain,
			0,               // expected counter
			expectedPrevSAI, // expected prevSAI
			counter,
			prevSAI,
			sae,
			sai,
		)

		if err != nil {
			t.Errorf("VerifyAction failed: %v", err)
		}
	})

	t.Run("invalid counter", func(t *testing.T) {
		kChain := make([]byte, KChainSize)
		prevSAI := make([]byte, SAISize)
		sae := []byte(`{"test":1}`)
		sai := make([]byte, SAISize)

		err := VerifyAction(
			kChain,
			5, // expected = 5
			prevSAI,
			10, // submitted = 10 (should be 6)
			prevSAI,
			sae,
			sai,
		)

		if err != ErrInvalidCounter {
			t.Errorf("expected ErrInvalidCounter, got %v", err)
		}
	})

	t.Run("invalid prevSAI", func(t *testing.T) {
		kChain := make([]byte, KChainSize)

		expectedPrevSAI := make([]byte, SAISize)
		for i := range expectedPrevSAI {
			expectedPrevSAI[i] = 0xAA
		}

		wrongPrevSAI := make([]byte, SAISize)
		for i := range wrongPrevSAI {
			wrongPrevSAI[i] = 0xBB
		}

		sae := []byte(`{"test":1}`)
		sai := make([]byte, SAISize)

		err := VerifyAction(
			kChain,
			0,
			expectedPrevSAI,
			1,
			wrongPrevSAI, // Different prevSAI
			sae,
			sai,
		)

		if err != ErrInvalidPrevSAI {
			t.Errorf("expected ErrInvalidPrevSAI, got %v", err)
		}
	})

	t.Run("invalid SAI", func(t *testing.T) {
		kChain := make([]byte, KChainSize)
		for i := range kChain {
			kChain[i] = 0x42
		}

		prevSAI := make([]byte, SAISize)
		for i := range prevSAI {
			prevSAI[i] = 0xAA
		}

		sae := []byte(`{"action":"test"}`)

		// Compute correct gi but provide wrong SAI
		gi, _ := ComputeGI(kChain, 1)
		_ = gi

		wrongSAI := make([]byte, SAISize)
		for i := range wrongSAI {
			wrongSAI[i] = 0xFF
		}

		err := VerifyAction(
			kChain,
			0,
			prevSAI,
			1,
			prevSAI,
			sae,
			wrongSAI,
		)

		if err != ErrSAIMismatch {
			t.Errorf("expected ErrSAIMismatch, got %v", err)
		}
	})
}

func TestChainSimulation(t *testing.T) {
	// Setup
	actorID := "alice:laptop"
	kChain := make([]byte, KChainSize)
	for i := range kChain {
		kChain[i] = 0x42
	}

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
	counter := uint16(1)
	sae1 := []byte(`{"action":"create","id":1}`)

	gi1, err := ComputeGI(kChain, counter)
	if err != nil {
		t.Fatalf("ComputeGI(1) failed: %v", err)
	}

	sai1, err := ComputeSAI(prevSAI, sae1, gi1)
	if err != nil {
		t.Fatalf("ComputeSAI(1) failed: %v", err)
	}
	t.Logf("SAI_1: %x", sai1)

	// Action 2
	counter = 2
	sae2 := []byte(`{"action":"update","id":1}`)

	gi2, err := ComputeGI(kChain, counter)
	if err != nil {
		t.Fatalf("ComputeGI(2) failed: %v", err)
	}

	sai2, err := ComputeSAI(sai1, sae2, gi2)
	if err != nil {
		t.Fatalf("ComputeSAI(2) failed: %v", err)
	}
	t.Logf("SAI_2: %x", sai2)

	// Verify chain properties
	if bytes.Equal(gi1, gi2) {
		t.Error("gi1 and gi2 should be different")
	}
	if bytes.Equal(sai1, sai2) {
		t.Error("sai1 and sai2 should be different")
	}
}

// Benchmark tests
func BenchmarkComputeGI(b *testing.B) {
	kChain := make([]byte, KChainSize)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeGI(kChain, uint16(i%65536))
	}
}

func BenchmarkComputeSAI(b *testing.B) {
	prevSAI := make([]byte, SAISize)
	sae := []byte(`{"action":"test","value":42}`)
	gi := make([]byte, GISize)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, _ = ComputeSAI(prevSAI, sae, gi)
	}
}

func BenchmarkVerifyAction(b *testing.B) {
	kChain := make([]byte, KChainSize)
	prevSAI := make([]byte, SAISize)
	sae := []byte(`{"action":"test"}`)

	// Prepare valid action
	gi, _ := ComputeGI(kChain, 1)
	sai, _ := ComputeSAI(prevSAI, sae, gi)

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = VerifyAction(kChain, 0, prevSAI, 1, prevSAI, sae, sai)
	}
}
