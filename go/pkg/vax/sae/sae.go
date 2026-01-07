package sae

import (
	"crypto/ed25519"
	"crypto/rand"
	"errors"
	"time"

	"vax/pkg/vax/jcs"
)

type Envelope struct {
	ActionType string         `json:"action_type"`
	Timestamp  int64          `json:"timestamp"`
	SDTO       map[string]any `json:"sdto"`
	Signature  []byte         `json:"signature,omitempty"`
}

// BuildSAE builds a Semantic Action Envelope using the project's JCS canonicalizer.
func BuildSAE(actionType string, sdto map[string]any) ([]byte, error) {
	env := Envelope{
		ActionType: actionType,
		Timestamp:  time.Now().UnixMilli(),
		SDTO:       sdto,
		Signature:  nil,
	}

	// IMPORTANT:
	// We do NOT use json.Marshal()
	// We MUST ONLY use our own JCS canonicalizer.
	canonical, err := jcs.Marshal(env)
	if err != nil {
		return nil, err
	}
	return canonical, nil
}

func (sae *Envelope) Sign(privateKey ed25519.PrivateKey) error {
	if len(privateKey) != ed25519.PrivateKeySize {
		return errors.New("invalid Ed25519 private key")
	}

	canonical, err := jcs.Marshal(sae)
	if err != nil {
		return err
	}

	sae.Signature = ed25519.Sign(privateKey, canonical)
	return nil
}

// GenerateKeyPair generates an Ed25519 public/private key pair.
func GenerateKeyPair() (ed25519.PublicKey, ed25519.PrivateKey, error) {
	// Generate a random Ed25519 key pair
	publicKey, privateKey, err := ed25519.GenerateKey(rand.Reader)
	if err != nil {
		return nil, nil, err
	}
	return publicKey, privateKey, nil
}
