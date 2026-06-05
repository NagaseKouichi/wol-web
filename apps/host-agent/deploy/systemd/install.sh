#!/usr/bin/env sh
set -eu

SCRIPT_DIR=$(CDPATH= cd -- "$(dirname -- "$0")" && pwd)
AGENT_DIR=$(CDPATH= cd -- "$SCRIPT_DIR/../.." && pwd)

if [ "$(id -u)" -ne 0 ]; then
	echo "Please run as root: sudo $0"
	exit 1
fi

if ! command -v go >/dev/null 2>&1; then
	echo "go is required to build wol-host-agent"
	exit 1
fi

cd "$AGENT_DIR"
go build -o /usr/local/bin/wol-host-agent .
chmod 0755 /usr/local/bin/wol-host-agent

if [ ! -f /etc/wol-host-agent.env ]; then
	cp "$SCRIPT_DIR/wol-host-agent.env.example" /etc/wol-host-agent.env
	chmod 0600 /etc/wol-host-agent.env
	echo "Created /etc/wol-host-agent.env. Edit HOST_AGENT_TOKEN before starting the service."
fi

cp "$SCRIPT_DIR/wol-host-agent.service" /etc/systemd/system/wol-host-agent.service
systemctl daemon-reload
systemctl enable wol-host-agent.service

echo "Installed wol-host-agent."
echo "Next:"
echo "  1. Edit /etc/wol-host-agent.env"
echo "  2. Run: systemctl start wol-host-agent.service"
echo "  3. Check: systemctl status wol-host-agent.service"
