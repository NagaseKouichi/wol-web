#!/bin/bash
set -euo pipefail

PLUGIN="wol-host-agent"
PLUGIN_DIR="/boot/config/plugins/${PLUGIN}"
CFG="${PLUGIN_DIR}/${PLUGIN}.cfg"
ENV_FILE="${PLUGIN_DIR}/${PLUGIN}.env"

read_cfg_value() {
	local key="$1"
	local fallback="$2"

	if [[ -f "$CFG" ]]; then
		local value
		value="$(grep -E "^${key}=" "$CFG" | tail -n 1 | cut -d= -f2- | sed -e 's/^"//' -e 's/"$//')"
		if [[ -n "$value" ]]; then
			echo "$value"
			return
		fi
	fi

	echo "$fallback"
}

mkdir -p "$PLUGIN_DIR"
chmod 700 "$PLUGIN_DIR"

enabled="$(read_cfg_value ENABLED yes)"
token="$(read_cfg_value TOKEN "")"
port="$(read_cfg_value PORT 8765)"

{
	printf 'HOST_AGENT_TOKEN=%q\n' "$token"
	printf 'HOST_AGENT_PORT=%q\n' "$port"
	printf 'HOST_AGENT_SHUTDOWN_CMD=%q\n' "powerdown"
	printf 'HOST_AGENT_REBOOT_CMD=%q\n' "reboot"
	printf 'HOST_AGENT_SLEEP_CMD=%q\n' "echo -n mem > /sys/power/state"
} > "$ENV_FILE"
chmod 600 "$ENV_FILE"

if [[ -x /etc/rc.d/rc.wol-host-agent ]]; then
	if [[ "$enabled" == "yes" ]]; then
		/etc/rc.d/rc.wol-host-agent restart
	else
		/etc/rc.d/rc.wol-host-agent stop
	fi
fi
