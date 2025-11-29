package jcs

import (
	"bytes"
	"math"
	"testing"
)

func TestMarshal(t *testing.T) {
	tests := []struct {
		name    string
		input   interface{}
		want    string
		wantErr bool
	}{
		// === 基本型別 ===
		{
			name:  "nil",
			input: nil,
			want:  `null`,
		},
		{
			name:  "bool true",
			input: true,
			want:  `true`,
		},
		{
			name:  "bool false",
			input: false,
			want:  `false`,
		},
		{
			name:  "int",
			input: 123,
			want:  `123`,
		},
		{
			name:  "negative int",
			input: -456,
			want:  `-456`,
		},
		{
			name:  "float64",
			input: 123.456,
			want:  `123.456`,
		},
		{
			name:  "string",
			input: "hello",
			want:  `"hello"`,
		},

		// === Slice / Array ===
		{
			name:  "empty slice",
			input: []int{},
			want:  `[]`,
		},
		{
			name:  "int slice",
			input: []int{1, 2, 3},
			want:  `[1,2,3]`,
		},
		{
			name:  "string slice",
			input: []string{"a", "b", "c"},
			want:  `["a","b","c"]`,
		},
		{
			name:  "mixed slice",
			input: []interface{}{1, "two", true, nil},
			want:  `[1,"two",true,null]`,
		},

		// === Map ===
		{
			name:  "empty map",
			input: map[string]interface{}{},
			want:  `{}`,
		},
		{
			name: "simple map",
			input: map[string]interface{}{
				"name": "Alice",
				"age":  30,
			},
			want: `{"age":30,"name":"Alice"}`, // 注意：keys 按字母排序
		},
		{
			name: "map with unsorted keys",
			input: map[string]interface{}{
				"z": 1,
				"a": 2,
				"m": 3,
			},
			want: `{"a":2,"m":3,"z":1}`, // 排序後
		},

		// === Struct ===
		{
			name: "simple struct",
			input: struct {
				Name string `json:"name"`
				Age  int    `json:"age"`
			}{
				Name: "Bob",
				Age:  25,
			},
			want: `{"age":25,"name":"Bob"}`, // 字段按字母排序
		},
		{
			name: "struct with omitempty",
			input: struct {
				Name  string `json:"name"`
				Email string `json:"email,omitempty"`
			}{
				Name: "Charlie",
			},
			want: `{"name":"Charlie"}`, // email 被省略
		},
		{
			name: "struct with unexported field",
			input: struct {
				Name   string `json:"name"`
				secret string // 不會被序列化
			}{
				Name:   "Dave",
				secret: "hidden",
			},
			want: `{"name":"Dave"}`,
		},
		{
			name: "nested struct",
			input: struct {
				User struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				} `json:"user"`
			}{
				User: struct {
					Name string `json:"name"`
					Age  int    `json:"age"`
				}{
					Name: "Eve",
					Age:  28,
				},
			},
			want: `{"user":{"age":28,"name":"Eve"}}`,
		},

		// === 巢狀結構 ===
		{
			name: "map with nested slice",
			input: map[string]interface{}{
				"items": []int{3, 1, 2},
				"name":  "test",
			},
			want: `{"items":[3,1,2],"name":"test"}`,
		},
		{
			name: "slice with nested map",
			input: []interface{}{
				map[string]interface{}{
					"z": 1,
					"a": 2,
				},
				map[string]interface{}{
					"y": 3,
					"b": 4,
				},
			},
			want: `[{"a":2,"z":1},{"b":4,"y":3}]`, // 每個 map 的 keys 都排序
		},

		// === 特殊字元 ===
		{
			name:  "string with quotes",
			input: `hello"world`,
			want:  `"hello\"world"`,
		},
		{
			name:  "string with backslash",
			input: `path\to\file`,
			want:  `"path\\to\\file"`,
		},
		{
			name:  "string with newline",
			input: "line1\nline2",
			want:  `"line1\nline2"`,
		},
		{
			name:  "string with unicode",
			input: "你好",
			want:  `"\u4f60\u597d"`,
		},

		// === 數字正規化 ===
		{
			name:  "negative zero float",
			input: -0.0,
			want:  `0`,
		},
		{
			name: "struct with negative zero",
			input: struct {
				Value float64 `json:"value"`
			}{
				Value: -0.0,
			},
			want: `{"value":0}`,
		},

		// === 複雜實例 ===
		{
			name: "realistic user object",
			input: struct {
				ID       int      `json:"id"`
				Username string   `json:"username"`
				Email    string   `json:"email"`
				Tags     []string `json:"tags"`
				Active   bool     `json:"active"`
			}{
				ID:       12345,
				Username: "alice",
				Email:    "alice@example.com",
				Tags:     []string{"admin", "developer"},
				Active:   true,
			},
			want: `{"active":true,"email":"alice@example.com","id":12345,"tags":["admin","developer"],"username":"alice"}`,
		},

		// === 指標 ===（Marshal 會自動解引用或輸出 null）
		{
			name: "pointer to int",
			input: func() *int {
				i := 42
				return &i
			}(),
			want: `42`,
		},
		{
			name:  "nil pointer",
			input: (*int)(nil),
			want:  `null`,
		},
		{
			name: "struct with pointer field",
			input: struct {
				Name  string  `json:"name"`
				Age   *int    `json:"age,omitempty"`
				Email *string `json:"email,omitempty"`
			}{
				Name: "Frank",
				Age: func() *int {
					i := 30
					return &i
				}(),
				Email: nil,
			},
			want: `{"age":30,"name":"Frank"}`,
		},

		// === 應該失敗的情況（Marshal 本身就不支援） ===
		{
			name:    "channel (unsupported)",
			input:   make(chan int),
			wantErr: true,
		},
		{
			name:    "function (unsupported)",
			input:   func() {},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Marshal(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("Marshal() expected error but got none")
				}
				return
			}

			if err != nil {
				t.Errorf("Marshal() unexpected error: %v", err)
				return
			}

			if string(got) != tt.want {
				t.Errorf("Marshal()\ngot:  %s\nwant: %s", string(got), tt.want)
			}
		})
	}
}

// Benchmark 測試（可選）
func BenchmarkMarshal(b *testing.B) {
	data := struct {
		ID       int                    `json:"id"`
		Username string                 `json:"username"`
		Tags     []string               `json:"tags"`
		Meta     map[string]interface{} `json:"meta"`
	}{
		ID:       12345,
		Username: "testuser",
		Tags:     []string{"tag1", "tag2", "tag3"},
		Meta: map[string]interface{}{
			"z": 100,
			"a": 200,
			"m": 300,
		},
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_, err := Marshal(data)
		if err != nil {
			b.Fatal(err)
		}
	}
}

func TestNormalizeJSONNumber(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// === 整數 ===
		{
			name:  "positive integer",
			input: "123",
			want:  "123",
		},
		{
			name:  "negative integer",
			input: "-456",
			want:  "-456",
		},
		{
			name:  "zero",
			input: "0",
			want:  "0",
		},
		{
			name:  "negative zero integer",
			input: "-0",
			want:  "0", // 正規化為 0
		},
		{
			name:  "large positive integer",
			input: "9007199254740991", // 2^53 - 1
			want:  "9007199254740991",
		},
		{
			name:  "large negative integer",
			input: "-9007199254740991",
			want:  "-9007199254740991",
		},

		// === 小數 ===
		{
			name:  "positive decimal",
			input: "123.456",
			want:  "123.456",
		},
		{
			name:  "negative decimal",
			input: "-123.456",
			want:  "-123.456",
		},
		{
			name:  "decimal starting with zero",
			input: "0.5",
			want:  "0.5",
		},
		{
			name:  "negative decimal starting with zero",
			input: "-0.5",
			want:  "-0.5",
		},
		{
			name:  "small decimal",
			input: "0.0001",
			want:  "0.0001",
		},
		{
			name:  "negative zero decimal",
			input: "-0.0",
			want:  "0", // 正規化為 0
		},

		// === 應該拒絕：科學記號 ===
		{
			name:    "scientific notation lowercase",
			input:   "1e10",
			wantErr: true,
		},
		{
			name:    "scientific notation uppercase",
			input:   "1E10",
			wantErr: true,
		},
		{
			name:    "scientific notation negative exponent",
			input:   "1.5e-3",
			wantErr: true,
		},
		{
			name:    "scientific notation positive exponent",
			input:   "2.5e+2",
			wantErr: true,
		},

		// === 應該拒絕：前導零 ===
		{
			name:    "positive leading zero",
			input:   "01",
			wantErr: true,
		},
		{
			name:    "negative leading zero",
			input:   "-01",
			wantErr: true,
		},
		{
			name:    "multiple leading zeros",
			input:   "00",
			wantErr: true,
		},
		{
			name:    "negative multiple leading zeros",
			input:   "-00",
			wantErr: true,
		},
		{
			name:    "leading zero with decimal",
			input:   "01.5",
			wantErr: true,
		},

		// === 應該拒絕：非法格式 ===
		{
			name:    "trailing dot",
			input:   "1.",
			wantErr: true,
		},
		{
			name:    "leading dot",
			input:   ".5",
			wantErr: true,
		},
		{
			name:    "multiple dots",
			input:   "1.2.3",
			wantErr: true,
		},
		{
			name:    "empty string",
			input:   "",
			wantErr: true,
		},
		{
			name:    "only minus sign",
			input:   "-",
			wantErr: true,
		},
		{
			name:    "letters",
			input:   "abc",
			wantErr: true,
		},
		{
			name:    "NaN string",
			input:   "NaN",
			wantErr: true,
		},
		{
			name:    "Infinity string",
			input:   "Infinity",
			wantErr: true,
		},
		{
			name:    "hex notation",
			input:   "0x1F",
			wantErr: true,
		},

		// === 邊界情況 ===
		{
			name:  "very small positive decimal",
			input: "0.000000000001",
			want:  "0.000000000001",
		},
		{
			name:  "very small negative decimal",
			input: "-0.000000000001",
			want:  "-0.000000000001",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := normalizeJSONNumber(tt.input)

			if tt.wantErr {
				if err == nil {
					t.Errorf("normalizeJSONNumber(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("normalizeJSONNumber(%q) unexpected error: %v", tt.input, err)
				return
			}

			if got != tt.want {
				t.Errorf("normalizeJSONNumber(%q) = %q, want %q", tt.input, got, tt.want)
			}
		})
	}
}

func TestCanonicalizeJSON(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		// === 基本型別 ===
		{
			name:  "null",
			input: `null`,
			want:  `null`,
		},
		{
			name:  "true",
			input: `true`,
			want:  `true`,
		},
		{
			name:  "false",
			input: `false`,
			want:  `false`,
		},
		{
			name:  "integer",
			input: `123`,
			want:  `123`,
		},
		{
			name:  "negative integer",
			input: `-456`,
			want:  `-456`,
		},
		{
			name:  "decimal",
			input: `123.456`,
			want:  `123.456`,
		},
		{
			name:  "string",
			input: `"hello"`,
			want:  `"hello"`,
		},

		// === 空結構 ===
		{
			name:  "empty object",
			input: `{}`,
			want:  `{}`,
		},
		{
			name:  "empty array",
			input: `[]`,
			want:  `[]`,
		},

		// === Object key 排序 ===
		{
			name:  "object keys sorted",
			input: `{"z":1,"a":2,"m":3}`,
			want:  `{"a":2,"m":3,"z":1}`,
		},
		{
			name:  "nested object keys sorted",
			input: `{"outer":{"z":1,"a":2}}`,
			want:  `{"outer":{"a":2,"z":1}}`,
		},

		// === 移除空白 ===
		{
			name:  "remove whitespace in object",
			input: `{ "a" : 1 , "b" : 2 }`,
			want:  `{"a":1,"b":2}`,
		},
		{
			name:  "remove whitespace in array",
			input: `[ 1 , 2 , 3 ]`,
			want:  `[1,2,3]`,
		},
		{
			name: "remove newlines and tabs",
			input: `{
				"a": 1,
				"b": 2
			}`,
			want: `{"a":1,"b":2}`,
		},

		// === 字串轉義 ===
		{
			name:  "escape quote",
			input: `"hello\"world"`,
			want:  `"hello\"world"`,
		},
		{
			name:  "escape backslash",
			input: `"path\\to\\file"`,
			want:  `"path\\to\\file"`,
		},
		{
			name:  "escape newline",
			input: `"line1\nline2"`,
			want:  `"line1\nline2"`,
		},
		{
			name:  "escape tab",
			input: `"col1\tcol2"`,
			want:  `"col1\tcol2"`,
		},

		// === Unicode 處理 ===
		{
			name:  "unicode escape",
			input: `"你好"`,
			want:  `"\u4f60\u597d"`, // 中文轉 \uXXXX
		},
		{
			name:  "emoji",
			input: `"😀"`,
			want:  `"\ud83d\ude00"`, // emoji 用 surrogate pair
		},
		{
			name:  "mixed ascii and unicode",
			input: `"Hello世界"`,
			want:  `"Hello\u4e16\u754c"`,
		},

		// === 數字正規化 ===
		{
			name:  "negative zero",
			input: `-0`,
			want:  `0`,
		},
		{
			name:  "negative zero decimal",
			input: `-0.0`,
			want:  `0`,
		},

		// === 巢狀結構 ===
		{
			name:  "nested object and array",
			input: `{"a":{"b":[1,2,3]},"c":null}`,
			want:  `{"a":{"b":[1,2,3]},"c":null}`,
		},
		{
			name:  "complex nested",
			input: `{"z":[{"y":2,"x":1}],"a":{"c":3,"b":4}}`,
			want:  `{"a":{"b":4,"c":3},"z":[{"x":1,"y":2}]}`,
		},

		// === 應該拒絕：科學記號 ===
		{
			name:    "scientific notation",
			input:   `1e10`,
			wantErr: true,
		},
		{
			name:    "scientific notation in object",
			input:   `{"a":1e10}`,
			wantErr: true,
		},

		// === 應該拒絕：非法 JSON ===
		{
			name:    "invalid json",
			input:   `{invalid}`,
			wantErr: true,
		},
		{
			name:    "trailing comma",
			input:   `{"a":1,}`,
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := CanonicalizeJSON([]byte(tt.input))

			if tt.wantErr {
				if err == nil {
					t.Errorf("CanonicalizeJSON(%q) expected error but got none", tt.input)
				}
				return
			}

			if err != nil {
				t.Errorf("CanonicalizeJSON(%q) unexpected error: %v", tt.input, err)
				return
			}

			if string(got) != tt.want {
				t.Errorf("CanonicalizeJSON(%q)\ngot:  %q\nwant: %q", tt.input, string(got), tt.want)
			}
		})
	}
}

// ===== 下面是我幫你加的「補 coverage」的小測試 =====

// 直接測 CanonicalizeValue 入口
func TestCanonicalizeValue_Object(t *testing.T) {
	input := map[string]interface{}{
		"b": 2,
		"a": 1,
	}

	got, err := CanonicalizeValue(input)
	if err != nil {
		t.Fatalf("CanonicalizeValue error: %v", err)
	}

	want := `{"a":1,"b":2}`
	if string(got) != want {
		t.Errorf("CanonicalizeValue()\ngot:  %s\nwant: %s", string(got), want)
	}
}

// 測試 writeCanonicalValue 遇到不支援型別時會回傳 error
func TestWriteCanonicalValue_UnsupportedType(t *testing.T) {
	var buf bytes.Buffer

	unsupported := []interface{}{
		make(chan int),
		func() {},
	}

	for _, v := range unsupported {
		if err := writeCanonicalValue(&buf, v); err == nil {
			t.Errorf("writeCanonicalValue(%T) expected error, got nil", v)
		}
	}
}

// 測 formatFloat 的 NaN / Infinity / -0 路徑
func TestFormatFloat_NaNAndInfinity(t *testing.T) {
	if got := formatFloat(math.NaN()); got != "0" {
		t.Errorf("formatFloat(NaN) = %q, want %q", got, "0")
	}

	// Infinity 預期 panic
	tests := []struct {
		name string
		in   float64
	}{
		{"+Inf", math.Inf(1)},
		{"-Inf", math.Inf(-1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defer func() {
				if r := recover(); r == nil {
					t.Errorf("formatFloat(%s) expected panic, got none", tt.name)
				}
			}()
			_ = formatFloat(tt.in)
		})
	}

	// -0 → 0
	if got := formatFloat(math.Copysign(0, -1)); got != "0" {
		t.Errorf("formatFloat(-0) = %q, want %q", got, "0")
	}
}

// 測 toInt64 / toUint64 所有合法型別 + panic 分支
func TestToInt64_AllIntTypes(t *testing.T) {
	values := []interface{}{
		int(1),
		int8(2),
		int16(3),
		int32(4),
		int64(5),
	}

	for _, v := range values {
		_ = toInt64(v) // 只要不 panic 即可
	}
}

func TestToInt64_PanicOnUnsupported(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("toInt64 should panic on unsupported type")
		}
	}()
	toInt64("not-int")
}

func TestToUint64_AllUintTypes(t *testing.T) {
	values := []interface{}{
		uint(1),
		uint8(2),
		uint16(3),
		uint32(4),
		uint64(5),
	}

	for _, v := range values {
		_ = toUint64(v)
	}
}

func TestToUint64_PanicOnUnsupported(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Errorf("toUint64 should panic on unsupported type")
		}
	}()
	toUint64("not-uint")
}
