## monitor-inverters
`monitor-inverters` monitors inverters using PI30, Solark, or EG4 18kPV Modbus protocol, with optional MQTT publishing.

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


### Arguments

`<device>,<command1[:command2:command3...]>,<mqtt_prefix>[,<inverter_type>]`

*   `<device>`: The serial port or TCP address of the inverter.
*   `<command1[:command2:command3...]>`: A colon-separated list of commands to run on the inverter. For PI30 inverters, valid commands include `QPIRI`, `QPIGS`, etc. For Solark inverters, the supported commands are `RealtimeData` and `IntrinsicAttributes`.
*   `<mqtt_prefix>`: A prefix for MQTT topics (e.g., `eg4_1`).
*   `<inverter_type>`: The type of inverter protocol. Must be `pi30` (default), `solark`, or `eg4_18kpv`.

### Options

*   `-B, --baud-rate uint`: Baud rate for serial ports (default 2400)
*   `--data-bits int`: Number of data bits for serial port (default 8)
*   `--stop-bits int`: Number of stop bits for serial port (default 1)
*   `--parity string`: Parity for serial port (N, E, O) (default "N")
*   `-R, --protocol string`: Modbus protocol (auto, ModbusRTU, ModbusTCP) (default "auto")
*   `-i, --id int`: Modbus slave ID (default 1)

The information will be published to the specified MQTT server, with prefixes `eg4_1`
and `eg4_2` depending on the inverter, along with HomeAssistant autodiscovery configuration.

The same infomation is made available via web on port 9000
(http://127.0.0.1:9000/inverters/1/Q1 and so on) as text or JSON (add
`?format=json` to the URL), with the ability to request specific
fields (`?fields=<name>`).


