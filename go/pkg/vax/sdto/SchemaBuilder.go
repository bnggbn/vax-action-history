package sdto

type SchemaBuilder struct {
	Actions map[string]FieldSpec
}

// 啟動點
func NewSchemaBuilder() *SchemaBuilder {
	return &SchemaBuilder{
		Actions: make(map[string]FieldSpec),
	}
}

// 設定行動字串長度限制
func (b *SchemaBuilder) SetActionStringLength(action string, min string, max string) *SchemaBuilder {
	b.Actions[action] = FieldSpec{
		Type: "string",
		Min:  &min,
		Max:  &max,
	}
	return b
}

// 設定行動數字範圍限制
func (b *SchemaBuilder) SetActionNumberRange(action string, min string, max string) *SchemaBuilder {
	b.Actions[action] = FieldSpec{
		Type: "number",
		Min:  &min,
		Max:  &max,
	}
	return b
}

// 設定行動列舉限制
func (b *SchemaBuilder) SetActionEnum(action string, values []string) *SchemaBuilder {
	b.Actions[action] = FieldSpec{
		Type: "string",
		Enum: values,
	}
	return b
}

// 支援的簽名類型
var SupportedSignTypes = []string{"ed25519", "rsa", "ecdsa"}

// 設定簽名欄位，指定簽名演算法類型
func (b *SchemaBuilder) SetActionSign(action string, signType string) *SchemaBuilder {
	b.Actions[action] = FieldSpec{
		Type: "sign",
		Enum: []string{signType},
	}
	return b
}

// 設定簽名欄位，允許多種簽名類型
func (b *SchemaBuilder) SetActionSignMulti(action string, signTypes []string) *SchemaBuilder {
	b.Actions[action] = FieldSpec{
		Type: "sign",
		Enum: signTypes,
	}
	return b
}

// BuildSchema 回傳給 constructor 用的 FieldSpec map
func (b *SchemaBuilder) BuildSchema() map[string]FieldSpec {
	return b.Actions
}

// Build 回傳 JSON 友善格式（跨語言傳輸用）
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
