# AGENTS.md — OpenRGB Batocera Server

This file provides context for AI agents and coding assistants working on this project.

## Project Overview

A single-binary web application that exposes a browser-based GUI for controlling RGB lighting via [OpenRGB](https://openrgb.org/) on [Batocera Linux](https://batocera.org/).

The entire application (Go REST API + Angular SPA) compiles into **one portable static binary** with no runtime dependencies, using Go's `//go:embed` to bundle the Angular build output.

## Architecture

```
openrgb-batocera-server/
├── main.go                        # Entry point: CLI flags, auto-connect, HTTP server, middleware
├── go.mod / go.sum                # Go module (only external dep: go-openrgb-sdk)
├── Makefile                       # All build and dev commands
├── web/                           # Angular build output — embedded into the Go binary
│   └── .gitkeep                   # Placeholder; actual files are gitignored (generated)
├── internal/
│   ├── api/
│   │   ├── handler.go             # All HTTP route handlers
│   │   └── models.go              # Go structs for API request/response (JSON tags)
│   ├── openrgb/
│   │   └── client.go              # OpenRGB connection Manager (mutex-safe)
│   └── state/
│       └── state.go               # JSON persistence layer (state.json)
└── frontend/                      # Angular 21 SPA (separate build, then embedded)
    └── src/app/
        ├── components/            # device-detail, device-list
        ├── pages/home/            # Main page
        ├── services/
        │   └── openrgb.service.ts # Angular service wrapping all API calls
        └── models/
            └── openrgb.models.ts  # TypeScript interfaces mirroring Go structs
```

### Key design decisions

- **Single binary**: Angular is built first (`make build-frontend`), output is copied to `web/`, then Go embeds `web/` via `//go:embed`. Never edit files in `web/` manually.
- **State persistence**: `state.json` stores the last connection and per-device colors/modes. It is auto-loaded on startup.
- **Connection priority on startup**: CLI flag `--openrgb-host` > `OPENRGB_HOST` env var > saved `state.json`.
- **No external router**: Uses Go 1.22+ pattern-based routing in `net/http` (`METHOD /path/{param}`).
- **CORS + logging**: Applied via middleware wrappers in `main.go` (`corsMiddleware`, `loggingMiddleware`).

## Tech Stack

| Layer | Technology |
|---|---|
| Backend language | Go (module: `github.com/raiper34/openrgb-batocera-server`) |
| Go version | 1.22+ (go.mod specifies 1.26.2) |
| External Go dep | `github.com/csutorasa/go-openrgb-sdk v1.0.1` |
| Frontend framework | Angular 21 + PrimeNG 21 (Aura dark theme) |
| Frontend language | TypeScript ~5.9.2, strict mode enabled |
| Frontend styling | SCSS |
| Frontend state | RxJS 7.8 + Angular `HttpClient` |
| Build tool (frontend) | `@angular/build` (esbuild-based) |

## Commands

### Development

```sh
# Terminal 1 — Go backend on :8080
make dev-backend

# Terminal 2 — Angular dev server on :4200 (proxies /api to :8080)
make dev-frontend
```

The Angular dev proxy is configured in `frontend/proxy.conf.json` — all `/api/*` requests are forwarded to `localhost:8080`.

### Build

```sh
# Full build (frontend + backend, local binary)
make build

# Cross-compile for Batocera x86-64 (static, stripped)
make build-batocera

# Cross-compile for Batocera ARM64 (e.g. Raspberry Pi)
make build-batocera-arm64

# Frontend only
make build-frontend

# Backend only (requires web/ to already be populated)
make build-backend
```

### Clean

```sh
make clean
```

### Frontend dependency install

```sh
cd frontend && npm install
```

## REST API

All endpoints are under `/api/`. The backend uses Go 1.22 method+pattern routing.

| Method | Endpoint | Description |
|---|---|---|
| `GET` | `/api/status` | Connection status, saved/env connection info |
| `POST` | `/api/connect` | Connect: `{ "host": string, "port": int }` |
| `POST` | `/api/disconnect` | Disconnect |
| `GET` | `/api/devices` | List all RGB devices |
| `GET` | `/api/devices/{id}` | Get a single device by index |
| `POST` | `/api/devices/{id}/colors` | Set device colors: `{ "colors": [{r,g,b}] }` |
| `POST` | `/api/devices/{id}/zones/{zone_id}/colors` | Set zone colors: `{ "colors": [{r,g,b}] }` |
| `POST` | `/api/devices/{id}/mode` | Set mode: `{ "mode_id": int }` |
| `POST` | `/api/all-colors` | Apply one color to all devices: `{ "r", "g", "b" }` |

Error responses always use `{ "error": "message" }`.

## Code Conventions

### Go

- All exported types and methods have doc-comments.
- Shared mutable state (`Manager`, `State`) is protected by a `sync.Mutex`.
- HTTP handlers follow the pattern: parse path params → decode body → call manager → persist state → respond.
- Use `respondJSON(w, status, v)` / `respondError(w, status, msg)` helpers — never write raw JSON manually.
- Colors are stored as 6-char uppercase hex strings (`"FF0000"`) in `state.json`.
- `trimNull(s string)` must be called on all strings coming from the OpenRGB SDK (they may contain `\x00` null bytes).
- Run `go fmt ./...` before committing.

### Angular / TypeScript

- TypeScript strict mode is enabled — no `any`, no implicit returns, no implicit overrides.
- All API shapes are defined as interfaces in `frontend/src/app/models/openrgb.models.ts` and must stay in sync with Go structs in `internal/api/models.go`.
- All HTTP calls go through `OpenrgbService` (`frontend/src/app/services/openrgb.service.ts`). Do not use `HttpClient` directly in components.
- Use Prettier for formatting: `cd frontend && npx prettier --write .`
- Angular schematics are configured with `"skipTests": true` — there are currently no tests.

## Environment Variables

| Variable | Description |
|---|---|
| `OPENRGB_HOST` | OpenRGB host to auto-connect on startup |
| `OPENRGB_PORT` | OpenRGB port (default `6742`) |

## CLI Flags

| Flag | Default | Description |
|---|---|---|
| `--port` | `8080` | HTTP server listen port |
| `--openrgb-host` | _(empty)_ | Auto-connect to this host on startup |
| `--openrgb-port` | `6742` | OpenRGB server port |
| `--state-file` | `state.json` | Path to state persistence file |

## What Does Not Exist (yet)

- **No tests** — no `*_test.go` files, no `*.spec.ts` files.
- **No CI/CD** — no GitHub Actions or other pipeline.
- **No linting config** — no `.golangci.yml`, no ESLint.
- **No OpenAPI/Swagger spec** — API is documented only in this file and README.md.
- **No Docker** — deployment is a single binary copied via `scp`.
- **No remote git** — repository is local only.
