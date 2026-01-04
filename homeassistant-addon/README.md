# Wombatt Home Assistant Add-on

This add-on integrates the [Wombatt](https://github.com/gonzalop/wombatt) application with Home Assistant, allowing you to monitor your inverters and batteries directly within your Home Assistant instance.

## Features

*   Monitor various inverter types (e.g., PI30, Solark)
*   Monitor different BMS types (e.g., EG4LLv2, LifePower4)
*   Publish data to MQTT for Home Assistant integration
*   Optional web server for data visualization

## Installation

To install this add-on, you need to add its GitHub repository as a custom add-on repository in Home Assistant.

1.  **Add the Repository:**
    *   In Home Assistant, navigate to **Supervisor** (or **Settings** -> **Add-ons**).
    *   Go to the **Add-on Store** tab.
    *   Click the three dots in the top right corner and select **Repositories**.
    *   Enter the URL of this GitHub repository (e.g., `https://github.com/gonzalop/wombatt`) and click **Add**.
    *   Close the repositories dialog and refresh the page.

2.  **Install the Add-on:**
    *   In the **Add-on Store**, find "Wombatt Home Assistant Add-on" under "Custom repositories".
    *   Click on it and then click **Install**.

## Configuration

After installation, go to the **Configuration** tab of the add-on. You will find various options to configure the add-on's behavior.

### Mode Selection

The `mode` option determines whether the add-on will monitor inverters or batteries.

*   `inverters`: For monitoring inverter devices.
*   `batteries`: For monitoring battery BMS systems.

### Common Options

These options apply to both `inverters` and `batteries` modes:

*   `log_level`: Set the logging level (debug, info, warn, error).
*   `mqtt_broker`: MQTT broker address (e.g., `core-mqtt`).
*   `mqtt_user`: MQTT username.
*   `mqtt_password`: MQTT password.
*   `mqtt_topic_prefix`: MQTT topic prefix for published data.
*   `poll_interval`: Time between polling cycles (e.g., `10s`).
*   `read_timeout`: Timeout for reading from devices (e.g., `5s`).
*   `web_server_address`: Address for the optional web server (e.g., `0.0.0.0:8080`).
*   `device_type`: Type of device connection (serial, hidraw, tcp).
*   `protocol`: Modbus protocol to use (auto, ModbusRTU, ModbusTCP, lifepower4).

### Inverter Specific Options (when `mode` is `inverters`)

*   `baud_rate`: Baud rate for serial ports (e.g., `2400`).
*   `data_bits`: Number of data bits for serial port (e.g., `8`).
*   `stop_bits`: Number of stop bits for serial port (e.g., `1`).
*   `parity`: Parity for serial port (N, E, O).
*   `monitors`: A list of strings, each defining a device to monitor. Format: `<device>,<command1[:command2:command3...]>,<mqtt_prefix>[,<inverter_type>]`.
    *   Example: `/dev/ttyUSB0,QPIRI:QPGS1,eg4_1,pi30`
*   `modbus_id`: Modbus slave ID (only used for solark, eg4_18kpv, and eg4_6000xp inverters, not pi30).

### Battery Specific Options (when `mode` is `batteries`)

*   `address`: Serial port attached to the batteries (e.g., `/dev/ttyUSB0`).
*   `baud_rate`: Baud rate for serial ports (e.g., `9600`).
*   `battery_ids`: A list of battery IDs to monitor (e.g., `[1, 2]`).
*   `bms_type`: Type of BMS (EG4LLv2, lifepower4, lifepowerv2, pacemodbus).
*   `mqtt_prefix`: MQTT prefix for battery fields (e.g., `eg4`).

## Usage

After configuring the add-on, start it from the **Info** tab. Check the **Logs** tab for any errors and to verify that data is being published.
