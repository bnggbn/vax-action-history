package builder

type SchemaBuilder struct {
	Actions map[string]Constraint
}

type Constraint struct {
	Type string   `json:"type"` // string / number
	Min  *string  `json:"min,omitempty"`
	Max  *string  `json:"max,omitempty"`
	Enum []string `json:"enum,omitempty"`
}

// 啟動點
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		Actions: make(map[string]Constraint),
	}
}

// 設定行動字串長度限制
func (b *SchemaBuilder) SetActionStringLength(action string, min string, max string,
) {
	b.Actions[action] = Constraint{
		Type: "string",
		Min:  &min,
		Max:  &max,
	}
}

// 設定行動數字範圍限制
func (b *SchemaBuilder) SetActionNumberRange(action string, min string, max string) {
	b.Actions[action] = Constraint{
		Type: "number",
		Min:  &min,
		Max:  &max,
	}
}

// 設定行動列舉限制
func (b *SchemaBuilder) SetActionEnum(action string, values []string) {
	b.Actions[action] = Constraint{
		Type: "string",
		Enum: values,
	}
}

// 建立最終的 Schema
func (b *SchemaBuilder) Build() map[string]any {
	props := map[string]any{}

	for name, c := range b.Actions {
		m := map[string]any{
			"type": c.Type,
		}
		if c.Min != nil {
			m["min"] = *c.Min
		}
		if c.Max != nil {
			m["max"] = *c.Max
		}
		if len(c.Enum) > 0 {
			m["enum"] = c.Enum
		}
		props[name] = m
	}

	return map[string]any{
		"type":       "object",
		"properties": props,
	}
}
