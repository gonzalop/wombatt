## inverter-query
`inverter-query` sends commands to various inverter types, including PI30, Solark, and EG4 18kPV protocols.

### Flags

*   `-p`, `--address` (required): Ports or addresses used for communication with the inverters (e.g., `/dev/ttyUSB0`, `COM1`, `tcp://localhost:502`).
*   `-c`, `--command` (required): Commands to send to the inverters. For PI30, examples are `Q1`, `QPIRI`, `QPIGS`. For Solark, `RealtimeData` and `IntrinsicAttributes`. For EG4 18kPV, `RealtimeData`.
*   `-B`, `--baud-rate` (default: `2400`): Baud rate for serial ports.
*   `--data-bits` (default: `8`): Number of data bits for serial port.
*   `--stop-bits` (default: `1`): Number of stop bits for serial port.
*   `--parity` (default: `N`): Parity for serial port (`N`, `E`, `O`).
*   `-t`, `--read-timeout` (default: `5s`): Per inverter timeout for processing all the commands being sent.
*   `-T`, `--device-type` (default: `serial`): One of `serial`, `hidraw`, `tcp`.
*   `-I`, `--inverter-type` (default: `pi30`): Type of inverter protocol (`pi30`, `solark`, `eg4_18kpv`).
*   `-R`, `--protocol` (default: `auto`): Modbus protocol (`auto`, `ModbusRTU`, `ModbusTCP`).
*   `-i`, `--modbus-id` (default: `1`): Modbus slave ID.

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