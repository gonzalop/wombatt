package cmd

import (
	"bytes"
	"io"
	"testing"
)

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

func TestPrintRegisters(t *testing.T) {
	var out bytes.Buffer
	data := []byte{0x0c, 0xcd, 0xff, 0x00, 0x00, 0x00, 0x00, 0x2a}
	if err := printRegisters(&out, "u16:Voltage:V:0.01,i8,u8:_,u32:Count", data); err != nil {
		t.Fatal(err)
	}
	if want := "Voltage: 32.77V\nField_1: -1\nCount: 42\n"; out.String() != want {
		t.Errorf("got %q, want %q", out.String(), want)
	}

	for _, format := range []string{"u16:a:b:c:d:e", "nope", "u32,u32,u32"} {
		if err := printRegisters(io.Discard, format, data); err == nil {
			t.Errorf("format %q: got no error, want one", format)
		}
	}
}
