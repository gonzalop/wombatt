## monitor-inverters
`monitor-inverters` monitors inverters using PI30, Solark, EG4 18kPV, or EG4 6000XP Modbus protocol, with optional MQTT publishing.

### Examples
The command below will monitor the inverters connected to /dev/ttyS0 and
/dev/ttyS1, run the `Q1`, `QPIGS`, and `QPIRI` commands on both of them,
and `QPGS1` or `QPGS2`.

~~~
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword /dev/ttyS0,Q1:QPIGS:QPIRI:QPGS2,eg4_1 /dev/ttyS1,Q1:QPIGS:QPIRI:QPGS1,eg4_2
~~~

To monitor a Solark inverter via Modbus RTU:

~~~
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword -R ModbusRTU -i 1 /dev/ttyUSB0,RealtimeData:IntrinsicAttributes,solark_1,solark
~~~

To monitor an EG4 18kPV inverter via Modbus RTU:

~~~
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword -R ModbusRTU -i 0 /dev/ttyUSB0,RealtimeData,eg4_18kpv_1,eg4_18kpv
~~~

To monitor an EG4 6000XP inverter via Modbus RTU:

~~~
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword -R ModbusRTU -i 1 /dev/ttyUSB0,RealtimeData,eg4_6000xp_1,eg4_6000xp
~~~


### Arguments and Flags

Usage: wombatt monitor-inverters <monitors> ... [flags]

Arguments:
  <monitors> ...    <device>,<command1[:command2:command3...]>,<mqtt_prefix>[,<inverter_type>].
                    E.g. /dev/ttyS0,QPIRI:QPGS1,eg4_1,pi30 or
                    /dev/ttyUSB0,RealtimeData:IntrinsicAttributes,solark_1,solark
                    or /dev/ttyUSB0,RealtimeData,eg4_18kpv_1,eg4_18kpv or
                    /dev/ttyUSB0,RealtimeData,eg4_6000xp_1,eg4_6000xp. Valid
                    solark commands are RealtimeData and IntrinsicAttributes.
                    Valid eg4_18kpv/eg4_6000xp commands are RealtimeData.

Flags:
  -h, --help                    Show context-sensitive help.
  -l, --log-level="info"        Set the logging level (debug|info|warn|error)
  -v, --version                 Print version information and quit

  -B, --baud-rate=2400          Baud rate for serial ports
      --data-bits=8             Number of data bits for serial port
      --stop-bits=1             Number of stop bits for serial port
      --parity="N"              Parity for serial port (N, E, O)
  -P, --poll-interval=10s       Time to wait between polling cycles
  -t, --read-timeout=5s         Timeout when reading from devices
  -w, --web-server-address=STRING
                                Address to use for serving HTTP. <IP>:<Port>,
                                i.e., 127.0.0.1:8080
  -T, --device-type="serial"    One of serial,hidraw,tcp
  -R, --protocol="auto"         Modbus protocol (auto, ModbusRTU, ModbusTCP)
  -i, --modbus-id=1             Modbus slave ID

MQTT
  --mqtt-broker=STRING      The MQTT server to publish battery data. E.g.
                            tcp://127.0.0.1:1883 ($MQTT_BROKER)
  --mqtt-password=STRING    Password for the MQTT connection ($MQTT_PASSWORD)
  --mqtt-topic-prefix="homeassistant"
                            Prefix for all topics published to MQTT
                            ($MQTT_TOPIC_PREFIX)
  --mqtt-user=STRING        User for the MQTT connection ($MQTT_USER)