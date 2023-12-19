## inverter-query
`inverter-query` sends PI30 protocol commands to inverters.

### Examples
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

