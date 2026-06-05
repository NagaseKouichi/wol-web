# WOL Host Agent

Small authenticated API that runs on a target host and lets `wol-web` request
shutdown or sleep while the host is online.

## Run

Generate a long random token and start the agent:

```bash
export HOST_AGENT_TOKEN="replace-with-a-long-random-token"
export HOST_AGENT_PORT=8765
go run .
```

The API listens on:

```text
http://<target-host-ip>:8765
```

Configure the matching `Agent URL` and `Agent Token` on the host record in
`wol-web`.

## Install As A Linux Service

On the target Linux host:

```bash
cd apps/host-agent
sudo ./deploy/systemd/install.sh
```

Edit the token:

```bash
sudo nano /etc/wol-host-agent.env
```

Start the service:

```bash
sudo systemctl start wol-host-agent.service
sudo systemctl status wol-host-agent.service
```

It is enabled for boot automatically by the installer. To stop or disable it:

```bash
sudo systemctl stop wol-host-agent.service
sudo systemctl disable wol-host-agent.service
```

## Install On Unraid

Unraid does not use `systemd`, so use `/boot/config/go` or the User Scripts
plugin for boot startup.

Build the Linux binary on another machine:

```bash
cd apps/host-agent
GOOS=linux GOARCH=amd64 go build -o wol-host-agent .
```

Copy these files to the Unraid flash drive:

```text
/boot/config/plugins/wol-host-agent/wol-host-agent
/boot/config/plugins/wol-host-agent/wol-host-agent.env
/boot/config/plugins/wol-host-agent/start-wol-host-agent.sh
```

You can use the templates from `deploy/unraid`:

```bash
mkdir -p /boot/config/plugins/wol-host-agent
cp wol-host-agent /boot/config/plugins/wol-host-agent/
cp deploy/unraid/wol-host-agent.env.example /boot/config/plugins/wol-host-agent/wol-host-agent.env
cp deploy/unraid/start-wol-host-agent.sh /boot/config/plugins/wol-host-agent/
chmod +x /boot/config/plugins/wol-host-agent/wol-host-agent
chmod +x /boot/config/plugins/wol-host-agent/start-wol-host-agent.sh
```

Edit the token:

```bash
nano /boot/config/plugins/wol-host-agent/wol-host-agent.env
```

Start it manually:

```bash
/boot/config/plugins/wol-host-agent/start-wol-host-agent.sh
```

To start at boot, add this line to `/boot/config/go`:

```bash
/boot/config/plugins/wol-host-agent/start-wol-host-agent.sh &
```

Alternatively, install the Unraid User Scripts plugin and run the same startup
script at array start.

Unraid shutdown/sleep defaults:

```text
shutdown -> powerdown
sleep    -> echo -n mem > /sys/power/state
```

Unraid's official docs recommend the Dynamix S3 Sleep plugin for user-friendly
sleep management, but the direct `/sys/power/state` command is also documented.

## API

```http
GET /health
```

```http
POST /api/power
Authorization: Bearer <token>
Content-Type: application/json

{"action":"shutdown"}
```

```http
POST /api/power
Authorization: Bearer <token>
Content-Type: application/json

{"action":"sleep"}
```

## Default Commands

Linux:

```text
shutdown -> systemctl poweroff
sleep    -> systemctl suspend
```

Windows:

```text
shutdown -> shutdown /s /t 0
sleep    -> rundll32.exe powrprof.dll,SetSuspendState 0,1,0
```

On Windows, disable hibernation if `SetSuspendState` enters hibernate instead
of sleep:

```powershell
powercfg /hibernate off
```

Custom commands can be configured:

```bash
export HOST_AGENT_SHUTDOWN_CMD="systemctl poweroff"
export HOST_AGENT_SLEEP_CMD="systemctl suspend"
```

The agent needs enough OS privileges to run the configured shutdown/sleep
commands.
