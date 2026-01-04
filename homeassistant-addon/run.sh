#!/bin/bash

echo "Starting Wombatt..."

CONFIG_PATH=/data/options.json
# ALL_OPTIONS=$(cat $CONFIG_PATH)
# echo "Options: ${ALL_OPTIONS}" >&2

WOMBATT_CMD=""
WOMBATT_ARGS=""

# Extract Common options
MODE=$(jq --raw-output '.mode // empty' $CONFIG_PATH)
COMMON_LOG_LEVEL=$(jq --raw-output '.common.log_level // empty' $CONFIG_PATH)
COMMON_MQTT_BROKER=$(jq --raw-output '.common.mqtt_broker // empty' $CONFIG_PATH)
COMMON_MQTT_USER=$(jq --raw-output '.common.mqtt_user // empty' $CONFIG_PATH)
COMMON_MQTT_PASSWORD=$(jq --raw-output '.common.mqtt_password // empty' $CONFIG_PATH)
COMMON_MQTT_TOPIC_PREFIX=$(jq --raw-output '.common.mqtt_topic_prefix // empty' $CONFIG_PATH)

if [ -n "${COMMON_LOG_LEVEL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --log-level ${COMMON_LOG_LEVEL}"
fi
if [ -n "${COMMON_MQTT_BROKER}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-broker ${COMMON_MQTT_BROKER}"
fi
if [ -n "${COMMON_MQTT_USER}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-user ${COMMON_MQTT_USER}"
fi
if [ -n "${COMMON_MQTT_PASSWORD}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-password ${COMMON_MQTT_PASSWORD}"
fi
if [ -n "${COMMON_MQTT_TOPIC_PREFIX}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-topic-prefix ${COMMON_MQTT_TOPIC_PREFIX}"
fi

case "${MODE}" in
  "inverters")
    WOMBATT_CMD="monitor-inverters"

    # Extract Inverter options
    POLL_INTERVAL=$(jq --raw-output '.inverters.poll_interval // empty' $CONFIG_PATH)
    READ_TIMEOUT=$(jq --raw-output '.inverters.read_timeout // empty' $CONFIG_PATH)
    WEB_SERVER_ADDRESS=$(jq --raw-output '.inverters.web_server_address // empty' $CONFIG_PATH)
    DEVICE_TYPE=$(jq --raw-output '.inverters.device_type // empty' $CONFIG_PATH)
    BAUD_RATE=$(jq --raw-output '.inverters.baud_rate // empty' $CONFIG_PATH)
    DATA_BITS=$(jq --raw-output '.inverters.data_bits // empty' $CONFIG_PATH)
    STOP_BITS=$(jq --raw-output '.inverters.stop_bits // empty' $CONFIG_PATH)
    PARITY=$(jq --raw-output '.inverters.parity // empty' $CONFIG_PATH)
    MODBUS_ID=$(jq --raw-output '.inverters.modbus_id // empty' $CONFIG_PATH)
    PROTOCOL=$(jq --raw-output '.inverters.protocol // empty' $CONFIG_PATH)
    MONITORS=$(jq --raw-output '.inverters.monitors // empty' $CONFIG_PATH)

    if [ -n "${POLL_INTERVAL}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --poll-interval ${POLL_INTERVAL}"
    fi
    if [ -n "${READ_TIMEOUT}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --read-timeout ${READ_TIMEOUT}"
    fi
    if [ -n "${WEB_SERVER_ADDRESS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --web-server-address ${WEB_SERVER_ADDRESS}"
    fi
    if [ -n "${DEVICE_TYPE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --device-type ${DEVICE_TYPE}"
    fi
    if [ -n "${BAUD_RATE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --baud-rate ${BAUD_RATE}"
    fi
    if [ -n "${DATA_BITS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --data-bits ${DATA_BITS}"
    fi
    if [ -n "${STOP_BITS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --stop-bits ${STOP_BITS}"
    fi
    if [ -n "${PARITY}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --parity ${PARITY}"
    fi
    if [ -n "${MODBUS_ID}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --modbus-id ${MODBUS_ID}"
    fi
    if [ -n "${PROTOCOL}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --protocol ${PROTOCOL}"
    fi

    if [ -n "${MONITORS}" ]; then
      # shellcheck disable=SC2001
      MONITORS_LIST=$(echo "${MONITORS}" | tr ',' ' ')
      for monitor in ${MONITORS_LIST}; do
        WOMBATT_ARGS="${WOMBATT_ARGS} ${monitor}"
      done
    fi
    ;;
  "batteries")
    WOMBATT_CMD="monitor-batteries"

    # Extract Battery options
    POLL_INTERVAL=$(jq --raw-output '.batteries.poll_interval // empty' $CONFIG_PATH)
    READ_TIMEOUT=$(jq --raw-output '.batteries.read_timeout // empty' $CONFIG_PATH)
    WEB_SERVER_ADDRESS=$(jq --raw-output '.batteries.web_server_address // empty' $CONFIG_PATH)
    DEVICE_TYPE=$(jq --raw-output '.batteries.device_type // empty' $CONFIG_PATH)
    ADDRESS=$(jq --raw-output '.batteries.address // empty' $CONFIG_PATH)
    BAUD_RATE=$(jq --raw-output '.batteries.baud_rate // empty' $CONFIG_PATH)
    BMS_TYPE=$(jq --raw-output '.batteries.bms_type // empty' $CONFIG_PATH)
    MQTT_PREFIX=$(jq --raw-output '.batteries.mqtt_prefix // empty' $CONFIG_PATH)
    PROTOCOL=$(jq --raw-output '.batteries.protocol // empty' $CONFIG_PATH)
    BATTERY_IDS=$(jq --raw-output '.batteries.battery_ids // empty' $CONFIG_PATH)

    if [ -n "${POLL_INTERVAL}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --poll-interval ${POLL_INTERVAL}"
    fi
    if [ -n "${READ_TIMEOUT}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --read-timeout ${READ_TIMEOUT}"
    fi
    if [ -n "${WEB_SERVER_ADDRESS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --web-server-address ${WEB_SERVER_ADDRESS}"
    fi
    if [ -n "${DEVICE_TYPE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --device-type ${DEVICE_TYPE}"
    fi
    if [ -n "${ADDRESS}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --address ${ADDRESS}"
    fi
    if [ -n "${BAUD_RATE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --baud-rate ${BAUD_RATE}"
    fi
    if [ -n "${BMS_TYPE}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --bms-type ${BMS_TYPE}"
    fi
    if [ -n "${MQTT_PREFIX}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-prefix ${MQTT_PREFIX}"
    fi
    if [ -n "${PROTOCOL}" ]; then
      WOMBATT_ARGS="${WOMBATT_ARGS} --protocol ${PROTOCOL}"
    fi

    if [ -n "${BATTERY_IDS}" ]; then
      # shellcheck disable=SC2001
      IDS=$(echo "${BATTERY_IDS}" | tr ',' ' ')
      for id in ${IDS}; do
        WOMBATT_ARGS="${WOMBATT_ARGS} --battery-id ${id}"
      done
    fi
    ;;
  *)
    echo "Invalid mode: ${MODE}. Must be 'inverters' or 'batteries'."
    exit 1
    ;;
esac

echo "Running: /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}"
exec /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}
