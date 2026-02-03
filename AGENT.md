# AGENT.md

## Overview
Goma Admin is a Go backend (Okapi) with a stubbed Vue 3 UI. The entry point is `main.go`, which wires config, routes, and server startup.

## Repo layout
- `main.go`: App bootstrap, Okapi CLI flags, server start.
- `config/`: Env parsing, DB/Redis config, OpenAPI toggle.
- `routes/`: Route group definitions.
- `services/`: HTTP handlers (mostly placeholder responses).
- `ui/`: UI placeholder (no build config yet).

## Run (backend)
1. `cp .env.example .env` and set required values.
2. `go run ./main.go` (optionally `--port 8080`).
3. OpenAPI docs are enabled when `ENABLE_DOCS=true`.

## Env and config
- `DB_HOST`, `DB_USER`, `DB_PASSWORD`, `DB_NAME`, `DB_PORT`, `DB_SSL_MODE` for Postgres.
- `REDIS_URL` must be non-empty or config validation fails.
- `PORT` overrides the CLI `--port` default.
- `CORS_ALLOWED_ORIGINS` is comma-separated.
- `JWT_SECRET` and `LOG_LEVEL` are optional.

## API routes (current)
- `GET /` and `GET /version`
- `GET/POST/GET/PUT/DELETE /api/v1/routes/:id` (CRUD placeholders)
- `GET/POST/GET/PUT/DELETE /api/v1/middlewares/:id` (CRUD placeholders)
- `GET /api/v1/provider/:name`
- `GET /api/v1/provider/:name/routes`
- `GET /api/v1/provider/:name/middlewares`
- `POST /api/v1/provider/:name/webhook`

## Lint and format
- `golangci-lint` is configured in `.golangci.yml` and `gofmt` is enabled.

## Notes
- `go.mod` specifies Go `1.25.5` while `README.md` claims `>=1.21`.
- `API.md` is empty and `README.md` points to a missing `docs/API.md`.
- `ui/` is a stub without `package.json`.
