package schema

import (
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
)

type SchemaGenerator struct{}

func Generate[T any]() map[string]interface{} {
	gen := &SchemaGenerator{}
	var zero T
	return gen.generate(reflect.TypeOf(zero))
}

func (g *SchemaGenerator) generate(t reflect.Type) map[string]interface{} {
	schema := map[string]interface{}{
		"$schema": "http://json-schema.org/draft-07/schema#",
		"type":    "object",
	}

	schema["properties"] = g.generateProperties(t)

	if req := g.getRequiredFields(t); len(req) > 0 {
		schema["required"] = req
	}

	return schema
}

/* ---------------- core ---------------- */

func (g *SchemaGenerator) generateProperties(t reflect.Type) map[string]interface{} {
	props := make(map[string]interface{})

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		name := strings.Split(jsonTag, ",")[0]
		validateTag := field.Tag.Get("validate")

		props[name] = g.schemaForField(field, validateTag)
	}

	return props
}

func (g *SchemaGenerator) schemaForField(field reflect.StructField, validateTag string,) map[string]interface{} {

	fieldType := field.Type
	isPointer := fieldType.Kind() == reflect.Ptr
	if isPointer {
		fieldType = fieldType.Elem()
	}

	schema := g.schemaForType(fieldType)

	if validateTag != "" {
		g.applyValidation(schema, validateTag)
	}

	if desc := field.Tag.Get("description"); desc != "" {
		schema["description"] = desc
	}

	if ex := field.Tag.Get("example"); ex != "" {
		schema["example"] = ex
	}

	return schema
}

/* ---------------- type mapping ---------------- */

func (g *SchemaGenerator) schemaForType(t reflect.Type) map[string]interface{} {
	switch t.Kind() {

	case reflect.Struct:
		s := map[string]interface{}{
			"type":       "object",
			"properties": g.generateProperties(t),
		}
		if req := g.getRequiredFields(t); len(req) > 0 {
			s["required"] = req
		}
		return s

	case reflect.Slice, reflect.Array:
		return map[string]interface{}{
			"type":  "array",
			"items": g.schemaForType(t.Elem()),
		}

	default:
		return map[string]interface{}{
			"type": g.goTypeToJSONType(t),
		}
	}
}

func (g *SchemaGenerator) goTypeToJSONType(t reflect.Type) string {
	switch t.Kind() {
	case reflect.String:
		return "string"
	case reflect.Int, reflect.Int8, reflect.Int16,
		reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8,
		reflect.Uint16, reflect.Uint32,
		reflect.Uint64:
		return "integer"
	case reflect.Float32, reflect.Float64:
		return "number"
	case reflect.Bool:
		return "boolean"
	default:
		return "string"
	}
}

/* ---------------- validation ---------------- */

func (g *SchemaGenerator) applyValidation(schema map[string]interface{}, validateTag string,) {
	rules := strings.Split(validateTag, ",")
	typ, _ := schema["type"].(string)

	for _, rule := range rules {
		rule = strings.TrimSpace(rule)

		// ---------- string ----------
		if typ == "string" {
			if strings.HasPrefix(rule, "min=") {
				if v, err := strconv.Atoi(strings.TrimPrefix(rule, "min=")); err == nil {
					schema["minLength"] = v
				}
			}
			if strings.HasPrefix(rule, "max=") {
				if v, err := strconv.Atoi(strings.TrimPrefix(rule, "max=")); err == nil {
					schema["maxLength"] = v
				}
			}

			switch rule {
			case "email":
				schema["format"] = "email"
			case "url":
				schema["format"] = "uri"
			case "uuid":
				schema["format"] = "uuid"
			case "datetime":
				schema["format"] = "date-time"
			}
		}

		// ---------- number ----------
		if typ == "number" || typ == "integer" {
			if strings.HasPrefix(rule, "gte=") {
				if v, err := strconv.ParseFloat(strings.TrimPrefix(rule, "gte="), 64); err == nil {
					schema["minimum"] = v
				}
			}
			if strings.HasPrefix(rule, "lte=") {
				if v, err := strconv.ParseFloat(strings.TrimPrefix(rule, "lte="), 64); err == nil {
					schema["maximum"] = v
				}
			}
		}

		// ---------- enum ----------
		if strings.HasPrefix(rule, "oneof=") {
			values := strings.Split(strings.TrimPrefix(rule, "oneof="), " ")
			schema["enum"] = values
		}

		// ---------- nullable ----------
		if rule == "nullable" {
			if t, ok := schema["type"].(string); ok {
				schema["type"] = []string{t, "null"}
			}
		}
	}
}

/* ---------------- required ---------------- */

func (g *SchemaGenerator) getRequiredFields(t reflect.Type) []string {
	var required []string

	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)

		jsonTag := field.Tag.Get("json")
		if jsonTag == "" || jsonTag == "-" {
			continue
		}

		name := strings.Split(jsonTag, ",")[0]
		hasOmitempty := strings.Contains(jsonTag, "omitempty")

		validateTag := field.Tag.Get("validate")
		hasRequired := strings.Contains(validateTag, "required")

		isPointer := field.Type.Kind() == reflect.Ptr

		if hasRequired {
			required = append(required, name)
			continue
		}

		if !isPointer && !hasOmitempty {
			required = append(required, name)
		}
	}

	return required
}

/* ---------------- output ---------------- */

func ToJSON(schema map[string]interface{}) (string, error) {
	b, err := json.MarshalIndent(schema, "", "  ")
	if err != nil {
		return "", err
	}
	return string(b), nil
}
