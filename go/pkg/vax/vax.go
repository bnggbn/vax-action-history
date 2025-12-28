package vax

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/binary"
	"errors"
)

// Error codes
var (
	ErrInvalidCounter  = errors.New("invalid counter")
	ErrInvalidPrevSAI  = errors.New("invalid prevSAI")
	ErrSAIMismatch     = errors.New("SAI mismatch")
	ErrOutOfMemory     = errors.New("out of memory")
	ErrInvalidInput    = errors.New("invalid input")
	ErrCounterOverflow = errors.New("counter overflow")
)

// Constants
const (
	SAISize         = 32
	GISize          = 32
	KChainSize      = 32
	GenesisSaltSize = 16
)

// ComputeGI computes gi_n = HMAC_SHA256(k_chain, "VAX-GI" || counter)
func ComputeGI(kChain []byte, counter uint16) ([]byte, error) {
	if len(kChain) != KChainSize {
		return nil, ErrInvalidInput
	}

	// Message: "VAX-GI" || counter (big-endian)
	message := make([]byte, 8)
	copy(message, "VAX-GI")
	binary.BigEndian.PutUint16(message[6:], counter)

	// HMAC-SHA256
	mac := hmac.New(sha256.New, kChain)
	mac.Write(message)
	return mac.Sum(nil), nil
}

// ComputeSAI computes SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)
func ComputeSAI(prevSAI, saeBytes, gi []byte) ([]byte, error) {
	if len(prevSAI) != SAISize {
		return nil, ErrInvalidInput
	}
	if len(gi) != GISize {
		return nil, ErrInvalidInput
	}
	if len(saeBytes) == 0 {
		return nil, ErrInvalidInput
	}

	// Two-stage hash
	saeHash := sha256.Sum256(saeBytes)

	// vax sai = 11
	// message = "VAX-SAI" || prevSAI || saeHash || gi
	message := make([]byte, 0, 7+SAISize+SAISize+GISize)
	message = append(message, "VAX-SAI"...)
	message = append(message, prevSAI...)
	message = append(message, saeHash[:]...)
	message = append(message, gi...)

	hash := sha256.Sum256(message)
	return hash[:], nil
}

// ComputeGenesisSAI computes genesis SAI_0 = SHA256("VAX-GENESIS" || actor_id || genesis_salt)
func ComputeGenesisSAI(actorID string, genesisSalt []byte) ([]byte, error) {
	if len(genesisSalt) != GenesisSaltSize {
		return nil, ErrInvalidInput
	}

	//VAX-GENESIS length = 11
	message := make([]byte, 0, 11+len(actorID)+GenesisSaltSize)
	message = append(message, "VAX-GENESIS"...)
	message = append(message, []byte(actorID)...)
	message = append(message, genesisSalt...)

	hash := sha256.Sum256(message)
	return hash[:], nil
}

// VerifyAction verifies an action submission (crypto only, no JSON validation)
func VerifyAction(
	kChain []byte,
	expectedCounter uint16,
	expectedPrevSAI []byte,
	counter uint16,
	prevSAI []byte,
	saeBytes []byte,
	sai []byte,
) error {
	if len(kChain) != KChainSize {
		return ErrInvalidInput
	}
	if len(expectedPrevSAI) != SAISize {
		return ErrInvalidInput
	}
	if len(prevSAI) != SAISize {
		return ErrInvalidInput
	}
	if len(sai) != SAISize {
		return ErrInvalidInput
	}
	if len(saeBytes) == 0 {
		return ErrInvalidInput
	}

	// Check counter overflow
	if expectedCounter == 65535 {
		return ErrCounterOverflow
	}

	// Verify counter is expected + 1
	if counter != expectedCounter+1 {
		return ErrInvalidCounter
	}

	// Verify prevSAI matches
	if !bytesEqual(prevSAI, expectedPrevSAI) {
		return ErrInvalidPrevSAI
	}

	// Recompute gi
	computedGI, err := ComputeGI(kChain, counter)
	if err != nil {
		return err
	}

	// Recompute SAI
	computedSAI, err := ComputeSAI(prevSAI, saeBytes, computedGI)
	if err != nil {
		return err
	}

	// Verify SAI matches
	if !bytesEqual(sai, computedSAI) {
		return ErrSAIMismatch
	}

	return nil
}

func bytesEqual(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
