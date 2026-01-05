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
#   COMMON_LOG_LEVEL='info'
#   COMMON_MQTT_BROKER=''
#   COMMON_MQTT_USER=''
#   COMMON_MQTT_PASSWORD=''
#   COMMON_MQTT_TOPIC_PREFIX='homeassistant'
#   INVERTERS_BAUD_RATE=2400
#   ...and so on...
eval "$(jq -r 'path(.. | scalars) as $p | ($p | map(ascii_upcase) | join("_")) + "=" + (getpath($p) | @sh)' ${CONFIG_PATH})"

if [ "${MQTT_CONFIG_AUTO_DISCOVERY}" = "true" ]; then
    echo "MQTT Auto-discovery enabled. Attempting to fetch configuration..."
    if [ -n "${SUPERVISOR_TOKEN}" ]; then
        SERVICE_RESULT=$(curl -s -H "Authorization: Bearer ${SUPERVISOR_TOKEN}" -H "Content-Type: application/json" http://supervisor/services/mqtt)
        if [ "$(echo "${SERVICE_RESULT}" | jq -r .result)" = "ok" ]; then
            echo "MQTT configuration found via Supervisor."
            BROKER_HOST=$(echo "${SERVICE_RESULT}" | jq -r .data.host)
            BROKER_PORT=$(echo "${SERVICE_RESULT}" | jq -r .data.port)
            BROKER_USER=$(echo "${SERVICE_RESULT}" | jq -r .data.username)
            BROKER_PASS=$(echo "${SERVICE_RESULT}" | jq -r .data.password)
            BROKER_SSL=$(echo "${SERVICE_RESULT}" | jq -r .data.ssl)

            if [ "${BROKER_SSL}" = "true" ]; then
                echo "Warning: SSL auto-configuration not fully implemented, using plain port if possible or assuming proxy handles it."
                # Note: If SSL is required, we might need adjustments to wombatt args.
                # Wombatt currently takes a broker address.
            fi

            MQTT_CONFIG_ADDRESS="${BROKER_HOST}:${BROKER_PORT}"
            MQTT_CONFIG_MQTT_USER="${BROKER_USER}"
            MQTT_CONFIG_MQTT_PASSWORD="${BROKER_PASS}"
        else
            echo "Could not fetch MQTT configuration from Supervisor: $(echo "${SERVICE_RESULT}" | jq -r .message)"
        fi
    else
        echo "SUPERVISOR_TOKEN not set, cannot perform auto-discovery."
    fi
else
    echo "MQTT Auto-discovery disabled."
fi

WOMBATT_CMD="monitor-inverters"
WOMBATT_ARGS=""

# Common options
if [ -n "${LOGGING_LEVEL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --log-level ${LOGGING_LEVEL}"
fi
if [ -n "${MQTT_CONFIG_ADDRESS}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-broker ${MQTT_CONFIG_ADDRESS}"
fi
if [ -n "${MQTT_CONFIG_MQTT_USER}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-user ${MQTT_CONFIG_MQTT_USER}"
fi
if [ -n "${MQTT_CONFIG_MQTT_PASSWORD}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-password ${MQTT_CONFIG_MQTT_PASSWORD}"
fi
if [ -n "${MQTT_CONFIG_TOPIC_PREFIX}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --mqtt-topic-prefix ${MQTT_CONFIG_TOPIC_PREFIX}"
fi

if [ -n "${INVERTERS_POLL_INTERVAL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --poll-interval ${INVERTERS_POLL_INTERVAL}"
fi
if [ -n "${INVERTERS_READ_TIMEOUT}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --read-timeout ${INVERTERS_READ_TIMEOUT}"
fi
if [ -n "${INVERTERS_DEVICE_TYPE}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --device-type ${INVERTERS_DEVICE_TYPE}"
fi
if [ -n "${INVERTERS_BAUD_RATE}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --baud-rate ${INVERTERS_BAUD_RATE}"
fi
if [ -n "${INVERTERS_DATA_BITS}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --data-bits ${INVERTERS_DATA_BITS}"
fi
if [ -n "${INVERTERS_STOP_BITS}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --stop-bits ${INVERTERS_STOP_BITS}"
fi
if [ -n "${INVERTERS_PARITY}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --parity ${INVERTERS_PARITY}"
fi
if [ -n "${INVERTERS_MODBUS_ID}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --modbus-id ${INVERTERS_MODBUS_ID}"
fi
if [ -n "${INVERTERS_PROTOCOL}" ]; then
  WOMBATT_ARGS="${WOMBATT_ARGS} --protocol ${INVERTERS_PROTOCOL}"
fi

if [ -n "${INVERTERS_MONITORS}" ]; then
  # shellcheck disable=SC2001
  MONITORS_LIST=$(echo "${INVERTERS_MONITORS}" | tr ',' ' ')
  for monitor in ${MONITORS_LIST}; do
    WOMBATT_ARGS="${WOMBATT_ARGS} ${monitor}"
  done
fi

echo "Running: /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}"
exec /usr/bin/wombatt ${WOMBATT_CMD} ${WOMBATT_ARGS}
