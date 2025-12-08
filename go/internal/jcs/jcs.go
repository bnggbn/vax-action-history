package jcs

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"unicode/utf16"
	"unicode/utf8"
)

// CanonicalizeJSON 入口 1：從原始 JSON bytes 轉成 VAX-JCS bytes。
// 會先用 encoding/json 解成 interface{}，再走我們自己的 canonical 寫回去。
func CanonicalizeJSON(input []byte) ([]byte, error) {
	var v any

	dec := json.NewDecoder(bytes.NewReader(input))
	dec.UseNumber() // 優先拿到 json.Number，數字字面量不會立刻變 float64

	if err := dec.Decode(&v); err != nil {
		return nil, fmt.Errorf("decode json: %w", err)
	}

	return CanonicalizeValue(v)
}

// CanonicalizeValue 入口 2：直接接受已經建好的物件 (map / struct 轉 map 等)。
func CanonicalizeValue(v any) ([]byte, error) {
	var buf bytes.Buffer
	if err := writeCanonicalValue(&buf, v); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

// ======== 寫入各型別 ========

func writeCanonicalValue(buf *bytes.Buffer, v any) error {
	switch x := v.(type) {

	case nil:
		buf.WriteString("null")

	case bool:
		if x {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}

	case string:
		writeJSONString(buf, x)

	case json.Number:
		s, err := normalizeJSONNumber(x.String())
		if err != nil {
			return err
		}
		buf.WriteString(s)

	case float32:
		buf.WriteString(formatFloat(float64(x)))

	case float64:
		buf.WriteString(formatFloat(x))

	case int, int8, int16, int32, int64:
		buf.WriteString(strconv.FormatInt(toInt64(x), 10))

	case uint, uint8, uint16, uint32, uint64:
		buf.WriteString(strconv.FormatUint(toUint64(x), 10))

	case map[string]any:
		return writeCanonicalObject(buf, x)

	case []any:
		return writeCanonicalArray(buf, x)

	default:
		// 如果是 struct 等，要先在外面轉成 map 再丟進來，這裡就先當 error。
		return fmt.Errorf("unsupported type in canonical encoder: %T", v)
	}

	return nil
}

// ======== Object / Array ========

func writeCanonicalObject(buf *bytes.Buffer, m map[string]any) error {
	buf.WriteByte('{')

	if len(m) == 0 {
		buf.WriteByte('}')
		return nil
	}

	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)

	for i, k := range keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		writeJSONString(buf, k)
		buf.WriteByte(':')
		if err := writeCanonicalValue(buf, m[k]); err != nil {
			return err
		}
	}

	buf.WriteByte('}')
	return nil
}

func writeCanonicalArray(buf *bytes.Buffer, arr []any) error {
	buf.WriteByte('[')

	for i, elem := range arr {
		if i > 0 {
			buf.WriteByte(',')
		}
		if err := writeCanonicalValue(buf, elem); err != nil {
			return err
		}
	}

	buf.WriteByte(']')
	return nil
}

// ======== String 處理（ASCII-only + escape） ========

func writeJSONString(buf *bytes.Buffer, s string) {
	buf.WriteByte('"')

	for len(s) > 0 {
		r, size := utf8.DecodeRuneInString(s)
		s = s[size:]

		switch r {
		case '"':
			buf.WriteString(`\"`)
		case '\\':
			buf.WriteString(`\\`)
		case '\b':
			buf.WriteString(`\b`)
		case '\f':
			buf.WriteString(`\f`)
		case '\n':
			buf.WriteString(`\n`)
		case '\r':
			buf.WriteString(`\r`)
		case '\t':
			buf.WriteString(`\t`)
		default:
			if r < 0x20 {
				// 其他控制字元 → \u00XX
				buf.WriteString(`\u00`)
				buf.WriteString(hex2(uint8(r)))
			} else if r >= 0x20 && r <= 0x7E {
				// 可視 ASCII 直接寫
				buf.WriteRune(r)
			} else {
				// 非 ASCII → 轉 UTF-16，再依碼元輸出 \uXXXX（支援 surrogate pair）
				for _, u := range utf16.Encode([]rune{r}) {
					buf.WriteString(`\u`)
					buf.WriteString(hex4(u))
				}
			}
		}
	}

	buf.WriteByte('"')
}

func hex2(b uint8) string {
	const hexdigits = "0123456789abcdef"
	return string([]byte{
		hexdigits[(b>>4)&0x0f],
		hexdigits[b&0x0f],
	})
}

func hex4(u uint16) string {
	const hexdigits = "0123456789abcdef"
	return string([]byte{
		hexdigits[(u>>12)&0x0f],
		hexdigits[(u>>8)&0x0f],
		hexdigits[(u>>4)&0x0f],
		hexdigits[u&0x0f],
	})
}

// ======== Number 正規化（禁止科學記號、-0 → 0） ========
var decimalNumber = regexp.MustCompile(`^-?(0|[1-9][0-9]*)(\.[0-9]+)?$`)

func normalizeJSONNumber(raw string) (string, error) {
	// Step 1: Reject any non-decimal number
	if !decimalNumber.MatchString(raw) {
		return "", fmt.Errorf("non-decimal number not allowed: %s", raw)
	}

	// Step 1.5: Reject negative leading zero like -01 or -00
	if strings.HasPrefix(raw, "-0") && raw != "-0" && !strings.HasPrefix(raw, "-0.") {
		return "", fmt.Errorf("invalid leading zero: %s", raw)
	}

	// Step 2: Try signed integer
	if i, err := strconv.ParseInt(raw, 10, 64); err == nil {
		return strconv.FormatInt(i, 10), nil
	}

	// Step 3: Try unsigned integer
	if u, err := strconv.ParseUint(raw, 10, 64); err == nil {
		return strconv.FormatUint(u, 10), nil
	}

	// Step 4: Decimal float
	f, err := strconv.ParseFloat(raw, 64)
	if err != nil {
		return "", fmt.Errorf("invalid JSON number: %s", raw)
	}

	if math.IsNaN(f) || math.IsInf(f, 0) {
		return "", fmt.Errorf("NaN/Infinity not allowed: %s", raw)
	}

	// Step 5: Normalize -0 → 0
	if f == 0 {
		return "0", nil
	}

	return formatFloat(f), nil
}

func formatFloat(f float64) string {
	if f == 0 || math.IsNaN(f) {
		// NaN 在 JSON 裡本來就不合法；這裡先全視為 0，之後你可以改成直接報錯
		return "0"
	}
	if math.IsInf(f, 0) {
		// Infinity 同樣不允許，這裡先報錯比較好；不過為了簡單先 panic
		panic("Infinity is not allowed in VAX-JCS")
	}

	// 去掉 -0
	if f == 0 {
		return "0"
	}

	// 'f' + -1 → 十進位、不用科學記號
	s := strconv.FormatFloat(f, 'f', -1, 64)

	// 去掉 -0.000000 這種極小值誤差（可選）
	if s == "-0" {
		return "0"
	}

	return s
}

// ======== 小工具：整數轉換 ========

func toInt64(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	default:
		panic(fmt.Sprintf("toInt64 unsupported: %T", v))
	}
}

func toUint64(v any) uint64 {
	switch n := v.(type) {
	case uint:
		return uint64(n)
	case uint8:
		return uint64(n)
	case uint16:
		return uint64(n)
	case uint32:
		return uint64(n)
	case uint64:
		return n
	default:
		panic(fmt.Sprintf("toUint64 unsupported: %T", v))
	}
}

// Marshal is the public entrypoint for canonicalizing any Go value
// into VAX-JCS canonical JSON bytes.
//
// It first marshals using encoding/json (to turn structs into maps),
// then applies the VAX-JCS canonical rules.
func Marshal(v interface{}) ([]byte, error) {

	// Step 1: marshal using standard JSON (non-canonical)
	raw, err := json.Marshal(v)
	if err != nil {
		return nil, fmt.Errorf("json marshal: %w", err)
	}

	// Step 2: apply VAX-JCS canonicalization
	return CanonicalizeJSON(raw)
}
