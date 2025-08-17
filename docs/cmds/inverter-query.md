## inverter-query
`inverter-query` sends commands to various inverter types, including PI30, Solark, EG4 18kPV and EG4 6000XP protocols.

### Flags

Usage: wombatt inverter-query --address=ADDRESS,... --command=COMMAND,... [flags]

Flags:
  -h, --help                    Show context-sensitive help.
  -l, --log-level="info"        Set the logging level (debug|info|warn|error)
  -v, --version                 Print version information and quit

  -p, --address=ADDRESS,...     Ports or addresses used for communication with
                                the inverters
  -c, --command=COMMAND,...     Commands to send to the inverters
  -B, --baud-rate=2400          Baud rate
      --data-bits=8             Number of data bits for serial port
      --stop-bits=1             Number of stop bits for serial port
      --parity="N"              Parity for serial port (N, E, O)
  -t, --read-timeout=5s         Per inverter timeout for processing all the
                                commands being sent
  -T, --device-type="serial"    One of serial,hidraw,tcp
  -I, --inverter-type="pi30"    Type of inverter protocol (pi30, solark,
                                eg4_18kpv, eg4_6000xp)
  -R, --protocol="auto"         Modbus protocol (auto, ModbusRTU, ModbusTCP)
  -i, --modbus-id=1             Modbus slave ID

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