## forward
`forward` read/writes between 2 ports, displaying the information exchanged in hexadecimal.

### Usage

```
wombatt forward --controller-port=STRING --subordinate-port=STRING [flags]
```

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
| `--controller-port` | Serial port or address of the controller | |
| `--subordinate-port` | Serial port or address of the subordinate device | |
| `-B`, `--baud-rate` | Baud rate | `9600` |
| `-T`, `--device-type` | One of serial,hidraw,tcp | `serial` |

### Examples

**Forward between two serial ports:**

```
$ ./wombatt forward --subordinate-port /dev/MasterBattery --controller-port /dev/Inverter
2023/09/17 12:10:11.831759 Inverter: 8 010300130010b5c3
2023/09/17 12:10:11.875255 MasterBattery: 32 01032000660000003114ad05c8001e753072d8ea6002040000000a0000000015
2023/09/17 12:10:11.882400 MasterBattery: 5 e000004a8e
```

**Forward between two TCP ports:**

```
$ ./wombatt forward --subordinate-port 192.168.1.100:502 --controller-port 192.168.1.101:502 --device-type tcp
```