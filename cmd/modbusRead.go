package cmd

import (
	"bytes"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log"
	"log/slog"
	"math"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"

	"wombatt/internal/common"
	"wombatt/internal/modbus"

	"go.bug.st/serial"
	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

type ModbusReadCmd struct {
	Address          string        `short:"p" required:"" help:"Port or TCP address used for communication"`
	ID               uint8         `required:"" help:"Device ID"`
	Start            uint16        `required:"" help:"Start address of the first register to read"`
	Count            uint8         `required:"" help:"Number of registers to read"`
	RegisterType     string        `default:"holding" help:"valid values are 'input' or 'holding'"`
	BaudRate         uint          `short:"B" default:"9600" help:"Baud rate"`
	ReadTimeout      time.Duration `short:"t" default:"500ms" help:"Timeout when reading from serial ports"`
	Protocol         string        `default:"auto" enum:"${protocols}" help:"One of ${protocols}"`
	DeviceType       string        `short:"T" default:"serial" enum:"${device_types}" help:"One of ${device_types}"`
	OutputFormat     string        `short:"o" help:"Output format for the registers read"`
	OutputFormatFile string        `short:"O" help:"Output format file for the registers read"`
}

func (cmd *ModbusReadCmd) Help() string {
	return `The register read are written as a hexadecimal dump. For a custom output, use the '-o' or '--output-format'.
	The format syntax is a comma-separated list of:

			<type>[:[<name>][:[<unit>][:[<multiplier>][:[<string>]]]]]

	All but <type> are optional, and all accept an array size prefix. I.e., for an array of 4 16-bits unsigned
	values, you can use '[4]u16' as the type.

	<type> can be one of byte, i8, u8, i16, u16, i32, u32.
		'u' is used for unsigned values and 'i' for signed. The number is the number of bits.

	<name> can be any name accepted by Go. The special name denoted by a single underscore (_) will
		omit that field from the output. Spaces are converted to underscores and the name is capitalized.

	<unit> is any unit you want displayed next to the value.

	<multiplier> is a floating point number to convert from the <unit> display value to the register value.
		For instance, if the register is in 10mV and has a decimal value of 3277, a multiplier of 0.01 and
		a unit of V would display the value in volts.

	<string> is literally 'string' and is used to display byte arrays as a string.

	The same comma-separated values for the -o option can be read from a file, one line per register, with
	comments starting with the '#' character. To read the formatting values from a file, use the -O option.

	Example output format values:
		u16,i8,u32 -- 3 fields: unsigned 16-bit, signed 8-bit, and unsigned 32-bit integers.

		u16:Voltage:V:0.01,i8,u32  -- same as above, but the first field will be named 'Voltage',
			expects the value in 10mV, and converts them to V

		[10]byte:Serial number:::string -- it will print 10 bytes as a string with the field name 'Serial_number'.

	`
}

func (cmd *ModbusReadCmd) Run(globals *Globals) error {
	if cmd.ID == 0 || cmd.ID > 247 {
		log.Fatal("id must be between 1 and 247")
	}
	if cmd.Count > 125 {
		log.Fatal("count must be between <= 125")
	}
	if cmd.OutputFormat != "" && cmd.OutputFormatFile != "" {
		log.Fatal("only one of -o and -O can be used")
	}
	if cmd.OutputFormatFile != "" {
		f, err := readOutputFormatFile(cmd.OutputFormatFile)
		if err != nil {
			log.Fatalf("error reading output format file '%v': %v", cmd.OutputFormatFile, err)
		}
		cmd.OutputFormat = f
	}
	portOptions := &common.PortOptions{
		Address: cmd.Address,
		Mode:    &serial.Mode{BaudRate: int(cmd.BaudRate)},
		Type:    common.DeviceTypeFromString[cmd.DeviceType],
	}
	port := common.OpenPortOrFatal(portOptions)
	reader, err := modbus.Reader(port, cmd.Protocol, "")
	if err != nil {
		log.Fatal(err.Error())
	}
	readFunc := reader.ReadHoldingRegisters
	if cmd.RegisterType == "input" {
		readFunc = reader.ReadInputRegisters
	}
	data, err := readFunc(cmd.ID, cmd.Start, cmd.Count)
	if err != nil {
		slog.Error("error reading registers", "address", cmd.Address, "error", err)
		log.Fatal(err.Error())
	}
	if cmd.OutputFormat == "" {
		fmt.Printf("%v ID#%d:\n%s\n", cmd.Address, cmd.ID, hex.Dump(data))
	} else {
		if err := printRegisters(cmd.OutputFormat, data); err != nil {
			log.Fatal(err.Error())
		}
	}
	return nil
}

func readOutputFormatFile(fileName string) (string, error) {
	data, err := os.ReadFile(fileName)
	if err != nil {
		return "", err
	}
	input := strings.Split(string(data), "\n")
	var lines []string
	for _, s := range input {
		line := ""
		if idx := strings.IndexByte(s, '#'); idx != -1 {
			line = s[:idx]
		} else {
			line = s
		}
		line = strings.TrimSpace(line)
		if line != "" {
			lines = append(lines, line)
		}
	}
	return strings.Join(lines, ","), nil
}

func printRegisters(format string, data []byte) error {
	var fields []reflect.StructField
	for ii, v := range strings.Split(format, ",") {
		s := strings.TrimSpace(v)
		name := fmt.Sprintf("Field_%d", ii)

		var err error
		var t reflect.Type
		var tag string

		for iii, p := range strings.Split(s, ":") {
			f := strings.TrimSpace(p)

			switch iii {
			case 0: // type
				t, err = parseFieldType(f)
				if err != nil {
					return err
				}
			case 1: // name
				name = strings.ReplaceAll(cases.Title(language.English, cases.Compact).String(f), " ", "_")
			case 2:
				if f != "" {
					tag = fmt.Sprintf(`unit:"%s"`, f)
				}
			case 3: // multiplier
				if f != "" {
					tag = fmt.Sprintf(`multiplier:"%s" %s`, f, tag)
				}
			case 4: // type tag. only applies to []byte with 'type:"string"'
				if f != "" {
					tag = fmt.Sprintf(`type:"%s" %s`, f, tag)
				}
			default:
				return fmt.Errorf("too many colons in '%s'", s)
			}
		}
		pkgPath := ""
		if name == "_" {
			pkgPath = "something" // fields named "_" won't be shown, but need PkgPath set.
		}
		fields = append(fields, reflect.StructField{
			Name:    name,
			Type:    t,
			Tag:     reflect.StructTag(tag),
			PkgPath: pkgPath,
		})
	}

	inst, err := createStruct(fields, data)
	if err != nil {
		log.Fatalf("%v: check the output format syntax\n", err.Error())
	}
	if err := printStruct(inst); err != nil {
		log.Fatal(err.Error())
	}

	return nil
}

func printStruct(inst any) (err error) {
	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("%v", r)
		}
	}()

	cb := func(info map[string]string, val any) {
		fmt.Fprintf(os.Stdout, "%s: %v%s\n", info["desc"], val, info["unit"])
	}
	common.TraverseStruct(inst, cb)
	return nil
}

func createStruct(fields []reflect.StructField, data []byte) (inst any, err error) {
	defer func() {
		if r := recover(); r != nil {
			inst = nil
			err = fmt.Errorf("%v", r)
		}
	}()
	r := reflect.StructOf(fields)
	inst = reflect.New(r).Interface()
	buf := bytes.NewBuffer(data)
	if err = binary.Read(buf, binary.BigEndian, inst); err != nil {
		return nil, err
	}

	return inst, nil
}

func parseFieldType(fieldType string) (reflect.Type, error) {
	parts := strings.SplitN(fieldType, "[", 2)
	if len(parts) == 2 {
		p := strings.SplitN(parts[1], "]", 2)
		return parseArrayType(p[1], p[0])
	}
	return parseSingleType(fieldType)
}

func parseArrayType(typeName string, sizeStr string) (reflect.Type, error) {
	size, err := strconv.ParseUint(sizeStr, 10, 32)
	if err != nil || math.MaxInt32 < size {
		return nil, fmt.Errorf("%w: invalid size value %q", err, sizeStr)
	}

	t, err := parseSingleType(typeName)
	if err != nil {
		return nil, err
	}
	arrayType := reflect.New(reflect.ArrayOf(int(size), t)).Elem().Type()
	return arrayType, nil
}

func parseSingleType(typeName string) (reflect.Type, error) {
	switch typeName {
	case "byte":
		return reflect.TypeOf(byte(0)), nil
	case "int8", "i8":
		return reflect.TypeOf(int8(0)), nil
	case "int16", "i16":
		return reflect.TypeOf(int16(0)), nil
	case "int32", "i32":
		return reflect.TypeOf(int32(0)), nil
	case "uint8", "u8":
		return reflect.TypeOf(uint8(0)), nil
	case "uint16", "u16":
		return reflect.TypeOf(uint16(0)), nil
	case "uint32", "u32":
		return reflect.TypeOf(uint32(0)), nil
	default:
		return nil, fmt.Errorf("invalid type name: %s", typeName)
	}
}
