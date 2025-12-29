package consumer

import (
	"errors"
	"fmt"
	"math/big"
)

type SAE struct {
	SDTO map[string]any `json:"sdto"`
}

type Schema struct {
	Properties map[string]Constraint `json:"properties"`
}

type Constraint struct {
	Type string   `json:"type"`          // string / number
	Min  *string  `json:"min,omitempty"` // semantic boundary
	Max  *string  `json:"max,omitempty"`
	Enum []string `json:"enum,omitempty"`
}

type Consumer struct {
	Schema Schema
}

func NewConsumer(schema Schema) *Consumer {
	return &Consumer{Schema: schema}
}

func (c *Consumer) Validate(sae SAE) error {
	for field, constraint := range c.Schema.Properties {

		value, exists := sae.SDTO[field]
		if !exists {
			// 暫時：缺欄位 = pass（required 之後再處理）
			continue
		}

		if err := validateValue(value, constraint); err != nil {
			return fmt.Errorf("field %q invalid: %w", field, err)
		}
	}

	return nil
}

func validateValue(value any, c Constraint) error {
	switch c.Type {

	case "string":
		return validateString(value, c)

	case "number":
		return validateNumber(value, c)

	default:
		return fmt.Errorf("unknown type %q", c.Type)
	}
}

func validateString(value any, c Constraint) error {
	v, ok := value.(string)
	if !ok {
		return errors.New("expected string")
	}

	// enum
	if len(c.Enum) > 0 {
		for _, allowed := range c.Enum {
			if v == allowed {
				return nil
			}
		}
		return fmt.Errorf("value %q not in enum", v)
	}

	// length boundary（用語意，不轉 int）
	if c.Min != nil {
		if len(v) < len(*c.Min) {
			return fmt.Errorf("string length < min")
		}
	}
	if c.Max != nil {
		if len(v) > len(*c.Max) {
			return fmt.Errorf("string length > max")
		}
	}

	return nil
}

func validateNumber(value any, c Constraint) error {
	var v float64

	switch n := value.(type) {
	case int:
		v = float64(n)
	case int64:
		v = float64(n)
	case float32:
		v = float64(n)
	case float64:
		v = n
	default:
		return errors.New("expected number")
	}

	// min / max 是語意邊界，用 consumer 解析
	if c.Min != nil {
		if !compareNumber(v, *c.Min, ">=") {
			return fmt.Errorf("number < min")
		}
	}
	if c.Max != nil {
		if !compareNumber(v, *c.Max, "<=") {
			return fmt.Errorf("number > max")
		}
	}

	return nil
}

func compareNumber(value float64, bound string, op string) bool {
	v := new(big.Rat).SetFloat64(value)

	b := new(big.Rat)
	if _, ok := b.SetString(bound); !ok {
		// schema 有問題 → consumer reject
		return false
	}

	switch op {
	case ">=":
		return v.Cmp(b) >= 0
	case "<=":
		return v.Cmp(b) <= 0
	default:
		return false
	}
}
