![wombatt logo](https://github.com/gonzalop/wombatt/blob/main/extras/wombatt-small.jpg?raw=true)
# wombatt

wombatt is a set of tools to monitor batteries and inverters, and to send commands to inverters.

Inverters supported:
- Any that can handle the PI30 protocol. Known to work:
    - EG4-6500EX
    - EG4-3000s (unconfirmed)

Battery/BMS supported:
- EG4-LL (--bms-type=EG4LLv2)
- EG4-LL-S (unconfirmed) (--bms-type=lifepower4)
- EG4 Lifepower (--bms-type=EG4LLv2)
- EG4 Lifepower v2 (--bms-type=lifepowerv2) (protocol switches: 1-off, 2 thru 6-on)
- Pace BMS Modbus (SOK, Jakiper) (unconfirmed) (--bms-type=pacemodbus)

wombatt can use direct RS232 or RS485 connections, or TCP to communicate using Modbus RTU, Modbus TCP,
and slight variations of Modbus ASCII.

The data can be exposed via console, web server (txt, json), or MQTT (Homeassistant auto-discovery topics automatically added).


See [wombatt command documentation](docs/cmds/wombatt.md) for more details.

## Releases
Get binary releases at https://wombatt.cc/releases/

## Docker images
Docker images are available at https://hub.docker.com/r/gonzalomono/wombatt.

Use any recent release tag or `latest` for docker image tag:

~~~
$ docker pull docker.io/gonzalomono/wombatt:latest
$ docker run --device /dev/ttyS1:/dev/ttyS1 -t gonzalomono/wombatt inverter-query -p /dev/ttyS1 --commands Q1
~~~

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

## Video showing how to install and run wombatt on a Raspberry Pi

The following video shows how to install from source and run wombatt to monitor 6 lifepower4 batteries.
Note that the steps to build the binary from source can be skipped if you get the ARM or ARM64 binaries from the
[releases link](https://github.com/gonzalop/wombatt#releases) above.

[Video from AmateurSolarBuild@dyisolarforums:](https://youtu.be/wwLMO1hMxnY)

## Reporting bugs and requesting features
Please use https://github.com/gonzalop/wombatt/issues to report any bug, request new features
or support for batteries, inverters, etc.

