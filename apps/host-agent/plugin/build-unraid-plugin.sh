#!/usr/bin/env bash
set -euo pipefail

SCRIPT_DIR="$(cd -- "$(dirname -- "${BASH_SOURCE[0]}")" && pwd)"
AGENT_DIR="$(cd -- "${SCRIPT_DIR}/.." && pwd)"
VERSION="${VERSION:-$(date +%Y.%m.%d)}"
BASE_URL="${BASE_URL:-https://github.com/NagaseKouichi/wol-web/releases/download/${VERSION}}"
PLUGIN_URL="${PLUGIN_URL:-${BASE_URL}/wol-host-agent.plg}"
GOCACHE="${GOCACHE:-/tmp/go-build}"
DIST_DIR="${SCRIPT_DIR}/dist"
BUILD_DIR="${SCRIPT_DIR}/build"
PKG_ROOT="${BUILD_DIR}/pkg"
PACKAGE_FILE="wol-host-agent-${VERSION}-x86_64-1.txz"

rm -rf "$BUILD_DIR"
mkdir -p "$DIST_DIR" "$PKG_ROOT"

cp -a "${SCRIPT_DIR}/package/." "$PKG_ROOT/"
mkdir -p "$PKG_ROOT/usr/local/sbin" "$PKG_ROOT/install"

(
	cd "$AGENT_DIR"
	GOCACHE="$GOCACHE" GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o "$PKG_ROOT/usr/local/sbin/wol-host-agent" .
)

cat > "$PKG_ROOT/install/slack-desc" <<'EOF'
wol-host-agent: wol-host-agent
wol-host-agent:
wol-host-agent: Authenticated power-management API for WOL Web.
wol-host-agent:
wol-host-agent: Provides shutdown and S3 sleep endpoints for a target Unraid host.
wol-host-agent:
wol-host-agent:
wol-host-agent:
wol-host-agent:
wol-host-agent:
wol-host-agent:
EOF

chmod 0755 "$PKG_ROOT/usr/local/sbin/wol-host-agent"
chmod 0755 "$PKG_ROOT/etc/rc.d/rc.wol-host-agent"
chmod 0755 "$PKG_ROOT/usr/local/emhttp/plugins/wol-host-agent/scripts/apply.sh"
chmod 0644 "$PKG_ROOT/usr/local/emhttp/plugins/wol-host-agent/scripts/control.php"
chmod 0644 "$PKG_ROOT/usr/local/emhttp/plugins/wol-host-agent/WOLHostAgent.page"
chmod 0644 "$PKG_ROOT/usr/local/emhttp/plugins/wol-host-agent/default.cfg"

tar -C "$PKG_ROOT" --owner=0 --group=0 -cJf "${DIST_DIR}/${PACKAGE_FILE}" .
PACKAGE_SHA256="$(sha256sum "${DIST_DIR}/${PACKAGE_FILE}" | awk '{print $1}')"
PACKAGE_URL="${BASE_URL}/${PACKAGE_FILE}"

sed \
	-e "s#__VERSION__#${VERSION}#g" \
	-e "s#__PLUGIN_URL__#${PLUGIN_URL}#g" \
	-e "s#__PACKAGE_URL__#${PACKAGE_URL}#g" \
	-e "s#__PACKAGE_SHA256__#${PACKAGE_SHA256}#g" \
	"${SCRIPT_DIR}/wol-host-agent.plg.in" > "${DIST_DIR}/wol-host-agent.plg"

echo "Built:"
echo "  ${DIST_DIR}/${PACKAGE_FILE}"
echo "  ${DIST_DIR}/wol-host-agent.plg"
echo
echo "Package SHA256: ${PACKAGE_SHA256}"
