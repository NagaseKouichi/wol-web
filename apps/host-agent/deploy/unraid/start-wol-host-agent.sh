#!/usr/bin/env sh
set -eu

AGENT_DIR=/boot/config/plugins/wol-host-agent
ENV_FILE="$AGENT_DIR/wol-host-agent.env"
BIN="$AGENT_DIR/wol-host-agent"
LOG_FILE=/var/log/wol-host-agent.log
PID_FILE=/var/run/wol-host-agent.pid

if [ ! -x "$BIN" ]; then
	echo "wol-host-agent binary not found or not executable: $BIN"
	exit 1
fi

if [ ! -f "$ENV_FILE" ]; then
	echo "wol-host-agent env file not found: $ENV_FILE"
	exit 1
fi

if [ -f "$PID_FILE" ] && kill -0 "$(cat "$PID_FILE")" >/dev/null 2>&1; then
	echo "wol-host-agent is already running"
	exit 0
fi

set -a
. "$ENV_FILE"
set +a

nohup "$BIN" >>"$LOG_FILE" 2>&1 &
echo "$!" >"$PID_FILE"
echo "wol-host-agent started with pid $(cat "$PID_FILE")"
