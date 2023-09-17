![wombatt logo](https://github.com/gonzalop/wombatt/blob/main/extras/wombatt-small.jpg?raw=true)
# wombatt

wombatt is a set of tools to monitor batteries and inverters, and to send commands to inverters.
The initial version only supports EG4LLv2 batteries and any inverter using the PI30 protocol family.

## Compilation from Source

To compile wombatt, you need a working Go setup. Then check out the project and run `make` to compile the wombatt binary:

~~~
$ git clone https://github.com/gonzalop/wombatt.git
$ cd wombatt
$ make
~~~

And you'll get a `wombatt` binary.

If you want to cross-compile for linux, windows, and Mac:

~~~
$ git clone https://github.com/gonzalop/wombatt.git
$ cd wombatt
$ make -f Makefile.release release
~~~

And you'll get the different binaries under `build/` and tarfiles under `releases/`.

## Releases
Get binary releases at https://wombatt.cc/releases/

## Docker images
Docker images are available at https://hub.docker.com/r/gonzalomono/wombatt.

Use any recent release tag or `latest` for docker image tag:

~~~
$ docker pull docker.io/gonzalomono/wombatt:latest
$ docker run --device /dev/ttyS1:/dev/ttyS1 -t gonzalomono/wombatt inverter-query --serial-ports /dev/ttyS1 --commands Q1
~~~

## Subcommands and usage
Run `wombatt <subcommand> -h` for help on any specific command.

### battery-info
`battery-info` displays battery status information.

For instance, to query the battery with ID #2:
~~~
$ ./wombatt battery-info --serial-port /dev/ttyUSB0 --battery-ids 2
Battery #2
===========
battery voltage: 52.9V
current: 2.5A
cell 1 voltage: 3.305V
cell 2 voltage: 3.306V
cell 3 voltage: 3.306V
cell 4 voltage: 3.306V
cell 5 voltage: 3.306V
cell 6 voltage: 3.307V
cell 7 voltage: 3.306V
cell 8 voltage: 3.307V
cell 9 voltage: 3.306V
cell 10 voltage: 3.307V
cell 11 voltage: 3.306V
cell 12 voltage: 3.306V
cell 13 voltage: 3.306V
cell 14 voltage: 3.307V
cell 15 voltage: 3.306V
cell 16 voltage: 3.307V
pcb temp: 31°C
max temp: 32°C
avg temp: 30°C
cap remaining: 46%
max charging current: 100A
soh: 100%
soc: 47%
status: inactive/charging
warning: 0
protection: 0
error code: 0
cycle counts: 10
full capacity: 100000mAh
temp1: 30°C
temp2: 30°C
temp3: 30°C
temp4: 30°C
temp5: 0
temp6: 0
cell num: 16
designed capacity: 100Ah
cell balance status: 0
max cell voltage: 3.307V
min cell voltage: 3.305V
mean cell voltage: 3.306V
median cell voltage: 3.306V
model: LFP-51.2V100Ah-V1.0
firmware version: Z02T04
serial: 2022-10-26
~~~

If `battery-ids` is omitted, it will scan IDs from 1 to 64.

### inverter-query
`inverter-query` sends PI30 protocol commands to inverters.

Below, an example of running a single command:
~~~
$ ./wombatt inverter-query -p /dev/ttyS1 --commands Q1
2023/09/17 11:35:13.619499 [00001 00006 00 00 07 037 039 043 038 02 00 000 0036 0000 0000 60.00 11 0 060 030 120 030 58.40 000 120 0 0000]
Device: /dev/ttyS1, Command: Q1
========================================
Time until the end of absorb charging: 1s
Time until the end of float charging: 6s
SCC flags: Not communicating
SCC PWM temperature: 37°C
Inverter temperature: 39°C
Battery temperature: 43°C
Transformer temperature: 38°C
GPIO13: 2
Fan lock status: not locked
Fan PWM speed: 36%
SCC charge power: 0W
Parallel warning: 0
Sync frequency: 60Hz
Inverter charger status: bulk stage
~~~

### monitor-inverters
`monitor-inverters` monitors inverters using PI30 protocol, MQTT publishing optional.

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

### monitor-batteries
`monitor-batteries` monitors batteries state, MQTT publishing optional.

For instance, to monitor batteries with IDs 2 thru 6, and publish to MQTT and a local web page on port 8000, you can run:
~~~
$ ./wombatt monitor-batteries -w :8000 -p /dev/ttyUSB1 --mqtt-broker tcp://127.0.0.1:1883 --mqtt-user youruser --mqtt-password yourpassword --battery-ids 2,3,4,5,6
~~~

The default prefix for the items added to MQTT is `eg4` (i.e., `homeassistant/eg4_battery2_info/...`).

The same infomation is made available via web on port 8000
(http://127.0.0.1:8000/battery/2 and so on) as text or JSON (add
`?format=json` to the URL), with the ability to request specific
fields (`?fields=<name>`).

### modbus-read
`modbus-read` reads registers from a specified device. This is used during development.

To read 38 registers from device ID #2 starting at address 0:
~~~
$ ./wombatt modbus-read -p /dev/ttyUSB1 --id 2 --start 0 --count 38
/dev/ttyUSB1:
00000000  02 03 4c 14 af 00 f0 0c  ed 0c ee 0c ed 0c ef 0c  |..L.............|
00000010  ed 0c ee 0c ed 0c ee 0c  ee 0c ef 0c ee 0c ee 0c  |................|
00000020  ed 0c ee 0c ed 0c ee 00  1f 00 20 00 1e 00 30 00  |.......... ...0.|
00000030  64 00 64 00 30 00 01 00  00 00 00 00 00 00 00 00  |d.d.0...........|
00000040  0a 15 75 2a 00 1e 1e 1e  1e 00 00 00 10 03 e8 45  |..u*...........E|
00000050  8d                                                |.|
~~~


### forward
`forward` read/writes between 2 ports, displaying the information exchanged in hexadecimal. This is used during development.

~~~
$ ./wombatt forward --subordinate-port /dev/MasterBattery --controller-port ~/dev/Inverter
2023/09/17 12:10:11.831759 Inverter: 8 010300130010b5c3
2023/09/17 12:10:11.875255 MasterBattery: 32 01032000660000003114ad05c8001e753072d8ea6002040000000a0000000015
2023/09/17 12:10:11.882400 MasterBattery: 5 e000004a8e
~~~

## Reporting bugs and requesting features
Please use https://github.com/gonzalop/wombatt/issues to report any bug, request new features
or support for batteries, inverters, etc.

