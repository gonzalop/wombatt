## monitor-inverters
`monitor-inverters` monitors inverters state, with optional MQTT publishing. It can be used with PI30, Solark, EG4 18kPV, or EG4 6000XP Modbus protocols.

### Usage

```
wombatt monitor-inverters <monitors> ... [flags]
```

### Arguments

| Argument | Description |
| --- | --- |
| `<monitors>...` | `<device>,<command1[:command2:command3...]>,<mqtt_prefix>[,<inverter_type>]`.<br>E.g. `/dev/ttyS0,QPIRI:QPGS1,eg4_1,pi30` or<br>`/dev/ttyUSB0,RealtimeData:IntrinsicAttributes,solark_1,solark` or<br>`/dev/ttyUSB0,RealtimeData,eg4_18kpv_1,eg4_18kpv` or<br>`/dev/ttyUSB0,RealtimeData,eg4_6000xp_1,eg4_6000xp`.<br>Valid solark commands are `RealtimeData` and `IntrinsicAttributes`.<br>Valid eg4_18kpv/eg4_6000xp commands are `RealtimeData`. |

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
| `-B`, `--baud-rate` | Baud rate for serial ports | `2400` |
| `--data-bits` | Number of data bits for serial port | `8` |
| `--stop-bits` | Number of stop bits for serial port | `1` |
| `--parity` | Parity for serial port (N, E, O) | `N` |
| `-P`, `--poll-interval` | Time to wait between polling cycles | `10s` |
| `-t`, `--read-timeout` | Timeout when reading from devices | `5s` |
| `-w`, `--web-server-address` | Address to use for serving HTTP. <IP>:<Port>, i.e., 127.0.0.1:8080 | |
| `-T`, `--device-type` | One of serial,hidraw,tcp | `serial` |
| `-R`, `--protocol` | Modbus protocol (auto, ModbusRTU, ModbusTCP) | `auto` |
| `-i`, `--modbus-id` | Modbus slave ID | `1` |

#### MQTT Flags

| Flag | Description | Environment Variable |
| --- | --- | --- |
| `--mqtt-broker` | The MQTT server to publish battery data. E.g. tcp://127.0.0.1:1883 | `$MQTT_BROKER` |
| `--mqtt-password` | Password for the MQTT connection | `$MQTT_PASSWORD` |
| `--mqtt-topic-prefix` | Prefix for all topics published to MQTT | `$MQTT_TOPIC_PREFIX` |
| `--mqtt-user` | User for the MQTT connection | `$MQTT_USER` |

### Examples

The command below will monitor the inverters connected to /dev/ttyS0 and
/dev/ttyS1, run the `Q1`, `QPIGS`, and `QPIRI` commands on both of them,
and `QPGS1` or `QPGS2`.

```
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword /dev/ttyS0,Q1:QPIGS:QPIRI:QPGS2,eg4_1 /dev/ttyS1,Q1:QPIGS:QPIRI:QPGS1,eg4_2
```

To monitor a Solark inverter via Modbus RTU:

```
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword -R ModbusRTU -i 1 /dev/ttyUSB0,RealtimeData:IntrinsicAttributes,solark_1,solark
```

To monitor an EG4 18kPV inverter via Modbus RTU:

```
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword -R ModbusRTU -i 0 /dev/ttyUSB0,RealtimeData,eg4_18kpv_1,eg4_18kpv
```

To monitor an EG4 6000XP inverter via Modbus RTU:

```
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword -R ModbusRTU -i 1 /dev/ttyUSB0,RealtimeData,eg4_6000xp_1,eg4_6000xp
```
