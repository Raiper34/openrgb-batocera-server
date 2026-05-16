---
name: deploy-batocera
description: Build and deploy the openrgb-batocera-server binary to a Batocera Linux machine over SSH using Python paramiko
compatibility: opencode
---

## What I do

Deploy the openrgb-batocera-server to a Batocera Linux machine:

1. Build a static x86_64 binary with `make build-batocera`
2. Accept the remote host key into `~/.ssh/known_hosts`
3. Stop the running process on the target machine (if any)
4. Upload the binary via SFTP using Python `paramiko`
5. Set executable permissions
6. Start the server in the background with `nohup`
7. Verify the process is running and print the server log

## When to use me

Use this skill when the user asks to deploy, release, ship, or push the project to Batocera.

## Parameters to collect

Before starting, identify:
- `BATOCERA_HOST` — IP address or hostname of the Batocera machine (ask user if not provided)
- `BATOCERA_USER` — SSH username (default: `root`)
- `BATOCERA_PASSWORD` — SSH password (default: `linux`)
- `REMOTE_DIR` — deployment directory (default: `/userdata/system/add-ons/openrgb-batocera-server`)
- `OPENRGB_PORT` — OpenRGB server port (default: `6742`)
- `SERVER_PORT` — web UI port (default: `8080`)

## Step-by-step workflow

### 1. Build

```sh
make build-batocera
```

Produces: `./openrgb-batocera-server-batocera` (static, stripped, ~7 MB)

### 2. Accept host key

```sh
ssh-keyscan -H <BATOCERA_HOST> >> ~/.ssh/known_hosts
```

Run this before any SSH/SFTP operations to avoid host key verification failures.

### 3. Deploy via Python paramiko

Use the following Python script (run with `python3 - <<'EOF' ... EOF`):

```python
import paramiko, time

HOST = "<BATOCERA_HOST>"
USER = "<BATOCERA_USER>"        # default: root
PASSWORD = "<BATOCERA_PASSWORD>" # default: linux
REMOTE_DIR = "<REMOTE_DIR>"     # default: /userdata/system/add-ons/openrgb-batocera-server
LOCAL_BIN = "./openrgb-batocera-server-batocera"
REMOTE_BIN = f"{REMOTE_DIR}/openrgb-batocera-server"
OPENRGB_PORT = 6742
SERVER_PORT = 8080

client = paramiko.SSHClient()
client.set_missing_host_key_policy(paramiko.AutoAddPolicy())
client.connect(HOST, username=USER, password=PASSWORD, timeout=10)

# 1. Create remote directory
stdin, stdout, stderr = client.exec_command(f"mkdir -p {REMOTE_DIR}")
stdout.channel.recv_exit_status()

# 2. Kill running process (ignore error if not running)
stdin, stdout, stderr = client.exec_command("pkill -f openrgb-batocera-server; sleep 1")
stdout.channel.recv_exit_status()

# 3. Upload binary
sftp = client.open_sftp()
sftp.put(LOCAL_BIN, REMOTE_BIN)
sftp.close()

# 4. Set executable
stdin, stdout, stderr = client.exec_command(f"chmod +x {REMOTE_BIN}")
stdout.channel.recv_exit_status()

# 5. Start server
cmd = (
    f"nohup {REMOTE_BIN} "
    f"--openrgb-host localhost "
    f"--openrgb-port {OPENRGB_PORT} "
    f"--port {SERVER_PORT} "
    f"--state-file {REMOTE_DIR}/state.json "
    f"> {REMOTE_DIR}/server.log 2>&1 &"
)
stdin, stdout, stderr = client.exec_command(cmd)
stdout.channel.recv_exit_status()

# 6. Verify
time.sleep(2)
stdin, stdout, stderr = client.exec_command(f"cat {REMOTE_DIR}/server.log")
print("Log:", stdout.read().decode().strip())
stdin, stdout, stderr = client.exec_command("pgrep -a openrgb-batocera")
print("Process:", stdout.read().decode().strip())

client.close()
```

### 4. Confirm success

After running the script, verify:
- Log shows `OpenRGB Batocera Server starting on http://0.0.0.0:<port>`
- `pgrep` output contains the running process
- Tell the user the UI is available at `http://<BATOCERA_HOST>:<SERVER_PORT>`

## Notes

- Batocera default SSH credentials: user `root`, password `linux`
- `paramiko` must be available (`pip install paramiko` if missing)
- "Text file busy" error from SFTP means the binary is still running — always `pkill` before uploading
- The server auto-connects to OpenRGB on startup and restores saved device state from `state.json`
- Architecture: Batocera x86_64 uses the `build-batocera` target; for ARM64 (Raspberry Pi) use `build-batocera-arm64`
