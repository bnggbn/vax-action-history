package vax

import (
	"crypto/sha256"
	"encoding/json"
	"errors"
	"vax/pkg/vax/sae"
	"vax/pkg/vax/sdto"
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
	GenesisSaltSize = 16
)

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

	// vax sai = 11
	// message = "VAX-SAI" || prevSAI || saeHash || gi
	message := make([]byte, 0, 7+SAISize+SAISize)
	message = append(message, "VAX-SAI"...)
	message = append(message, prevSAI...)
	message = append(message, saeHash[:]...)

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

// VerifyAction verifies an action submission (crypto + schema validation)
// saeBytes: canonical JSON bytes from client (already JCS-marshaled by Finalize)
func VerifyAction(
	expectedPrevSAI []byte,
	prevSAI []byte,
	saeBytes []byte,
	clientProvidedSAI []byte,
	schema map[string]sdto.FieldSpec,
) (*sae.Envelope, error) {

	// Input validation
	if len(expectedPrevSAI) != SAISize {
		return nil, ErrInvalidInput
	}
	if len(prevSAI) != SAISize {
		return nil, ErrInvalidInput
	}
	if len(saeBytes) == 0 {
		return nil, ErrInvalidInput
	}

	// Parse SAE from bytes
	var s sae.Envelope
	if err := json.Unmarshal(saeBytes, &s); err != nil {
		return nil, ErrInvalidInput
	}

	// Verify prevSAI matches
	if !bytesEqual(prevSAI, expectedPrevSAI) {
		return nil, ErrInvalidPrevSAI
	}

	// Verify SDTO against schema
	if err := sdto.ValidateData(s.SDTO, schema); err != nil {
		return nil, err
	}

	// Verify clientProvidedSAI length
	if len(clientProvidedSAI) != SAISize {
		return nil, ErrInvalidInput
	}
    // Verify SAI
	computedSAI, err := ComputeSAI(prevSAI, saeBytes)
	if err != nil {
		return nil, err
	}
	if !bytesEqual(computedSAI, clientProvidedSAI) {
		return nil, ErrSAIMismatch
	}
	return &s, nil
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
