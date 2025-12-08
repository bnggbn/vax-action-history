package sae

import (
	"time"

	"vax/internal/jcs"
)

type SAE struct {
	ActionType string                 `json:"action_type"`
	Timestamp  int64                  `json:"timestamp"`
	Payload    map[string]interface{} `json:"payload"`
}

// BuildSAE builds a Semantic Action Envelope using the project's JCS canonicalizer.
func BuildSAE(actionType string, payload map[string]interface{}) ([]byte, error) {
	env := SAE{
		ActionType: actionType,
		Timestamp:  time.Now().UnixMilli(),
		Payload:    payload,
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
