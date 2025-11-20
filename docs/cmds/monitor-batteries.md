## monitor-batteries
`monitor-batteries` monitors batteries state, MQTT publishing optional.

### Usage

```
wombatt monitor-batteries --address=STRING --battery-id=BATTERY-ID,... [flags]
```

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
| `-p`, `--address` | Serial port attached to the batteries | |
| `-B`, `--baud-rate` | Baud rate for serial ports | `9600` |
| `-i`, `--battery-id` | IDs of the batteries to monitor | |
| `-P`, `--poll-interval` | Time to wait between polling cycles | `10s` |
| `-t`, `--read-timeout` | Timeout when reading from devices | `500ms` |
| `--bms-type` | One of EG4LLv2,lifepower4,lifepowerv2,pacemodbus | `EG4LLv2` |
| `--mqtt-prefix` | MQTT prefix for the fields published | `eg4` |
| `-w`, `--web-server-address` | Address to use for serving the web dashboard and prometheus metrics. <IP>:<Port>, i.e., 127.0.0.1:8080 | |
| `--protocol` | One of auto,ModbusRTU,ModbusTCP,lifepower4 | `auto` |
| `-T`, `--device-type` | One of serial,hidraw,tcp | `serial` |

#### MQTT Flags

| Flag | Description | Environment Variable |
| --- | --- | --- |
| `--mqtt-broker` | The MQTT server to publish battery data. E.g. tcp://127.0.0.1:1883 | `$MQTT_BROKER` |
| `--mqtt-password` | Password for the MQTT connection | `$MQTT_PASSWORD` |
| `--mqtt-topic-prefix` | Prefix for all topics published to MQTT | `$MQTT_TOPIC_PREFIX` |
| `--mqtt-user` | User for the MQTT connection | `$MQTT_USER` |

### Examples

To monitor batteries with IDs 2 thru 6, and publish to MQTT and a local web page on port 8000, you can run:
```
$ ./wombatt monitor-batteries -w :8000 -p /dev/ttyUSB1 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword --battery-id 2,3,4,5,6
```

The default prefix for the items added to MQTT is `eg4` (i.e., `homeassistant/eg4_battery2_info/...`).

The same infomation is made available via a web dashboard and prometheus metrics on port 8000.
The battery information is also available as text or JSON (add `?format=json` to the URL),
with the ability to request specific fields (`?fields=<name>`).
Prometheus metrics are available at the `/metrics` endpoint.