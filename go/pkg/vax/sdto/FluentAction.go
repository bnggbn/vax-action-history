package sdto

import (
	"errors"
	"fmt"
	"math/big"
	"strconv"

	"vax/pkg/vax/sae"
)

// FluentAction 是你給 Consumer 的「量尺」
type FluentAction struct {
	actionType string
	schema     map[string]FieldSpec // 從後端拉回來的驗證規則
	data       map[string]any
	errs       []error
}

func NewAction(actionType string, rules map[string]FieldSpec) *FluentAction {
	return &FluentAction{
		actionType: actionType,
		schema:     rules,
		data:       make(map[string]any),
	}
}

// Set 在賦值的瞬間進行驗證
func (f *FluentAction) Set(key string, value any) *FluentAction {
	spec, exists := f.schema[key]
	if !exists {
		f.errs = append(f.errs, fmt.Errorf("unknown field: %s", key))
		return f
	}

	if err := validateValue(value, spec); err != nil {
		f.errs = append(f.errs, fmt.Errorf("field %s: %w", key, err))
		return f
	}

	f.data[key] = value
	return f
}

func validateValue(value any, c FieldSpec) error {
	switch c.Type {
	case "string":
		return validateString(value, c)
	case "number":
		return validateNumber(value, c)
	case "sign":
		return validateSign(value, c)
	default:
		return fmt.Errorf("unknown type %q", c.Type)
	}
}

func validateSign(value any, c FieldSpec) error {
	// 簽名值只能是 string（類型已在 schema 層定義）
	v, ok := value.(string)
	if !ok {
		return errors.New("sign field expects string value")
	}

	// 可擴展：根據 c.Enum[0] 做格式驗證（hex/base64 等）
	if len(v) == 0 {
		return errors.New("sign value cannot be empty")
	}

	return nil
}

func validateString(value any, c FieldSpec) error {
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

	// length boundary (數值解析)
	if c.Min != nil {
		minLen, err := strconv.Atoi(*c.Min)
		if err == nil && len(v) < minLen {
			return fmt.Errorf("string length %d < min %d", len(v), minLen)
		}
	}
	if c.Max != nil {
		maxLen, err := strconv.Atoi(*c.Max)
		if err == nil && len(v) > maxLen {
			return fmt.Errorf("string length %d > max %d", len(v), maxLen)
		}
	}

	return nil
}

func validateNumber(value any, c FieldSpec) error {
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

// Finalize 最終產出 SAE
func (f *FluentAction) Finalize() ([]byte, error) {
	// Check for missing required fields (all schema fields are required)
	for key := range f.schema {
		if _, exists := f.data[key]; !exists {
			f.errs = append(f.errs, fmt.Errorf("missing required field: %s", key))
		}
	}

	if len(f.errs) > 0 {
		// Aggregate errors into a single message
		msg := ""
		for i, err := range f.errs {
			if i > 0 {
				msg += "; "
			}
			msg += err.Error()
		}
		return nil, errors.New(msg)
	}
	// 調用你剛剛寫好的 SAE.BuildSAE
	return sae.BuildSAE(f.actionType, f.data)
}

// ValidateData validates a map against schema (for server-side verification)
func ValidateData(data map[string]any, schema map[string]FieldSpec) error {
	var errs []error

	// Check all required fields in schema exist
	for key, spec := range schema {
		value, exists := data[key]
		if !exists {
			errs = append(errs, fmt.Errorf("missing field: %s", key))
			continue
		}
		if err := validateValue(value, spec); err != nil {
			errs = append(errs, fmt.Errorf("field %s: %w", key, err))
		}
	}

	// Check no extra fields
	for key := range data {
		if _, exists := schema[key]; !exists {
			errs = append(errs, fmt.Errorf("unknown field: %s", key))
		}
	}

	if len(errs) > 0 {
		// Aggregate errors into a single message
		msg := ""
		for i, err := range errs {
			if i > 0 {
				msg += "; "
			}
			msg += err.Error()
		}
		return errors.New(msg)
	}
	return nil
}
