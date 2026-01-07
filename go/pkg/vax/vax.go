package vax

import (
	"crypto/ed25519"
	"crypto/sha256"
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
func VerifyAction(
	expectedPrevSAI []byte,
	prevSAI []byte,
	sae *sae.SAE,
	schema map[string]sdto.FieldSpec,
	privateKey ed25519.PrivateKey,
) error {

	if len(expectedPrevSAI) != SAISize {
		return ErrInvalidInput
	}
	if len(prevSAI) != SAISize {
		return ErrInvalidInput
	}
	// Check that SAE is unsigned because sign sae just for records that service provider can't massaging the action
	if sae.Signature != nil {
		return ErrInvalidInput
	}

	// Verify prevSAI matches
	if !bytesEqual(prevSAI, expectedPrevSAI) {
		return ErrInvalidPrevSAI
	}

	// Verify SDTO against schema
	if err := sdto.ValidateData(sae.SDTO, schema); err != nil {
		return err
	}

	// all good and sign SAE settle down the history at here
	return sae.Sign(privateKey)
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
