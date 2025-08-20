## modbus-read
`modbus-read` reads registers from a specified device. This is used during development.

### Usage

```
wombatt modbus-read --address=STRING --id=UINT-8 --start=UINT-16 --count=UINT-8 [flags]
```

### Description

The registers read are written as a hexadecimal dump. For a custom
output, use the '-o' or '--output-format'. The format syntax is a
comma-separated list:

```
<type>[:[<name>][:[<unit>][:[<multiplier>][:[<string>]]]]]
```

All but <type> are optional, and all accept an array size prefix.
For example, for an array of 4 16-bits unsigned values, you can
use '[4]u16' as the type.

<type> can be one of byte, i8, u8, i16, u16, i32, u32.
'u' is used for unsigned values and 'i' for signed.
The number is the number of bits.

<name> can be any name accepted by Go. The special name denoted by a
single underscore (_) will omit that field from the output. Spaces are
converted to underscores and the name is capitalized.

<unit> is any unit you want displayed next to the value.

<multiplier> is a floating point number to convert from the <unit> display
value to the register value.  For instance, if the register is in 10mV
and has a decimal value of 3277, a multiplier of 0.01 and a unit of V
would display the value in volts.

<string> is literally 'string' and is used to display byte arrays as
a string.

The same comma-separated values for the -o option can be read from a file,
one line per register, with comments starting with the '#' character.
To read formatting values from a file, use the -O option.

Example output format values:
	u16,i8,u32 -- 3 fields: unsigned 16-bit, signed 8-bit, and
				  unsigned 32-bit integers.

u16:Voltage:V:0.01,i8,u32  -- same as above, but the first field will be
		named 'Voltage', expects the value in 10mV, and converts it to V

[10]byte:Serial number:::string -- it will print 10 bytes as a string
		with the field name 'Serial_number'.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
| `-p`, `--address` | Port or TCP address used for communication | |
| `--id` | Device ID | |
| `--start` | Start address of the first register to read | |
| `--count` | Number of registers to read | |
| `--register-type` | valid values are 'input' or 'holding' | `holding` |
| `-B`, `--baud-rate` | Baud rate | `9600` |
| `--protocol` | One of auto,ModbusRTU,ModbusTCP,lifepower4 | `auto` |
| `-T`, `--device-type` | One of serial,hidraw,tcp | `serial` |
| `-o`, `--output-format` | Output format for the registers read | |
| `-O`, `--output-format-file` | Output format file for the registers read | |

### Examples
To read 38 registers from device ID #2 starting at address 0:
```
$ ./wombatt modbus-read -p /dev/ttyUSB1 --id 2 --start 0 --count 38
/dev/ttyUSB1:
00000000  02 03 4c 14 af 00 f0 0c  ed 0c ee 0c ed 0c ef 0c  |..L.............|
00000010  ed 0c ee 0c ed 0c ee 0c  ee 0c ef 0c ee 0c ee 0c  |................|
00000020  ed 0c ee 0c ed 0c ee 00  1f 00 20 00 1e 00 30 00  |.......... ...0.|
00000030  64 00 64 00 30 00 01 00  00 00 00 00 00 00 00 00  |d.d.0...........|
00000040  0a 15 75 2a 00 1e 1e 1e  1e 00 00 00 10 03 e8 45  |..u*...........E|
00000050  8d                                                |.|
```

To read the model, firmware, and serial in a EG4v2LL battery that is connected to a device
that uses the Modbus TCP protocol, and then formatting the 3 fields as a string:
```
$ ./wombatt modbus-read -T tcp -p 192.168.1.123:502 --id 2 --start 105 --count 23 -o "[24]byte:Model:::string, [6]byte:Firmware version:::string, [16]byte:Serial:::string"
2023/12/18 21:58:42.304161 Opening 192.168.1.123:502...
Model: LFP-51.2V100Ah-V1.0
Firmware_Version: Z02T04
Serial: 2022-10-26
```