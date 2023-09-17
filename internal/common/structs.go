// Package common has utility functions used by differnt commands.
package common

import (
	"fmt"
	"log"
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
	for i := 0; i < nfields; i++ {
		f := stType.Field(i)
		if f.Type.Kind() == reflect.Struct {
			TraverseStruct(stValue.Field(i).Interface(), cb)
			continue
		}
		v := stValue.Field(i)
		if f.Name == "_" || f.Tag.Get("skip") != "" {
			continue
		}
		desc := f.Tag.Get("desc")
		if desc == "" {
			desc = f.Name
		}
		info := make(map[string]string)
		name := f.Tag.Get("name")
		unit := f.Tag.Get("unit")
		mult := f.Tag.Get("multiplier")
		info["name"] = name
		info["unit"] = unit
		info["desc"] = desc
		info["icon"] = f.Tag.Get("icon")
		info["dclass"] = f.Tag.Get("dclass")
		val := v.Interface()
		if f.Type.Kind() == reflect.Array {
			if f.Tag.Get("type") == "string" {
				cb(info, string(v.Bytes()))
				continue
			}
			aValue := reflect.ValueOf(val)
			for k := 0; k < aValue.Len(); k++ {
				newVal, err := handleMultiplier(mult, aValue.Index(k))
				if err != nil {
					log.Printf("error converting: %v\n", err)
					continue
				}
				info["name"] = fmt.Sprintf(name, k+1)
				cb(info, newVal)
			}
			continue
		}
		vals := f.Tag.Get("values")
		if vals != "" {
			val = fmt.Sprintf("%v", v.Interface())
			m := parseValues(vals)
			if v, ok := m[val.(string)]; ok {
				val = v
			}
		}
		if mult != "" {
			newVal, err := handleMultiplier(mult, v)
			if err != nil {
				log.Printf("error converting: %v\n", err)
				continue
			}
			cb(info, newVal)
			continue
		}
		flags := f.Tag.Get("flags")
		if flags != "" && (f.Type.Name() == "uint8" || f.Type.Name() == "uint16") {
			u16 := v.Convert(reflect.TypeOf(uint16(0))).Interface().(uint16)
			val = handleFlagsTag(flags, u16)
			if val == "" {
				val = u16
			}
		}
		bgroups := f.Tag.Get("bitgroups")
		if bgroups != "" {
			str := v.String()
			if str != "" {
				val = handleBitgroupsTag(bgroups, str)
			}
		}

		cb(info, val)
	}
}

func parseValues(values string) map[string]string {
	result := make(map[string]string)
	for _, kv := range strings.Split(values, ",") {
		p := strings.SplitN(kv, ":", 2)
		if len(p) != 2 {
			log.Printf("error in value tag: %v", values)
			continue
		}
		result[strings.TrimSpace(p[0])] = strings.TrimSpace(p[1])
	}
	return result

}

func handleMultiplier(multiplier string, field reflect.Value) (interface{}, error) {
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

func handleFlagsTag(flags string, val uint16) string {
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

// The inverters might return some bits groups in one ASCII character.
func handleBitgroupsTag(bgroups string, val string) string {
	groups := strings.Split(bgroups, "|")
	groupIdx := 0
	var result string
	for _, group := range groups {
		descriptions := strings.Split(group, ",")
		if (groupIdx + len(descriptions)) > len(val) {
			log.Printf("error in bitgroup: got %s with descriptions '%s'", val, bgroups)
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
		f := handleFlagsTag(strings.Join(strs, ","), uint16(v))
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
