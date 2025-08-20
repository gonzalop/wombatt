## battery-info
`battery-info` displays battery status information.

### Usage

```
wombatt battery-info --address=STRING --battery-id=BATTERY-ID,... [flags]
```

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
| `-p`, `--address` | Serial port or address used for communication | |
| `-i`, `--battery-id` | IDs of the batteries to get info from. | |
| `-t`, `--read-timeout` | Timeout when reading from serial ports | `500ms` |
| `-B`, `--baud-rate` | Baud rate | `9600` |
| `--bms-type` | One of EG4LLv2,lifepower4,lifepowerv2,pacemodbus | `EG4LLv2` |
| `--protocol` | One of auto,ModbusRTU,ModbusTCP,lifepower4 | `auto` |
| `-T`, `--device-type` | One of serial,hidraw,tcp | `serial` |

### Examples

For instance, to query the battery with ID #2:
```
$ ./wombatt battery-info --address /dev/ttyUSB0 --battery-id 2
Battery #2
===========
battery voltage: 52.9V
current: 2.5A
cell 1 voltage: 3.305V
cell 2 voltage: 3.306V
cell 3 voltage: 3.306V
cell 4 voltage: 3.306V
cell 5 voltage: 3.306V
cell 6 voltage: 3.307V
cell 7 voltage: 3.306V
cell 8 voltage: 3.307V
cell 9 voltage: 3.306V
cell 10 voltage: 3.307V
cell 11 voltage: 3.306V
cell 12 voltage: 3.306V
cell 13 voltage: 3.306V
cell 14 voltage: 3.307V
cell 15 voltage: 3.306V
cell 16 voltage: 3.307V
pcb temp: 31°C
max temp: 32°C
avg temp: 30°C
cap remaining: 46%
max charging current: 100A
soh: 100%
soc: 47%
status: inactive/charging
warning: 0
protection: 0
error code: 0
cycle counts: 10
full capacity: 100000mAh
temp1: 30°C
temp2: 30°C
temp3: 30°C
temp4: 30°C
temp5: 0
temp6: 0
cell num: 16
designed capacity: 100Ah
cell balance status: 0
max cell voltage: 3.307V
min cell voltage: 3.305V
mean cell voltage: 3.306V
median cell voltage: 3.306V
model: LFP-51.2V100Ah-V1.0
firmware version: Z02T04
serial: 2022-10-26
```

**Query multiple batteries:**

```
$ ./wombatt battery-info --address /dev/ttyUSB0 --battery-id 1,2,3
```

**Using a different BMS type:**

```
$ ./wombatt battery-info --address /dev/ttyUSB0 --battery-id 1 --bms-type lifepower4
```

**Connecting via TCP:**

```
$ ./wombatt battery-info --address 192.168.1.100:502 --battery-id 1 --device-type tcp --protocol ModbusTCP
```