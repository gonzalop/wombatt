## monitor-inverters
`monitor-inverters` monitors inverters using PI30 protocol, MQTT publishing optional.

### Examples
The command below will monitor the inverters connected to /dev/ttyS0 and
/dev/ttyS1, run the `Q1`, `QPIGS`, and `QPIRI` commands on both of them,
and `QPGS1` or `QPGS2`.

~~~
$ ./wombatt monitor-inverters -w :9000 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword /dev/ttyS0,Q1:QPIGS:QPIRI:QPGS2,eg4_1 /dev/ttyS1,Q1:QPIGS:QPIRI:QPGS1,eg4_2
~~~

The information will be published to the specified MQTT server, with prefixes `eg4_1`
and `eg4_2` depending on the inverter, along with HomeAssistant autodiscovery configuration.

The same infomation is made available via web on port 9000
(http://127.0.0.1:9000/inverters/1/Q1 and so on) as text or JSON (add
`?format=json` to the URL), with the ability to request specific
fields (`?fields=<name>`).


