#!/usr/bin/with-contenv bash
# shellcheck disable=SC1091
source /usr/lib/hassio-addons/base/functions.sh

echo "Starting Wombatt..."

WOMBATT_CMD=""
WOMBATT_ARGS=""

# Common options
if [ -n "${ADDON_CONFIG_LOG_LEVEL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --log-level ${ADDON_CONFIG_LOG_LEVEL}"
fi
if [ -n "${ADDON_CONFIG_MQTT_BROKER}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-broker ${ADDON_CONFIG_MQTT_BROKER}"
fi
if [ -n "${ADDON_CONFIG_MQTT_USER}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-user ${ADDON_CONFIG_MQTT_USER}"
fi
if [ -n "${ADDON_CONFIG_MQTT_PASSWORD}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-password ${ADDON_CONFIG_MQTT_PASSWORD}"
fi
if [ -n "${ADDON_CONFIG_MQTT_TOPIC_PREFIX}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-topic-prefix ${ADDON_CONFIG_MQTT_TOPIC_PREFIX}"
fi
if [ -n "${ADDON_CONFIG_POLL_INTERVAL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --poll-interval ${ADDON_CONFIG_POLL_INTERVAL}"
fi
if [ -n "${ADDON_CONFIG_READ_TIMEOUT}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --read-timeout ${ADDON_CONFIG_READ_TIMEOUT}"
fi
if [ -n "${ADDON_CONFIG_WEB_SERVER_ADDRESS}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --web-server-address ${ADDON_CONFIG_WEB_SERVER_ADDRESS}"
fi
if [ -n "${ADDON_CONFIG_DEVICE_TYPE}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --device-type ${ADDON_CONFIG_DEVICE_TYPE}"
fi
if [ -n "${ADDON_CONFIG_PROTOCOL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --protocol ${ADDON_CONFIG_PROTOCOL}"
fi

case "${ADDON_CONFIG_MODE}" in
  "inverters")
    WOMBATT_CMD="monitor-inverters"
    if [ -n "${ADDON_CONFIG_BAUD_RATE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --baud-rate ${ADDON_CONFIG_BAUD_RATE}"
    fi
    if [ -n "${ADDON_CONFIG_DATA_BITS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --data-bits ${ADDON_CONFIG_DATA_BITS}"
    fi
    if [ -n "${ADDON_CONFIG_STOP_BITS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --stop-bits ${ADDON_CONFIG_STOP_BITS}"
    fi
    if [ -n "${ADDON_CONFIG_PARITY}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --parity ${ADDON_CONFIG_PARITY}"
    fi
    if [ -n "${ADDON_CONFIG_MODBUS_ID}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --modbus-id ${ADDON_CONFIG_MODBUS_ID}"
    fi
    # Monitors is a list of strings
    for monitor in ${ADDON_CONFIG_MONITORS}; do
      WOMBATT_ARGS="${WOMBATT_ARGS} ${monitor}"
    done
    ;;
  "batteries")
    WOMBATT_CMD="monitor-batteries"
    if [ -n "${ADDON_CONFIG_ADDRESS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --address ${ADDON_CONFIG_ADDRESS}"
    fi
    if [ -n "${ADDON_CONFIG_BAUD_RATE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --baud-rate ${ADDON_CONFIG_BAUD_RATE}"
    fi
    if [ -n "${ADDON_CONFIG_BMS_TYPE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --bms-type ${ADDON_CONFIG_BMS_TYPE}"
    fi
    if [ -n "${ADDON_CONFIG_MQTT_PREFIX}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-prefix ${ADDON_CONFIG_MQTT_PREFIX}"
    fi
    # Battery IDs is a list of integers
    for id in ${ADDON_CONFIG_BATTERY_IDS}; do
      WOMBATT_ARGS="${WOMBATT_ARGS} --battery-id ${id}"
    done
    ;;
  *)
    echo "Invalid mode: ${ADDON_CONFIG_MODE}. Must be 'inverters' or 'batteries'."
    exit 1
    ;;
esac

exec /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}
