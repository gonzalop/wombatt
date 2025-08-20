## inverter-query
`inverter-query` sends commands to various inverter types, including PI30, Solark, EG4 18kPV and EG4 6000XP protocols.

### Usage

```
wombatt inverter-query --address=ADDRESS,... --command=COMMAND,... [flags]
```

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
| `-p`, `--address` | Ports or addresses used for communication with the inverters | |
| `-c`, `--command` | Commands to send to the inverters | |
| `-B`, `--baud-rate` | Baud rate | `2400` |
| `--data-bits` | Number of data bits for serial port | `8` |
| `--stop-bits` | Number of stop bits for serial port | `1` |
| `--parity` | Parity for serial port (N, E, O) | `N` |
| `-t`, `--read-timeout` | Per inverter timeout for processing all the commands being sent | `5s` |
| `-T`, `--device-type` | One of serial,hidraw,tcp | `serial` |
| `-I`, `--inverter-type` | Type of inverter protocol (pi30, solark, eg4_18kpv, eg4_6000xp) | `pi30` |
| `-R`, `--protocol` | Modbus protocol (auto, ModbusRTU, ModbusTCP) | `auto` |
| `-i`, `--modbus-id` | Modbus slave ID | `1` |

### Examples

**Querying a PI30 inverter (e.g., EG4-6500EX):**

```bash
./wombatt inverter-query -p /dev/ttyS1 -c Q1 -I pi30
```

**Querying a Solark inverter:**

```bash
./wombatt inverter-query -p /dev/ttyUSB0 -c RealtimeData -I solark -R ModbusRTU -i 1
```

**Querying an EG4 18kPV inverter:**

```bash
./wombatt inverter-query -p /dev/ttyUSB0 -c RealtimeData -I eg4_18kpv -R ModbusRTU -i 0
```

**Querying an EG4 6000XP inverter:**

```bash
./wombatt inverter-query -p /dev/ttyUSB0 -c RealtimeData -I eg4_6000xp -R ModbusRTU -i 1
```
