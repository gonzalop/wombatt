## modbus-read
`modbus-read` reads registers from a specified device. This is used during development.

### Examples
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

To read the model, firmware, and serial in a EG4v2LL battery that is connected to a device
that uses the Modbus TCP protocol, and then formatting the 3 fields as a string:
~~~
$ ./wombatt modbus-read -T tcp -p 192.168.1.123:502 --id 2 --start 105 --count 23 -o "[24]byte:Model:::string, [6]byte:Firmware version:::string, [16]byte:Serial:::string"
2023/12/18 21:58:42.304161 Opening 192.168.1.123:502...
Model: LFP-51.2V100Ah-V1.0
Firmware_Version: Z02T04
Serial: 2022-10-26
~~~
