# OpenRGB Batocera Server

A lightweight web server that exposes a browser-based GUI for controlling RGB lighting via [OpenRGB](https://openrgb.org/) on [Batocera Linux](https://batocera.org/). The entire application — backend + frontend — ships as a **single static binary** with no runtime dependencies.

## Table of Contents

- [Features](#features)
- [Technologies](#technologies)
- [Prerequisites](#prerequisites)
- [Installation on Batocera](#installation-on-batocera)
- [CLI flags](#cli-flags)
- [Environment variables](#environment-variables)
- [Connection priority on startup](#connection-priority-on-startup)
- [Using the UI](#using-the-ui)
- [Building from source](#building-from-source)
- [REST API](#rest-api)

## Features

- Dark-themed web UI accessible from any device on the local network (desktop, phone, tablet)
- Connect to any OpenRGB server by host and port
- Per-device color and mode control
- Per-zone color control
- Apply a single color to all devices at once
- Persistent state — colors and modes are restored automatically after a restart
- Auto-connect on startup using saved connection or environment variables
- Responsive design — fully usable on mobile
- **PWA support** — installable as a standalone app on mobile and desktop (works offline for cached assets)

## Technologies

| Layer | Technology |
|---|---|
| Backend | [Go](https://go.dev/) — single static binary (`CGO_ENABLED=0`) |
| OpenRGB protocol | [go-openrgb-sdk](https://github.com/csutorasa/go-openrgb-sdk) v1.0.1 |
| Frontend | [Angular](https://angular.dev/) 21 + [PrimeNG](https://primeng.org/) 21 (Aura dark theme) |
| Frontend bundling | Embedded into the Go binary via `//go:embed` |
| State persistence | JSON file on disk |
| PWA | Angular Service Worker (`ngsw`) with asset caching and API freshness strategy |

## Prerequisites

- **OpenRGB** must be running with the network server enabled:
  ```
  openrgb --server --port 6742
  ```
- On Batocera, OpenRGB can be installed via the [Batocera Unofficial Addons](https://github.com/batocera-unofficial-addons/batocera-unofficial-addons) project. After installation the AppImage is located at:
  ```
  /userdata/system/add-ons/openrgb/openrgb.AppImage
  ```

## Installation on Batocera

1. Copy the binary to Batocera:
   ```sh
   scp openrgb-batocera-server-batocera root@<batocera-ip>:/userdata/system/add-ons/openrgb-batocera-server/openrgb-batocera-server
   ```

2. Make it executable:
   ```sh
   ssh root@<batocera-ip> chmod +x /userdata/system/add-ons/openrgb-batocera-server/openrgb-batocera-server
   ```

3. Add it to `/userdata/system/custom.sh` so it starts with Batocera:
   ```sh
   #!/bin/bash

   # Start OpenRGB server
   /userdata/system/add-ons/openrgb/openrgb.AppImage --server --port 6742 &

   # Start OpenRGB Batocera Server
   nohup /userdata/system/add-ons/openrgb-batocera-server/openrgb-batocera-server \
     --openrgb-host localhost \
     --openrgb-port 6742 \
     --state-file /userdata/system/add-ons/openrgb-batocera-server/state.json \
     > /userdata/system/add-ons/openrgb-batocera-server/server.log 2>&1 &
   ```

4. Open the UI in a browser:
   ```
   http://<batocera-ip>:8080
   ```

## CLI flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `8080` | HTTP server port |
| `--openrgb-host` | _(empty)_ | Auto-connect to this OpenRGB host on startup |
| `--openrgb-port` | `6742` | OpenRGB server port |
| `--state-file` | `state.json` | Path to the persistent state file |

## Environment variables

| Variable | Description |
|---|---|
| `OPENRGB_HOST` | OpenRGB host to connect to on startup (lower priority than `--openrgb-host`) |
| `OPENRGB_PORT` | OpenRGB port — applied only when `OPENRGB_HOST` is used (not when `--openrgb-host` is set) |

> **Note:** When the server is started with an environment variable, the UI shows a warning that after the next restart the env variable will take priority over any connection change made in the UI.

## Connection priority on startup

1. `--openrgb-host` CLI flag (highest priority — env vars are ignored entirely when this is set)
2. `OPENRGB_HOST` environment variable (port is taken from `OPENRGB_PORT` if set, otherwise from `--openrgb-port`)
3. Connection saved from the last session (`state.json`)

## Using the UI

### Header
The header shows the current connection status. Click the **connection chip** (`host:port`) to open the **Connection Settings** dialog where you can change the host/port or disconnect.

If the server is not connected, a connect form is shown directly in the header.

### Sidebar (left panel)
- **Color picker + Apply All** — pick a color and apply it to every LED on every device at once
- **Refresh** — reload the device list from OpenRGB
- **Device list** — click any device to open its detail view

### Device detail (right panel / full screen on mobile)
Tabs available for each device:

| Tab | Description |
|---|---|
| **Colors** | Pick a color and apply it to all LEDs on the device |
| **Zones** | Select a zone and apply a color to that zone only |
| **Modes** | Select and apply a lighting mode |
| **LED Map** | Visual map of all LEDs with their current colors |

On mobile, tap a device to slide to the detail view. Use the **← Devices** button to go back.

## Building from source

### Requirements
- Go 1.26+
- Node.js 20+ / npm 11+

### Install frontend dependencies
```sh
cd frontend && npm install
```

### Build for local machine
```sh
make build
```

### Build for Batocera (x86-64, static binary)
```sh
make build-batocera
```

### Build for Batocera (ARM64, e.g. Raspberry Pi)
```sh
make build-batocera-arm64
```

### Development mode
Run the Go backend and Angular dev server in separate terminals:
```sh
make dev-backend   # terminal 1 — serves on :8080
make dev-frontend  # terminal 2 — Angular dev server on :4200 with proxy to :8080
```

### Clean build artifacts
```sh
make clean
```

## REST API

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/status` | Connection status + saved/env connection info |
| `POST` | `/api/connect` | Connect to OpenRGB `{ host, port }` |
| `POST` | `/api/disconnect` | Disconnect |
| `GET` | `/api/devices` | List all devices |
| `GET` | `/api/devices/{id}` | Get single device |
| `POST` | `/api/devices/{id}/colors` | Set device colors `{ colors: [{r,g,b}] }` |
| `POST` | `/api/devices/{id}/zones/{zone_id}/colors` | Set zone colors `{ colors: [{r,g,b}] }` |
| `POST` | `/api/devices/{id}/mode` | Set mode `{ mode_id }` |
| `POST` | `/api/all-colors` | Apply one color to all devices `{ r, g, b }` |
