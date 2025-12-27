// Package common has utility functions used by different commands.
package common

import (
	"fmt"
	"io"
	"log/slog"
	"math"
	"reflect"
	"strconv"
	"strings"
)

// TraverseStructCallback will be called for every field in structs passed to TraverseStruct.
type TraverseStructCallback func(map[string]string, any)

// TraverseStruct inspects the struct or pointer to struct passed as argument and for every
// field it will call cb with tags from the field and its value.
func TraverseStruct(data any, cb TraverseStructCallback) {
	stValue := reflect.ValueOf(data)
	if reflect.TypeOf(data).Kind() == reflect.Pointer {
		stValue = stValue.Elem()
	}
	stType := stValue.Type()
	nfields := stType.NumField()
	for i := range nfields {
		f := stType.Field(i)
		if f.Type.Kind() == reflect.Struct {
			TraverseStruct(stValue.Field(i).Interface(), cb)
			continue
		}
		v := stValue.Field(i)
		processField(f, v, cb)
	}
}

func processField(f reflect.StructField, v reflect.Value, cb TraverseStructCallback) {
	if f.Name == "_" || f.Tag.Get("skip") != "" {
		return
	}
	info := getFieldInfo(f)
	if f.Type.Kind() == reflect.Array {
		handleArrayField(f, v, info, f.Tag.Get("multiplier"), info["name"], cb)
		return
	}
	if mult := f.Tag.Get("multiplier"); mult != "" {
		newVal, err := handleMultiplier(mult, v)
		if err != nil {
			slog.Error("error converting multiplier", "error", err, "value", v, "multiplier", mult)
			return
		}
		cb(info, newVal)
		return
	}
	cb(info, getProcessedValue(f, v))
}

func getFieldInfo(f reflect.StructField) map[string]string {
	desc := f.Tag.Get("desc")
	if desc == "" {
		desc = f.Name
	}
	info := make(map[string]string)
	info["name"] = f.Tag.Get("name")
	info["unit"] = f.Tag.Get("unit")
	info["desc"] = desc
	info["icon"] = f.Tag.Get("icon")
	info["dclass"] = f.Tag.Get("dclass")
	info["precision"] = f.Tag.Get("precision")
	return info
}

func getProcessedValue(f reflect.StructField, v reflect.Value) any {
	val := v.Interface()
	vals := f.Tag.Get("values")
	if vals != "" {
		val = fmt.Sprintf("%v", v.Interface())
		m := parseValues(vals)
		if v, ok := m[val.(string)]; ok {
			val = v
		}
	}
	flags := f.Tag.Get("flags")
	if flags != "" && (f.Type.Name() == "uint8" || f.Type.Name() == "uint16" || f.Type.Name() == "uint32") {
		u32 := v.Convert(reflect.TypeOf(uint32(0))).Interface().(uint32)
		val = handleFlagsTag(flags, u32)
		if val == "" {
			val = u32
		}
	}
	bgroups := f.Tag.Get("bitgroups")
	if bgroups != "" {
		str := v.String()
		if str != "" {
			val = handleBitgroupsTag(bgroups, str)
		}
	}
	return val
}

func handleArrayField(f reflect.StructField, v reflect.Value, info map[string]string, mult, name string, cb TraverseStructCallback) {
	if f.Tag.Get("type") == "string" {
		cb(info, string(v.Bytes()))
		return
	}
	aValue := reflect.ValueOf(v.Interface()) // Use v.Interface() to get the underlying array
	for k := range aValue.Len() {
		newVal, err := handleMultiplier(mult, aValue.Index(k))
		if err != nil {
			slog.Error("error converting array value", "error", err, "value", aValue.Index(k), "multiplier", mult)
			continue
		}
		// Create a copy of info map to avoid modifying the original for subsequent array elements
		elementInfo := make(map[string]string)
		for key, val := range info {
			elementInfo[key] = val
		}
		elementInfo["name"] = fmt.Sprintf(name, k+1)
		cb(elementInfo, newVal)
	}
}

// parseValues parses a comma-separated string of key:value pairs into a map.
func parseValues(values string) map[string]string {
	result := make(map[string]string)
	for kv := range strings.SplitSeq(values, ",") {
		p := strings.SplitN(kv, ":", 2)
		if len(p) != 2 {
			slog.Error("error in value tag", "values", values)
			continue
		}
		result[strings.TrimSpace(p[0])] = strings.TrimSpace(p[1])
	}
	return result

}

// handleMultiplier applies a numerical multiplier to a field's value.
// It converts the field to int64, applies the multiplier, and rounds the result.
func handleMultiplier(multiplier string, field reflect.Value) (any, error) {
	if multiplier == "" {
		return field.Interface(), nil
	}
	m, err := strconv.ParseFloat(multiplier, 64)
	if err != nil {
		return nil, err
	}
	toType := reflect.TypeOf(int64(0))
	if !field.CanConvert(toType) {
		return nil, fmt.Errorf("can't convert %v to int64", field.Interface())
	}
	v := field.Convert(reflect.TypeOf(int64(0)))
	minv := 1 / m
	r := math.Round(m*float64(v.Interface().(int64))*minv) / minv
	return r, nil
}

// handleFlagsTag processes a bitmask value based on a comma-separated string of flag descriptions.
// It returns a comma-separated string of active flag descriptions.
func handleFlagsTag(flags string, val uint32) string {
	result := ""
	fl := strings.Split(flags, ",")
	nbits := len(fl)
	for n, bitval := range fl {
		if (val & (1 << (nbits - n - 1))) != 0 {
			if result != "" {
				result = fmt.Sprintf("%s, %s", result, strings.TrimSpace(bitval))
			} else {
				result = strings.TrimSpace(bitval)
			}
		}
	}
	return result
}

// handleBitgroupsTag processes a string value containing multiple bit groups,
// typically returned by inverters as ASCII characters. Each bit group is parsed
// based on provided descriptions and converted into a human-readable string.
func handleBitgroupsTag(bgroups string, val string) string {
	groups := strings.Split(bgroups, "|")
	groupIdx := 0
	var result string
	for _, group := range groups {
		descriptions := strings.Split(group, ",")
		if (groupIdx + len(descriptions)) > len(val) {
			slog.Error("error in bitgroup", "value", val, "bitgroup", bgroups)
			if result == "" {
				return val
			}
			return result
		}
		values := val[groupIdx : groupIdx+len(descriptions)]
		groupIdx += len(descriptions)
		v, _ := strconv.ParseUint(string(values), 16, 32)
		var strs []string
		gi := 0
		for m := 7; m >= 0; m-- {
			if m >= len(descriptions) {
				strs = append(strs, "")
			} else {
				strs = append(strs, descriptions[gi])
				gi++
			}
		}
		f := handleFlagsTag(strings.Join(strs, ","), uint32(v))
		if f == "" {
			continue
		}
		if result != "" {
			result = fmt.Sprintf("%s, %s", result, strings.TrimSpace(f))
		} else {
			result = strings.TrimSpace(f)
		}
	}
	return result
}

// WriteTo writes the struct fields' description value and unit as text to the given writer.
func WriteTo(writer io.Writer, data any) {
	cb := func(info map[string]string, val any) {
		fmt.Fprintf(writer, "%s: %v%s\n", info["desc"], val, info["unit"])
	}
	TraverseStruct(data, cb)
}
