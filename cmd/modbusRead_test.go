package cmd

import "testing"

func TestParseFieldType(t *testing.T) {
	for _, good := range []string{"u16", "[4]u16", "[10]byte"} {
		if _, err := parseFieldType(good); err != nil {
			t.Errorf("parseFieldType(%q): %v", good, err)
		}
	}
	for _, bad := range []string{"[4u16", "u64", "[x]u16"} {
		if _, err := parseFieldType(bad); err == nil {
			t.Errorf("parseFieldType(%q) succeeded; want error", bad)
		}
	}
}
