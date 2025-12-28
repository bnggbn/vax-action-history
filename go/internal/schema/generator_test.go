// internal/schema/generator_test.go
package schema_test

import (
    "testing"
    "vax/internal/schema"
)

func TestGenerate_Simple(t *testing.T) {
    type SimpleDTO struct {
        Name string `json:"name" validate:"required,max=50"`
        Age  int    `json:"age" validate:"gte=0,lte=150"`
    }

    s := schema.Generate[SimpleDTO]()

    if s["type"] != "object" {
        t.Error("Expected type object")
    }

    props := s["properties"].(map[string]interface{})

    nameSchema := props["name"].(map[string]interface{})
    if nameSchema["type"] != "string" {
        t.Error("Expected name to be string")
    }
    if nameSchema["maxLength"] != 50 {
        t.Error("Expected maxLength 50")
    }

    required := s["required"].([]string)
    if len(required) != 2 || required[0] != "name" {
        t.Error("Expected name to be required")
    }
}

func TestGenerate_Nested(t *testing.T) {
    type Address struct {
        City string `json:"city"`
        ZIP  string `json:"zip"`
    }

    type UserDTO struct {
        Name    string  `json:"name"`
        Address Address `json:"address"`
    }

    s := schema.Generate[UserDTO]()

    props := s["properties"].(map[string]interface{})
    addressSchema := props["address"].(map[string]interface{})

    if addressSchema["type"] != "object" {
        t.Error("Expected address to be object")
    }

    addressProps := addressSchema["properties"].(map[string]interface{})
    if _, ok := addressProps["city"]; !ok {
        t.Error("Expected city in address properties")
    }
}

func TestGenerate_Array(t *testing.T) {
    type ListDTO struct {
        Tags []string `json:"tags" validate:"max=10"`
    }

    s := schema.Generate[ListDTO]()

    props := s["properties"].(map[string]interface{})
    tagsSchema := props["tags"].(map[string]interface{})

    if tagsSchema["type"] != "array" {
        t.Error("Expected tags to be array")
    }

    items := tagsSchema["items"].(map[string]interface{})
    if items["type"] != "string" {
        t.Error("Expected array items to be string")
    }
}

func TestToJSON(t *testing.T) {
    type SimpleDTO struct {
        Name string `json:"name"`
    }

    s := schema.Generate[SimpleDTO]()
    json, err := schema.ToJSON(s)

    if err != nil {
        t.Error("ToJSON failed:", err)
    }

    if json == "" {
        t.Error("Expected non-empty JSON")
    }
}
