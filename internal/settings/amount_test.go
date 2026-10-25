package settings

import "testing"

func TestParseTokenAmount(t *testing.T) {
	cases := map[string]int64{"500000": 500000, "500k": 500000, "10m": 10000000, "1.5g": 1500000000}
	for input, want := range cases {
		got, err := ParseTokenAmount(input)
		if err != nil || got != want {
			t.Fatalf("ParseTokenAmount(%q) = %d, %v; want %d", input, got, err, want)
		}
	}
}

func TestParseTokenAmountRejectsInvalid(t *testing.T) {
	for _, input := range []string{"", "0", "-2m", "banana"} {
		if _, err := ParseTokenAmount(input); err == nil {
			t.Fatalf("expected %q to fail", input)
		}
	}
}
