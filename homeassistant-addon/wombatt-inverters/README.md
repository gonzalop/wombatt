# Wombatt for Inverters - Home Assistant Add-on

This add-on integrates the [Wombatt](https://github.com/gonzalop/wombatt) application with Home Assistant, allowing you to monitor your inverters directly within your Home Assistant instance.

## Features

*   Monitor various inverter types (PI30, Solark, EG4 18kPV, EG4 6000XP)
*   Publish data to MQTT for Home Assistant integration
*   Support for Modbus RTU and Modbus TCP protocols
*   Configurable polling intervals and timeouts

## Installation

To install this add-on, you need to add its GitHub repository as a custom add-on repository in Home Assistant.

1.  **Add the Repository:**
    *   In Home Assistant, navigate to **Settings** → **Add-ons**.
    *   Go to the **Add-on Store** tab.
    *   Click the three dots in the top right corner and select **Repositories**.
    *   Enter the URL: `https://github.com/gonzalop/wombatt`
    *   Click **Add**, then close the dialog and refresh the page.

2.  **Install the Add-on:**
    *   In the **Add-on Store**, find "Wombatt for Inverters" under "Custom repositories".
    *   Click on it and then click **Install**.

## Configuration

After installation, go to the **Configuration** tab of the add-on.

### Common Options

*   `log_level`: Set the logging level (debug, info, warn, error). Default: `info`
*   `mqtt_broker`: MQTT broker address (e.g., `core-mosquitto:1883`). Default: `core-mosquitto:1883`
*   `mqtt_user`: MQTT username (optional).
*   `mqtt_password`: MQTT password (optional).
*   `mqtt_topic_prefix`: MQTT topic prefix for Home Assistant discovery. **Must be `homeassistant`** for automatic sensor discovery. Default: `homeassistant`

### Inverter Options

*   `baud_rate`: Baud rate for serial communication. Default: `2400`
*   `data_bits`: Number of data bits for serial port. Default: `8`
*   `stop_bits`: Number of stop bits for serial port. Default: `1`
*   `parity`: Parity for serial port (N, E, O). Default: `N`
*   `modbus_id`: Modbus slave ID (only used for solark, eg4_18kpv, and eg4_6000xp inverters). Default: `1`
*   `monitors`: Comma-separated list of monitors. Format: `<device>,<commands>,<mqtt_prefix>[,<inverter_type>]`
    *   The `mqtt_prefix` in this string will prefix all sensor names for this inverter (e.g., `eg4_1` creates sensors like `eg4_1_battery_voltage`)
    *   Example: `/dev/ttyUSB0,QPIRI:QPGS1,eg4_1,pi30`
*   `protocol`: Communication protocol (auto, ModbusRTU, ModbusTCP). Default: `auto`
*   `poll_interval`: Time between polling cycles. Default: `10s`
*   `read_timeout`: Timeout when reading from devices. Default: `5s`
*   `device_type`: Device type (serial, hidraw, tcp). Default: `serial`

## Supported Inverter Types

*   **pi30**: PI30 protocol inverters
*   **solark**: Solark inverters (uses Modbus)
*   **eg4_18kpv**: EG4 18kPV inverters (uses Modbus)
*   **eg4_6000xp**: EG4 6000XP inverters (uses Modbus)

## Usage

After configuring the add-on, start it from the **Info** tab. Check the **Logs** tab for any errors and to verify that data is being published to MQTT.

Your inverter data will appear as sensors in Home Assistant. Sensors will be named using the `mqtt_prefix` specified in the monitors configuration (e.g., if you use `eg4_1` as the prefix, sensors will be named like `eg4_1_battery_voltage`, `eg4_1_ac_output_power`).
