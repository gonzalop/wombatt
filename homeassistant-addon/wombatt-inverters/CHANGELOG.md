# Changelog

## 0.5.1 - 2026-02-08

### ✨ New Features
- **Optimized MQTT communication**: Switched to a new MQTT library and implemented Home Assistant abbreviations for discovery keys, significantly reducing the size of discovery messages.
- **Improved MQTT performance**: Optimized the use of topic aliases and message retention for more efficient data publishing.

### 🛠 Maintenance
- **Internal code modernization**: Updated the codebase to use modern Go features (Go 1.22+).
- **Dependency updates**: Updated internal and external dependencies.

## 0.5.0 - 2026-01-05

### 🚀 Major Changes

#### Add-on Split
The Home Assistant add-on has been split into two dedicated add-ons:
- **Wombatt for Batteries**: For monitoring BMS systems (EG4, LifePower4, etc.).
- **Wombatt for Inverters**: For monitoring inverters (Solark, EG4, PI30, etc.).

#### Configuration Refactor
The configuration structure has been significantly refactored for better organization and usability.

- **`common` section removed**: Options have been moved to dedicated sections.
- **New `mqtt_config` section**:
    - `mqtt_broker` is now `address`.
    - `mqtt_user` and `mqtt_password` are kept as is (to prevent browser autofill issues).
    - `auto_discovery` option added.
- **New `logging` section**:
    - `log_level` is now `level` under this section.

### ✨ New Features

- **MQTT Auto-Discovery**: Both add-ons now support automatic discovery of MQTT settings from Home Assistant. This is enabled by default.
    - *Note*: SSL is currently not supported for auto-discovery.
    - *Note*: If the Home Assistant MQTT service is found, it will overwrite manual `address`, `user`, and `password` settings.


## 0.0.16 - 2026-01-04

Initial add-on release.
