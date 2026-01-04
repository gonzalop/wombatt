#!/bin/bash

echo "Starting Wombatt..."

CONFIG_PATH=/data/options.json

# Sets environment variables to the configuration options set in CONFIG_PATH.
#
#   - path(.. | scalars): This recursively finds the path to every "leaf" node (strings, numbers, booleans)
#   - map(ascii_upcase) | join("_"): This takes the path (e.g., ["batteries", "baud_rate"]) and turns it into a shell-friendly string (BATTERIES_BAUD_RATE).
#   - getpath($p): This retrieves the actual value at that path.
#   - @sh: automatically escapes and quotes the values so they are safe for the shell to evaluate.
#
# The output ends up being something like:
#
#   MODE='inverters'
#   COMMON_LOG_LEVEL='info'
#   COMMON_MQTT_BROKER=''
#   COMMON_MQTT_USER=''
#   COMMON_MQTT_PASSWORD=''
#   COMMON_MQTT_TOPIC_PREFIX='homeassistant'
#   BATTERIES_BAUD_RATE=9600
#   ...and so on...
eval "$(jq -r 'path(.. | scalars) as $p | ($p | map(ascii_upcase) | join("_")) + "=" + (getpath($p) | @sh)' ${CONFIG_PATH})"

WOMBATT_CMD="monitor-batteries"
WOMBATT_ARGS=""

# Common options
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


if [ -n "${BATTERIES_POLL_INTERVAL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --poll-interval ${BATTERIES_POLL_INTERVAL}"
fi
if [ -n "${BATTERIES_READ_TIMEOUT}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --read-timeout ${BATTERIES_READ_TIMEOUT}"
fi
if [ -n "${BATTERIES_DEVICE_TYPE}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --device-type ${BATTERIES_DEVICE_TYPE}"
fi
if [ -n "${BATTERIES_ADDRESS}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --address ${BATTERIES_ADDRESS}"
fi
if [ -n "${BATTERIES_BAUD_RATE}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --baud-rate ${BATTERIES_BAUD_RATE}"
fi
if [ -n "${BATTERIES_BMS_TYPE}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --bms-type ${BATTERIES_BMS_TYPE}"
fi
if [ -n "${BATTERIES_MQTT_PREFIX}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-prefix ${BATTERIES_MQTT_PREFIX}"
fi
if [ -n "${BATTERIES_PROTOCOL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --protocol ${BATTERIES_PROTOCOL}"
fi

if [ -n "${BATTERIES_BATTERY_IDS}" ]; then
  # shellcheck disable=SC2001
  IDS=$(echo "${BATTERIES_BATTERY_IDS}" | tr ',' ' ')
  for id in ${IDS}; do
    WOMBATT_ARGS="${WOMBATT_ARGS} --battery-id ${id}"
  done
fi


echo "Running: /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}"
exec /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}
