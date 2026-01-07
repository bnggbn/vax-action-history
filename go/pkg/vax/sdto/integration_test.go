package sdto

import (
	"strings"
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

func TestBuilderToConstructor_SignField(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionSign("signature", "ed25519")

	sae, err := NewAction("signedAction", schema.BuildSchema()).
		Set("signature", "base64encodedSignature==").
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes, got nil")
	}
}

func TestBuilderToConstructor_SignMulti(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionSignMulti("signature", []string{"ed25519", "rsa"})

	sae, err := NewAction("signedAction", schema.BuildSchema()).
		Set("signature", "signatureValue123").
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes, got nil")
	}
}

func TestBuilderToConstructor_SignRejectMap(t *testing.T) {
	schema := NewSchemaBuilder()
	schema.SetActionSign("signature", "ed25519")

	// map should be rejected - sign only accepts string
	_, err := NewAction("signedAction", schema.BuildSchema()).
		Set("signature", map[string]any{
			"type":  "ed25519",
			"value": "sig",
		}).
		Finalize()

	if err == nil {
		t.Fatal("expected error for map value in sign field")
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

func TestBuilderToConstructor_ValidationError_MissingField(t *testing.T) {
	schema := NewSchemaBuilder().
		SetActionStringLength("name", "1", "50").
		SetActionStringLength("email", "5", "100").
		SetActionNumberRange("age", "0", "150").
		BuildSchema()

	// Only set 'name', missing 'email' and 'age'
	_, err := NewAction("createUser", schema).
		Set("name", "Alice").
		Finalize()

	if err == nil {
		t.Fatal("expected error for missing required fields")
	}

	// Check error mentions missing fields
	errStr := err.Error()
	if !strings.Contains(errStr, "email") || !strings.Contains(errStr, "age") {
		t.Errorf("expected error to mention missing fields email and age, got: %v", err)
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

// Test ValidateData for server-side validation
func TestValidateData_Valid(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionStringLength("name", "1", "50")
	builder.SetActionNumberRange("amount", "0", "1000")
	schema := builder.BuildSchema()

	data := map[string]any{
		"name":   "alice",
		"amount": 500.0,
	}

	err := ValidateData(data, schema)
	if err != nil {
		t.Errorf("ValidateData failed: %v", err)
	}
}

func TestValidateData_MissingField(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionStringLength("name", "1", "50")
	builder.SetActionNumberRange("amount", "0", "1000")
	schema := builder.BuildSchema()

	data := map[string]any{
		"name": "alice",
		// missing "amount"
	}

	err := ValidateData(data, schema)
	if err == nil {
		t.Error("expected error for missing field")
	}
}

func TestValidateData_ExtraField(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionStringLength("name", "1", "50")
	schema := builder.BuildSchema()

	data := map[string]any{
		"name":  "alice",
		"extra": "field", // not in schema
	}

	err := ValidateData(data, schema)
	if err == nil {
		t.Error("expected error for extra field")
	}
}

func TestValidateData_InvalidValue(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionNumberRange("amount", "0", "100")
	schema := builder.BuildSchema()

	data := map[string]any{
		"amount": 999.0, // exceeds max
	}

	err := ValidateData(data, schema)
	if err == nil {
		t.Error("expected error for value out of range")
	}
}

// Test Build for JSON transport format
func TestBuild_JSONFormat(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionStringLength("name", "1", "50")
	builder.SetActionNumberRange("amount", "0", "1000")
	builder.SetActionEnum("status", []string{"pending", "done"})

	result := builder.Build()

	// Check structure
	if result["type"] != "object" {
		t.Errorf("type = %v, want object", result["type"])
	}

	props, ok := result["properties"].(map[string]any)
	if !ok {
		t.Fatal("properties should be map[string]any")
	}

	// Check name field
	nameField := props["name"].(map[string]any)
	if nameField["type"] != "string" {
		t.Errorf("name.type = %v, want string", nameField["type"])
	}
	if nameField["min"] != "1" {
		t.Errorf("name.min = %v, want 1", nameField["min"])
	}
	if nameField["max"] != "50" {
		t.Errorf("name.max = %v, want 50", nameField["max"])
	}

	// Check amount field
	amountField := props["amount"].(map[string]any)
	if amountField["type"] != "number" {
		t.Errorf("amount.type = %v, want number", amountField["type"])
	}

	// Check status field (enum)
	statusField := props["status"].(map[string]any)
	enum := statusField["enum"].([]string)
	if len(enum) != 2 || enum[0] != "pending" || enum[1] != "done" {
		t.Errorf("status.enum = %v, want [pending done]", enum)
	}
}

func TestBuild_RoundTrip(t *testing.T) {
	// Provider builds schema
	builder := NewSchemaBuilder()
	builder.SetActionStringLength("name", "3", "20")
	builder.SetActionNumberRange("score", "0", "100")

	// Export to JSON format
	exported := builder.Build()
	props := exported["properties"].(map[string]any)

	// Consumer parses schema
	parsed := ParseSchema(props)

	// Validation should work the same
	_, err := NewAction("test", parsed).
		Set("name", "AB"). // too short (min 3)
		Set("score", 50.0).
		Finalize()

	if err == nil {
		t.Error("expected validation error after round-trip")
	}
}

// Sign field integration tests
func TestValidateData_SignField(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionSign("signature", "ed25519")
	schema := builder.BuildSchema()

	data := map[string]any{
		"signature": "validSignatureString",
	}

	err := ValidateData(data, schema)
	if err != nil {
		t.Errorf("ValidateData failed: %v", err)
	}
}

func TestValidateData_SignRejectMap(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionSign("signature", "ed25519")
	schema := builder.BuildSchema()

	data := map[string]any{
		"signature": map[string]any{"type": "ed25519", "value": "sig"},
	}

	err := ValidateData(data, schema)
	if err == nil {
		t.Error("expected error for map value in sign field")
	}
}

func TestValidateData_SignEmptyValue(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionSign("signature", "ed25519")
	schema := builder.BuildSchema()

	data := map[string]any{
		"signature": "",
	}

	err := ValidateData(data, schema)
	if err == nil {
		t.Error("expected error for empty sign value")
	}
}

func TestBuild_SignField(t *testing.T) {
	builder := NewSchemaBuilder()
	builder.SetActionSign("signature", "ed25519")
	builder.SetActionSignMulti("multi_sig", []string{"ed25519", "rsa"})

	result := builder.Build()
	props := result["properties"].(map[string]any)

	// Check sign field
	sigField := props["signature"].(map[string]any)
	if sigField["type"] != "sign" {
		t.Errorf("signature.type = %v, want sign", sigField["type"])
	}
	sigEnum := sigField["enum"].([]string)
	if len(sigEnum) != 1 || sigEnum[0] != "ed25519" {
		t.Errorf("signature.enum = %v, want [ed25519]", sigEnum)
	}

	// Check multi sign field
	multiField := props["multi_sig"].(map[string]any)
	multiEnum := multiField["enum"].([]string)
	if len(multiEnum) != 2 {
		t.Errorf("multi_sig.enum = %v, want [ed25519 rsa]", multiEnum)
	}
}

func TestSign_RoundTrip(t *testing.T) {
	// Provider
	builder := NewSchemaBuilder()
	builder.SetActionSign("signature", "ed25519")
	builder.SetActionStringLength("message", "1", "100")

	// Export
	exported := builder.Build()
	props := exported["properties"].(map[string]any)

	// Consumer
	parsed := ParseSchema(props)

	// Valid action
	sae, err := NewAction("signedMessage", parsed).
		Set("message", "hello world").
		Set("signature", "base64sig==").
		Finalize()

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if sae == nil {
		t.Fatal("expected SAE bytes")
	}
}
