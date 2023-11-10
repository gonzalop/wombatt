package common

import (
	"testing"

	"golang.org/x/exp/maps"
)

type emptyStruct struct {
}

type oneFieldNoTags struct {
	Fld int
}

type oneField struct {
	Fld int16 `multiplier:"10"`
	_   int   // ignored
}

type oneFieldString struct {
	Fld string `multiplier:"10"`
	_   int    // ignored
}

type oneFieldValues struct {
	Fld string `values:"00:2zeros,01:zeroone"`
}

type oneFieldBadValues struct {
	Fld string `values:"00,01:zeroone"`
}

type OneFieldFlags struct {
	Fld uint8 `flags:"a,b,c,d,e,f,g,h"`
}

type OneFieldFlags2 struct {
	Fld uint16 `flags:"a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p"`
}

type OneFieldFlags3 struct {
	Fld uint32 `flags:"a,b,c,d,e,f,g,h,i,j,k,l,m,n,o,p,aa,ab,ac,ad,ae,af,ag,ah,ai,aj,ak,al,am,an,ao,ap"`
}

type oneFieldArray struct {
	Fld []uint16 `name:"name_%d" multiplier:"1.5"`
}

type embeddedStruct struct {
	OneFieldFlags
	Fld2 uint16 `name:"name" multiplier:"1.5"`
}

type byteArrayStruct struct {
	Fld [5]byte `type:"string"`
}

type int16ArrayStruct struct {
	Fld [3]int16 `multiplier:"5"`
}

type badMultiplier struct {
	Fld [3]int16 `multiplier:"a"`
}

type emptyMultiplier struct {
	Fld [3]int16 `multiplier:""`
}

type bitGroup struct {
	Fld string `bitgroups:"SCC OK|AC charging|SCC charging|Battery over voltage,Battery under voltage|Line loss|Load on|Configuration changed"`
}

func TestTraverseStruct(t *testing.T) {
	tests := []struct {
		st      any
		nfields int
		ntags   int
		tags    []string
		values  []any
	}{
		{st: &emptyStruct{}},
		{st: &oneFieldNoTags{}, nfields: 1, ntags: 5, tags: []string{"name", "unit", "desc", "dclass", "icon"}},
		{st: &oneField{Fld: 5}, nfields: 1, ntags: 5, values: []any{50.0}},
		{st: &oneFieldString{Fld: "5"}},
		{st: &oneFieldValues{Fld: "00"}, values: []any{"2zeros"}},
		{st: &oneFieldValues{Fld: "01"}, values: []any{"zeroone"}},
		{st: &oneFieldValues{Fld: "001"}, values: []any{"001"}},
		{st: &oneFieldBadValues{Fld: "00"}, values: []any{"00"}},
		{st: &OneFieldFlags{Fld: 0x03}, values: []any{"g, h"}},
		{st: &OneFieldFlags{Fld: 0}, values: []any{uint32(0)}},
		{st: &OneFieldFlags{Fld: 127}, values: []any{"b, c, d, e, f, g, h"}},
		{st: &OneFieldFlags{Fld: 16}, values: []any{"d"}},
		{st: &OneFieldFlags2{Fld: 16}, values: []any{"l"}},
		{st: &OneFieldFlags2{Fld: 0xc000}, values: []any{"a, b"}},
		{st: &OneFieldFlags3{Fld: 0xc000}, values: []any{"aa, ab"}},
		{st: &OneFieldFlags3{Fld: 0xa000c000}, values: []any{"a, c, aa, ab"}},
		{st: &oneFieldArray{Fld: []uint16{1, 2}}, values: []any{1.5, 3.0}},
		{st: &embeddedStruct{OneFieldFlags: OneFieldFlags{1}, Fld2: 2}, values: []any{"h", 3.0}, nfields: 2},
		{st: &byteArrayStruct{Fld: [5]byte{'H', 'e', 'l', 'l', 'o'}}, values: []any{"Hello"}},
		{st: &int16ArrayStruct{Fld: [3]int16{10, 15, 42}}, values: []any{50.0, 75.0, 210.0}},
		{st: &badMultiplier{Fld: [3]int16{1, 2, 3}}, nfields: 0},
		{st: &emptyMultiplier{Fld: [3]int16{1, 2, 3}}, values: []any{int16(1), int16(2), int16(3)}, nfields: 3},
		{st: &bitGroup{Fld: ""}},
		{st: &bitGroup{Fld: "0"}},
		{st: &bitGroup{Fld: "1100000"}, values: []any{"SCC OK, AC charging"}},
	}

	var fieldCount int
	var tagCount int
	var tags []string

	for ii, tt := range tests {
		fieldCount = 0
		tagCount = 0
		tags = []string{}

		cb := func(m map[string]string, val any) {
			if tt.values != nil {
				if len(tt.values) > fieldCount {
					if tt.values[fieldCount] != val {
						t.Errorf("#%d wrong value for %v: got %T '%v' ; want %T '%v'", ii, m["name"], val, val, tt.values[fieldCount], tt.values[fieldCount])
					}
				} else {
					t.Errorf("#%d wrong values array for the test. Fix it.", ii)
				}
			}
			fieldCount++
			tagCount += len(m)
			for _, tag := range tt.tags {
				if _, ok := m[tag]; !ok {
					t.Errorf("#%dunexpected tag: %s", ii, tag)
				}
			}
			tags = maps.Keys(m)
		}
		TraverseStruct(tt.st, cb)
		if tt.nfields > 0 && tt.nfields != fieldCount {
			t.Errorf("#%d number of fields is wrong: got %d; want %d", ii, fieldCount, tt.nfields)
		}
		if tt.ntags > 0 && tt.ntags != tagCount {
			t.Errorf("#%d number of tags is wrong: got %d; want %d (tags: %v)", ii, tagCount, tt.ntags, tags)
		}
	}
}
