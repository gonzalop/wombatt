## monitor-batteries
`monitor-batteries` monitors batteries state, MQTT publishing optional.

### Examples
To monitor batteries with IDs 2 thru 6, and publish to MQTT and a local web page on port 8000, you can run:
~~~
$ ./wombatt monitor-batteries -w :8000 -p /dev/ttyUSB1 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword --battery-ids 2,3,4,5,6
~~~

The default prefix for the items added to MQTT is `eg4` (i.e., `homeassistant/eg4_battery2_info/...`).

The same infomation is made available via web on port 8000
(http://127.0.0.1:8000/battery/2 and so on) as text or JSON (add
`?format=json` to the URL), with the ability to request specific
fields (`?fields=<name>`).

