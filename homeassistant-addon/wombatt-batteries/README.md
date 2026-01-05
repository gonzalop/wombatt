# Wombatt for Batteries - Home Assistant Add-on

This add-on integrates the [Wombatt](https://github.com/gonzalop/wombatt) application with Home Assistant, allowing you to monitor your battery BMS systems directly within your Home Assistant instance.

## Features

*   Monitor different BMS types (EG4LLv2, LifePower4, LifePowerv2, Pace BMS)
*   Publish data to MQTT for Home Assistant integration
*   Support for multiple batteries on the same bus
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
    *   In the **Add-on Store**, find "Wombatt for Batteries" under "Custom repositories".
    *   Click on it and then click **Install**.

## Configuration

After installation, go to the **Configuration** tab of the add-on.

### MQTT Configuration
*   `auto_discovery`: Automatically discover MQTT settings from Home Assistant. If enabled and the Home Assistant MQTT service is found, it will overwrite the `address`, `user`, and `password` fields. Note that SSL is not supported at this time.
*   `address`: MQTT broker address (e.g., `core-mosquitto:1883`). Default: `core-mosquitto:1883`
*   `mqtt_user`: MQTT username (optional).
*   `mqtt_password`: MQTT password (optional).
*   `topic_prefix`: MQTT topic prefix for Home Assistant discovery. **Must be `homeassistant`** for automatic sensor discovery. Default: `homeassistant`

### Logging
*   `level`: Set the logging level (debug, info, warn, error). Default: `info`

### Battery Options

*   `address`: Device address (serial port or IP:port). Default: `/dev/ttyUSB0`
*   `baud_rate`: Baud rate for serial communication. Default: `9600`
*   `bms_type`: BMS type (EG4LLv2, lifepower4, lifepowerv2, pacemodbus). Default: `EG4LLv2`
*   `mqtt_prefix`: Prefix for all sensor names (e.g., `eg4` creates sensors like `eg4_battery_1_voltage`). Default: `eg4`
*   `battery_ids`: Comma-separated list of battery IDs to monitor (e.g., `1,2,3`).
*   `protocol`: Communication protocol (auto, lifepower4, lifepowerv2, pacemodbus). Default: `auto`
*   `poll_interval`: Time between polling cycles. Default: `10s`
*   `read_timeout`: Timeout when reading from devices. Default: `500ms`
*   `device_type`: Device type (serial, hidraw, tcp). Default: `serial`

## Supported BMS Types

*   **EG4LLv2**: EG4 LifePower4 batteries (newer protocol)
*   **lifepower4**: EG4 LifePower4 batteries (original protocol)
*   **lifepowerv2**: EG4 LifePower batteries (version 2 protocol)
*   **pacemodbus**: Pace BMS (used in SOK, Jakiper batteries)

## Usage

After configuring the add-on, start it from the **Info** tab. Check the **Logs** tab for any errors and to verify that data is being published to MQTT.

Your battery data will appear as sensors in Home Assistant. Each battery will have its own set of sensors named using the pattern `{mqtt_prefix}_battery{id}_{sensor_name}` (e.g., `eg4_battery2_voltage`, `eg4_battery3_soc`).

## Example Configuration

For monitoring two EG4LLv2 batteries:

```yaml
battery_ids: "2,3"
bms_type: EG4LLv2
address: /dev/ttyUSB0
```
