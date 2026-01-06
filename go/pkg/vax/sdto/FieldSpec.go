package sdto

type FieldSpec struct {
	Type string   `json:"type"` // string / number
	Min  *string  `json:"min,omitempty"`
	Max  *string  `json:"max,omitempty"`
	Enum []string `json:"enum,omitempty"`
}

// ParseSchema converts map[string]any to map[string]FieldSpec
// Used for cross-service deserialization
func ParseSchema(raw map[string]any) map[string]FieldSpec {
	result := make(map[string]FieldSpec)

	for key, val := range raw {
		m, ok := val.(map[string]any)
		if !ok {
			continue
		}

		spec := FieldSpec{}

		if t, ok := m["type"].(string); ok {
			spec.Type = t
		}
		if min, ok := m["min"].(string); ok {
			spec.Min = &min
		}
		if max, ok := m["max"].(string); ok {
			spec.Max = &max
		}
		if enumRaw, ok := m["enum"].([]any); ok {
			for _, e := range enumRaw {
				if s, ok := e.(string); ok {
					spec.Enum = append(spec.Enum, s)
				}
			}
		}
		// Support []string directly
		if enumStr, ok := m["enum"].([]string); ok {
			spec.Enum = enumStr
		}

		result[key] = spec
	}

	return result
}
