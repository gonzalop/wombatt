### forward
`forward` read/writes between 2 ports, displaying the information exchanged in hexadecimal.

### Examples
~~~
$ ./wombatt forward --subordinate-port /dev/MasterBattery --controller-port /dev/Inverter
2023/09/17 12:10:11.831759 Inverter: 8 010300130010b5c3
2023/09/17 12:10:11.875255 MasterBattery: 32 01032000660000003114ad05c8001e753072d8ea6002040000000a0000000015
2023/09/17 12:10:11.882400 MasterBattery: 5 e000004a8e
~~~


