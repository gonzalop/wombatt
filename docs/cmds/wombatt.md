# Wombatt

A wanna-be Swiss army knife for inverter and battery monitoring.

### Usage

```
wombatt <command> [flags]
```

### Commands

Run `wombatt <command> --help` for more information on a command.

- **[battery-info](battery-info.md)**: Displays battery information
- **[forward](forward.md)**: Forwards commands between a two devices
- **[inverter-query](inverter-query.md)**: Sends PI30 protocol commands to inverters
- **[modbus-read](modbus-read.md)**: Reads Modbus holding registers
- **[monitor-batteries](monitor-batteries.md)**: Monitors batteries state, MQTT publishing optional
- **[monitor-inverters](monitor-inverters.md)**: Monitors inverters state, with optional MQTT publishing. It can be used with PI30, Solark, EG4 18kPV, or EG4 6000XP Modbus protocols.

### Flags

| Flag | Description | Default |
| --- | --- | --- |
| `-h`, `--help` | Show context-sensitive help. | |
| `-l`, `--log-level` | Set the logging level (debug|info|warn|error) | `info` |
| `-v`, `--version` | Print version information and quit | |
