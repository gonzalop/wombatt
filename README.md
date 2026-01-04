<div align="center">
  <img src="extras/wombatt-2026-small.webp" alt="wombatt logo" />
</div>

# Wombatt

wombatt is a set of tools to monitor batteries and inverters, send commands to inverters and more.

## Supported Inverters
- PI30 protocol (e.g., EG4-6500EX, EG4-3000s)
- Solark (12k and 15k)
- EG4 18kPV
- EG4 6000XP

## Supported Battery/BMS
- EG4-LL (BMS Type: `EG4LLv2`)
- EG4-LL-S (BMS Type: `lifepower4`)
- EG4 Lifepower (BMS Type: `EG4LLv2`)
- EG4 Lifepower v2 (BMS Type: `lifepowerv2`) (protocol switches: 1-off, 2 thru 6-on)
- Pace BMS Modbus (SOK, Jakiper) (BMS Type: `pacemodbus`)

wombatt can use direct RS232 or RS485 connections, or TCP to communicate using Modbus RTU, Modbus TCP,
and slight variations of Modbus ASCII.

The data can be exposed via console, web server (txt, json), or MQTT (Homeassistant auto-discovery topics automatically added).

## Home Assistant Add-ons

[![Open your Home Assistant instance and show the add add-on repository dialog with a specific repository URL pre-filled.](https://my.home-assistant.io/badges/supervisor_add_addon_repository.svg)](https://my.home-assistant.io/redirect/supervisor_add_addon_repository/?repository_url=https%3A%2F%2Fgithub.com%2Fgonzalop%2Fwombatt)

Wombatt is available as Home Assistant add-ons for easy integration. Click the button above to add the repository to your Home Assistant instance.

## Features
- **Web Dashboard**: A web-based user interface to monitor your inverters and batteries in real-time.
- **Prometheus Metrics**: Expose metrics in a Prometheus-compatible format for easy integration with monitoring systems.
- **YAML Configuration**: Configure wombatt using a YAML file for advanced and flexible setups.

See [wombatt command documentation](docs/cmds/wombatt.md) for more details.

## Commands

- **battery-info**: Displays battery information
- **forward**: Forwards commands between a two devices
- **inverter-query**: Sends PI30 protocol commands to inverters
- **modbus-read**: Reads Modbus holding registers
- **monitor-batteries**: Monitors batteries state, MQTT publishing optional
- **monitor-inverters**: Monitors inverters state, with optional MQTT publishing.

## Releases
Get binary releases at https://github.com/gonzalop/wombatt/releases

## Docker images
Docker images are available at https://hub.docker.com/r/gonzalomono/wombatt.

Use any recent release tag or `latest` for docker image tag:

```
$ docker pull docker.io/gonzalomono/wombatt:latest
$ docker run --device /dev/ttyS1:/dev/ttyS1 -t gonzalomono/wombatt inverter-query -p /dev/ttyS1 -c RealtimeData -I eg4_18kpv
...or...
$ docker pull docker.io/gonzalomono/wombatt:latest
$ docker run --device /dev/ttyS1:/dev/ttyS1 -t gonzalomono/wombatt inverter-query -p /dev/ttyS1 -c Q1  # PI30 inverter.
```

## Compilation from Source

To compile wombatt, you need a working Go setup. Then check out the project and run `make` to compile the wombatt binary:

```
$ git clone https://github.com/gonzalop/wombatt.git
$ cd wombatt
$ make
```

And you'll get a `wombatt` binary.

If you want to cross-compile for linux, windows, and Mac:

```
$ git clone https://github.com/gonzalop/wombatt.git
$ cd wombatt
$ make -f Makefile.release release
```

And you'll get the different binaries under `build/` and tarfiles under `releases/`.

## Video showing how to install and run wombatt on a Raspberry Pi

The following video shows how to install from source and run wombatt to monitor 6 lifepower4 batteries.
Note that the steps to build the binary from source can be skipped if you get the ARM or ARM64 binaries from the
[releases link](https://github.com/gonzalop/wombatt#releases) above.

[Video from AmateurSolarBuild@dyisolarforums:](https://youtu.be/wwLMO1hMxnY)

## Home Assistant Add-on
For instructions on how to install and configure the Home Assistant add-on, see the [Home Assistant Add-on README](homeassistant-addon/README.md).

## Reporting bugs and requesting features
Please use https://github.com/gonzalop/wombatt/issues to report any bug, request new features
or support for batteries, inverters, etc.

