// Package pi30 for interfacing with EG4 6500 inverters
package pi30

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"reflect"
	"strconv"
	"strings"

	"wombatt/internal/common"

	"github.com/howeyc/crc16"
)

func RunCommands(ctx context.Context, port io.ReadWriter, commands []string) ([]any, []error) {
	var result []any
	var resultErr []error
	for _, cmd := range commands {
		r, err := RunCommand(ctx, port, cmd)
		result = append(result, r)
		resultErr = append(resultErr, err)
	}
	return result, resultErr
}

func RunCommand(ctx context.Context, port io.ReadWriter, cmd string) (any, error) {
	type data struct {
		strs []string
		err  error
	}
	ch := make(chan *data, 1)
	go func() {
		err := sendCommand(port, cmd)
		if err != nil {
			err = fmt.Errorf("send error in %s: %v\n", cmd, err)
			ch <- &data{nil, err}
			return
		}
		resp, err := readResponse(port)
		if err != nil {
			err = fmt.Errorf("error reading response in %s: %v\n", cmd, err)
			ch <- &data{nil, err}
			return
		}
		ch <- &data{resp, nil}
	}()

	var resp *data
	select {
	case <-ctx.Done():
		return nil, fmt.Errorf("timed out sending %s", cmd)
	case resp = <-ch:
		if resp.strs == nil && resp.err != nil {
			return nil, resp.err
		}
	}

	switch len(resp.strs) {
	case 0:
		return nil, fmt.Errorf("invalid response %s\n", cmd)
	case 1:
		if resp.strs[0] == "NAK" {
			return nil, fmt.Errorf("NAK received")
		}
	}
	result := StructForCommand(cmd)
	if err := decodeResponse(resp.strs, result); err != nil {
		return nil, fmt.Errorf("decode error for %s: %v\n", cmd, err)
	}
	return result, nil
}

func StructForCommand(cmd string) any {
	var result any
	switch cmd {
	case "Q1":
		result = &Q1Response{}
	case "QPIRI":
		result = &QPIRIResponse{}
	case "QPIGS":
		result = &QPIGSResponse{}
	case "QPIGS2":
		result = &QPIGS2Response{}
	default:
		if len(cmd) > 4 && cmd[0:4] == "QPGS" {
			result = &QPGSResponse{}
		} else {
			result = &EmptyResponse{}
		}
	}
	return result
}

func sendCommand(port io.Writer, command string) error {
	var b bytes.Buffer
	b.WriteString(command)
	c := crc([]byte(command))
	b.WriteByte(byte((c >> 8)))
	b.WriteByte(byte(c & 0x0ff))
	b.WriteByte('\r')
	if _, err := port.Write(b.Bytes()); err != nil {
		return err
	}
	return nil
}

func readResponse(port io.Reader) ([]string, error) {
	r := bufio.NewReader(port)
	b, err := r.ReadSlice('\r')
	if err != nil {
		return nil, err
	}
	if b[0] != '(' {
		return nil, fmt.Errorf("wrong start for a response: got '%s' want '(')", string(b))
	}
	if len(b) < 4 {
		// Minimum response would be one for '(', one for CRC, one for '\r' and one for data
		return nil, fmt.Errorf("short response: '%v'", b)
	}

	computed := crc(b[0 : len(b)-3])
	received := uint16(b[len(b)-3])*256 + uint16(b[len(b)-2])
	if received != computed {
		log.Printf("crc error: got %04x, want %04x\n", received, computed)
		return nil, fmt.Errorf("crc error: got %04x, want %04x\n", received, computed)
	}
	s := string(b[1 : len(b)-3])
	result := strings.Split(s, " ")
	return result, nil
}

func decodeResponse(parts []string, target any) error {
	log.Printf("%v\n", parts)
	lenParts := len(parts)
	stValue := reflect.ValueOf(target).Elem()
	stType := stValue.Type()
	nfields := stType.NumField()
	if lenParts < nfields {
		log.Printf("wrong number of fields: got %d, want: %d\b", lenParts, nfields)
		nfields = lenParts
	}
	for i := 0; i < nfields; i++ {
		f := stType.Field(i)
		if f.Name == "_" {
			continue
		}
		v := stValue.Field(i)
		isInt := false
		isUint := false
		bits := 0
		base := 10
		switch v.Interface().(type) {
		case int8:
			isInt = true
			bits = 8
			if f.Tag.Get("parseas") == "binary" {
				base = 2
			}
		case uint8:
			isUint = true
			bits = 8
			if f.Tag.Get("parseas") == "binary" {
				base = 2
			}
		case uint16:
			isUint = true
			bits = 16
			if f.Tag.Get("parseas") == "binary" {
				base = 2
			}
		case int16:
			isInt = true
			bits = 16
		case int:
			isInt = true
			bits = 32
		case float32:
			val, err := strconv.ParseFloat(parts[i], 32)
			if err != nil {
				return fmt.Errorf("error converting float type for %s: value '%s'", f.Name, parts[i])
			}
			v.SetFloat(val)
		case string:
			v.SetString(parts[i])
		default:
			return fmt.Errorf("unknown type: add %v", v.Type())
		}
		if isInt {
			num, err := strconv.ParseInt(parts[i], base, bits)
			if err != nil {
				return fmt.Errorf("error converting integer type for %s, value '%s: %v'", f.Name, parts[i], err)
			}
			v.SetInt(num)
		} else if isUint {
			num, err := strconv.ParseUint(parts[i], base, bits)
			if err != nil {
				return fmt.Errorf("error converting integer type for %s, value '%s: %v'", f.Name, parts[i], err)
			}
			v.SetUint(num)
		}
	}
	return nil
}

func WriteTo(writer io.Writer, data any) {
	cb := func(info map[string]string, val any) {
		fmt.Fprintf(writer, "%s: %v%s\n", info["desc"], val, info["unit"])
	}
	common.TraverseStruct(data, cb)
}

func crc(data []byte) uint16 {
	crc := crc16.Checksum(data, crc16.CCITTFalseTable)
	low := crc & 0xff
	if low == 0x28 || low == 0x0D || low == 0x0A {
		crc += 1
	}
	high := crc >> 8
	if high == 0x28 || high == 0x0D || high == 0x0A {
		crc += 256
	}
	return crc
}
