## solark-query

Query Solark inverter via Modbus RTU/TCP

### Synopsis

This command allows you to query Solark inverters using the Modbus RTU or Modbus TCP protocol.
It reads real-time data and intrinsic attributes (like serial number) from the inverter.

```
wombatt solark-query [flags]
```

### Options

```
  -b, --baud int       Baud rate for serial port (default 9600)
  -D, --data-bits int  Data bits for serial port (default 8)
  -h, --help           help for solark-query
  -i, --id int         Modbus slave ID (default 1)
  -P, --parity string  Parity for serial port (N, E, O) (default "N")
  -p, --port string    Port name (e.g., /dev/ttyUSB0, COM1, localhost:502)
  -R, --protocol string  Modbus protocol (auto, ModbusRTU, ModbusTCP) (default "auto")
  -S, --stop-bits int  Stop bits for serial port (default 1)
  -t, --timeout int    Timeout in seconds (default 5)
```
