#!/bin/bash
set -eo pipefail

echo "Starting Wombatt..."

CONFIG_PATH=/data/options.json

# Parse configuration and export as environment variables securely
while IFS='=' read -r key value; do
    export "$key"="$value"
done < <(jq -r 'path(.. | scalars) as $p | ($p | map(ascii_upcase) | join("_")) + "=" + (getpath($p) | tostring)' "${CONFIG_PATH}")

if [ "${MQTT_CONFIG_AUTO_DISCOVERY}" = "true" ]; then
    echo "MQTT Auto-discovery enabled. Attempting to fetch configuration..."
    if [ -n "${SUPERVISOR_TOKEN}" ]; then
        SERVICE_RESULT=$(curl -s -H "Authorization: Bearer ${SUPERVISOR_TOKEN}" -H "Content-Type: application/json" http://supervisor/services/mqtt)
        if [ "$(echo "${SERVICE_RESULT}" | jq -r .result)" = "ok" ]; then
            echo "MQTT configuration found via Supervisor."
            BROKER_HOST=$(echo "${SERVICE_RESULT}" | jq -r .data.host)
            BROKER_PORT=$(echo "${SERVICE_RESULT}" | jq -r .data.port)
            export MQTT_USER=$(echo "${SERVICE_RESULT}" | jq -r .data.username)
            export MQTT_PASSWORD=$(echo "${SERVICE_RESULT}" | jq -r .data.password)
            export MQTT_BROKER="tcp://${BROKER_HOST}:${BROKER_PORT}"
        else
            echo "Could not fetch MQTT configuration from Supervisor: $(echo "${SERVICE_RESULT}" | jq -r .message)"
        fi
    else
        echo "SUPERVISOR_TOKEN not set, cannot perform auto-discovery."
    fi
fi

# Fallback to manual config if not set by auto-discovery
[ -z "$MQTT_BROKER" ] && [ -n "$MQTT_CONFIG_ADDRESS" ] && export MQTT_BROKER="$MQTT_CONFIG_ADDRESS"
[ -z "$MQTT_USER" ] && [ -n "$MQTT_CONFIG_MQTT_USER" ] && export MQTT_USER="$MQTT_CONFIG_MQTT_USER"
[ -z "$MQTT_PASSWORD" ] && [ -n "$MQTT_CONFIG_MQTT_PASSWORD" ] && export MQTT_PASSWORD="$MQTT_CONFIG_MQTT_PASSWORD"
[ -n "$MQTT_CONFIG_TOPIC_PREFIX" ] && export MQTT_TOPIC_PREFIX="$MQTT_CONFIG_TOPIC_PREFIX"

WOMBATT_ARGS=("monitor-inverters")

# Common options
if [ -n "${LOGGING_LEVEL}" ]; then
  WOMBATT_ARGS+=("--log-level" "${LOGGING_LEVEL}")
fi

if [ -n "${INVERTERS_POLL_INTERVAL}" ]; then
  WOMBATT_ARGS+=("--poll-interval" "${INVERTERS_POLL_INTERVAL}")
fi
if [ -n "${INVERTERS_READ_TIMEOUT}" ]; then
  WOMBATT_ARGS+=("--read-timeout" "${INVERTERS_READ_TIMEOUT}")
fi
if [ -n "${INVERTERS_DEVICE_TYPE}" ]; then
  WOMBATT_ARGS+=("--device-type" "${INVERTERS_DEVICE_TYPE}")
fi
if [ -n "${INVERTERS_BAUD_RATE}" ]; then
  WOMBATT_ARGS+=("--baud-rate" "${INVERTERS_BAUD_RATE}")
fi
if [ -n "${INVERTERS_DATA_BITS}" ]; then
  WOMBATT_ARGS+=("--data-bits" "${INVERTERS_DATA_BITS}")
fi
if [ -n "${INVERTERS_STOP_BITS}" ]; then
  WOMBATT_ARGS+=("--stop-bits" "${INVERTERS_STOP_BITS}")
fi
if [ -n "${INVERTERS_PARITY}" ]; then
  WOMBATT_ARGS+=("--parity" "${INVERTERS_PARITY}")
fi
if [ -n "${INVERTERS_MODBUS_ID}" ]; then
  WOMBATT_ARGS+=("--modbus-id" "${INVERTERS_MODBUS_ID}")
fi
if [ -n "${INVERTERS_PROTOCOL}" ]; then
  WOMBATT_ARGS+=("--protocol" "${INVERTERS_PROTOCOL}")
fi

if [ -n "${INVERTERS_MONITORS}" ]; then
  # shellcheck disable=SC2001
  MONITORS_LIST=$(echo "${INVERTERS_MONITORS}" | tr ',' ' ')
  for monitor in ${MONITORS_LIST}; do
    WOMBATT_ARGS+=("${monitor}")
  done
fi

# We don't print full args to avoid accidental leakage (even though we moved secrets to ENV)
echo "Running: /usr/bin/wombatt ${WOMBATT_ARGS[0]} (args omitted for security)"
exec /usr/bin/wombatt "${WOMBATT_ARGS[@]}"
