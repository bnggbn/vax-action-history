package vax

import (
	"crypto/rand"
	"crypto/sha256"
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
	GenesisSaltSize = 16
)

// ComputeGI = random 32 bytes
func computeGI() ([]byte, error) {
    gi := make([]byte, 32) // 256-bit
    if _, err := rand.Read(gi); err != nil {
        return nil, err
    }
    return gi, nil
}

// ComputeSAI computes SAI_n = SHA256("VAX-SAI" || prevSAI || SHA256(SAE) || gi)
func ComputeSAI(prevSAI, saeBytes []byte) ([]byte, error) {
	if len(prevSAI) != SAISize {
		return nil, ErrInvalidInput
	}
	if len(saeBytes) == 0 {
		return nil, ErrInvalidInput
	}

	// Two-stage hash
	saeHash := sha256.Sum256(saeBytes)

	gi, err := computeGI()
	if err != nil {
		return nil, err
	}

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
	expectedPrevSAI []byte,
	prevSAI []byte,
) error {

	if len(expectedPrevSAI) != SAISize {
		return ErrInvalidInput
	}
	if len(prevSAI) != SAISize {
		return ErrInvalidInput
	}

	// Verify prevSAI matches
	if !bytesEqual(prevSAI, expectedPrevSAI) {
		return ErrInvalidPrevSAI
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
