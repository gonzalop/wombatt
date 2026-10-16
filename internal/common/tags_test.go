package common

import (
	"go/ast"
	"go/parser"
	"go/token"
	"io/fs"
	"path/filepath"
	"reflect"
	"strconv"
	"strings"
	"testing"
)

// TestStructTags checks every values and flags tag under internal/. Malformed
// tags don't fail at runtime: values silently fall back to the raw number and
// flags get attached to the wrong bits.
func TestStructTags(t *testing.T) {
	bits := map[string]int{"uint8": 8, "uint16": 16, "uint32": 32}
	fset := token.NewFileSet()
	err := filepath.WalkDir("..", func(path string, d fs.DirEntry, err error) error {
		if err != nil || d.IsDir() || !strings.HasSuffix(path, ".go") || strings.HasSuffix(path, "_test.go") {
			return err
		}
		f, err := parser.ParseFile(fset, path, nil, 0)
		if err != nil {
			return err
		}
		ast.Inspect(f, func(n ast.Node) bool {
			field, ok := n.(*ast.Field)
			if !ok || field.Tag == nil {
				return true
			}
			tag := reflect.StructTag(strings.Trim(field.Tag.Value, "`"))
			pos := fset.Position(field.Pos())

			ident, _ := field.Type.(*ast.Ident)
			isInt := ident != nil && strings.Contains(ident.Name, "int")
			if values := tag.Get("values"); values != "" {
				for kv := range strings.SplitSeq(values, ",") {
					k, _, ok := strings.Cut(kv, ":")
					k = strings.TrimSpace(k)
					// Keys are matched against fmt's %v of the field, so integer fields need canonical decimal keys.
					n, err := strconv.Atoi(k)
					if !ok || isInt && (err != nil || strconv.Itoa(n) != k) {
						t.Errorf("%s: bad values entry %q: want <decimal>:<name>, comma-separated", pos, kv)
					}
				}
			}

			if flags := tag.Get("flags"); flags != "" {
				typ := field.Type
				if arr, ok := typ.(*ast.ArrayType); ok {
					typ = arr.Elt
				}
				ident, _ = typ.(*ast.Ident)
				if ident == nil || bits[ident.Name] == 0 {
					t.Errorf("%s: flags tag on unsupported type", pos)
				} else if got := len(strings.Split(flags, ",")); got != bits[ident.Name] {
					t.Errorf("%s: flags has %d entries, want one per bit of %s (%d)", pos, got, ident.Name, bits[ident.Name])
				}
			}
			return true
		})
		return nil
	})
	if err != nil {
		t.Fatal(err)
	}
}
