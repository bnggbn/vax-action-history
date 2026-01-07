package sae

import (
	"time"

	"vax/pkg/vax/jcs"
)

type Envelope struct {
	ActionType string         `json:"action_type"`
	Timestamp  int64          `json:"timestamp"`
	SDTO       map[string]any `json:"sdto"`
}

// BuildSAE builds a Semantic Action Envelope using the project's JCS canonicalizer.
func BuildSAE(actionType string, sdto map[string]any) ([]byte, error) {
	env := Envelope{
		ActionType: actionType,
		Timestamp:  time.Now().UnixMilli(),
		SDTO:       sdto,
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


