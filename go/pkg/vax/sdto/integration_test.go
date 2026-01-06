package sdto

import (
	"testing"
)

func TestBuilderToConstructor_StringField(t *testing.T) {
	// Provider defines schema
	schema := NewSchemaBuilder()
	schema.SetActionStringLength("name", "1", "50")
	schema.SetActionStringLength("email", "5", "100")

	// Consumer uses schema to construct action
	sae, err := NewAction("createUser", schema.BuildSchema()).
		Set("name", "Alice").
		Set("email", "alice@example.com").
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes, got nil")
	}
}

func TestBuilderToConstructor_NumberField(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionNumberRange("amount", "0", "1000000")
	schema.SetActionNumberRange("quantity", "1", "99")

	sae, err := NewAction("purchase", schema.BuildSchema()).
		Set("amount", 500.0).
		Set("quantity", 3).
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes, got nil")
	}
}

func TestBuilderToConstructor_EnumField(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionEnum("status", []string{"pending", "completed", "cancelled"})

	sae, err := NewAction("updateOrder", schema.BuildSchema()).
		Set("status", "pending").
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes, got nil")
	}
}

func TestBuilderToConstructor_ValidationError_UnknownField(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionStringLength("name", "1", "50")

	_, err := NewAction("createUser", schema.BuildSchema()).
		Set("name", "Alice").
		Set("unknown", "value"). // undefined field
		Finalize()

	if err == nil {
		t.Fatal("expected error for unknown field")
	}
}

func TestBuilderToConstructor_ValidationError_StringTooShort(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionStringLength("name", "3", "50") // min length 3

	_, err := NewAction("createUser", schema.BuildSchema()).
		Set("name", "AB"). // length 2, less than min
		Finalize()

	if err == nil {
		t.Fatal("expected error for string too short")
	}
}

func TestBuilderToConstructor_ValidationError_NumberOutOfRange(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionNumberRange("amount", "0", "100")

	_, err := NewAction("purchase", schema.BuildSchema()).
		Set("amount", 150.0). // exceeds max
		Finalize()

	if err == nil {
		t.Fatal("expected error for number out of range")
	}
}

func TestBuilderToConstructor_ValidationError_EnumInvalid(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionEnum("status", []string{"pending", "completed"})

	_, err := NewAction("updateOrder", schema.BuildSchema()).
		Set("status", "invalid"). // not in enum
		Finalize()

	if err == nil {
		t.Fatal("expected error for invalid enum value")
	}
}

// Test ParseSchema: cross-service map[string]any conversion
func TestParseSchema_FromMapAny(t *testing.T) {
	// Simulate map[string]any from JSON deserialization
	raw := map[string]any{
		"name": map[string]any{
			"type": "string",
			"min":  "1",
			"max":  "50",
		},
		"amount": map[string]any{
			"type": "number",
			"min":  "0",
			"max":  "1000",
		},
		"status": map[string]any{
			"type": "string",
			"enum": []any{"pending", "completed"},
		},
	}

	// Convert
	schema := ParseSchema(raw)

	// Use converted schema
	sae, err := NewAction("test", schema).
		Set("name", "Alice").
		Set("amount", 100.0).
		Set("status", "pending").
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes, got nil")
	}
}

func TestParseSchema_ValidationStillWorks(t *testing.T) {
	raw := map[string]any{
		"name": map[string]any{
			"type": "string",
			"min":  "5",
			"max":  "10",
		},
	}

	schema := ParseSchema(raw)

	_, err := NewAction("test", schema).
		Set("name", "AB"). // too short
		Finalize()

	if err == nil {
		t.Fatal("expected validation error")
	}
}
