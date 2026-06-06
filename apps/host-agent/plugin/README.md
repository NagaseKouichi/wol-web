# WOL Host Agent Unraid Plugin

This plugin installs `wol-host-agent` on Unraid, adds a Settings page, and runs
the agent as an rc.d service.

## Build

From this directory:

```bash
VERSION=2026.06.05 \
BASE_URL=https://github.com/NagaseKouichi/wol-web/releases/download/2026.06.05 \
./build-unraid-plugin.sh
```

Outputs:

```text
dist/wol-host-agent-<version>-x86_64-1.txz
dist/wol-host-agent.plg
```

Upload both files to the same GitHub release. The generated `.plg` references
the `.txz` at `BASE_URL`.

## Install On Unraid

Install from a hosted `.plg` URL:

```bash
installplg https://github.com/NagaseKouichi/wol-web/releases/download/2026.06.05/wol-host-agent.plg
```

After installation, open:

```text
Settings > WOL Host Agent
```

Configure:

```text
Enable Agent: Yes
Agent Token: a long random token
Port: 8765
```

Click `Apply`. The plugin writes:

```text
/boot/config/plugins/wol-host-agent/wol-host-agent.cfg
/boot/config/plugins/wol-host-agent/wol-host-agent.env
```

and starts:

```text
/etc/rc.d/rc.wol-host-agent
```

## Configure wol-web

In wol-web, edit the matching host:

```text
Host IP: <Unraid LAN IP>
Agent URL: http://<Unraid LAN IP>:8765
Agent Token: the same token configured in the plugin
```

## Service Commands

```bash
/etc/rc.d/rc.wol-host-agent start
/etc/rc.d/rc.wol-host-agent stop
/etc/rc.d/rc.wol-host-agent restart
/etc/rc.d/rc.wol-host-agent status
```

Logs:

```text
/var/log/wol-host-agent.log
```

## Power Commands

The plugin configures Unraid-specific commands:

```text
shutdown -> powerdown
reboot   -> reboot
sleep    -> echo -n mem > /sys/power/state
```
