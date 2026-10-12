package jcs

import (
	"slices"
	"testing"
	"unicode/utf16"
)

func TestKeyComparisonMatchesUTF16CodeUnits(t *testing.T) {
	alphabet := []rune{0, 'a', '中', 0xd7ff, 0xe000, 0xff01, 0xffff, 0x10000, 0x103ff, 0x10400, 0x1f600, 0x1f601, 0x10ffff}
	keys := []string{""}
	for _, first := range alphabet {
		keys = append(keys, string(first))
		for _, second := range alphabet {
			keys = append(keys, string([]rune{first, second}))
		}
	}
	for _, a := range keys {
		for _, b := range keys {
			reference := slices.Compare(utf16.Encode([]rune(a)), utf16.Encode([]rune(b)))
			actual := compareUTF16(a, b)
			if (actual < 0) != (reference < 0) || (actual > 0) != (reference > 0) {
				t.Fatalf("UTF-16 comparison differs for %q and %q", a, b)
			}
		}
	}
}
