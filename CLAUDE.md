# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
make build          # Build binary to build/go-pb
make test           # Run unit tests (short mode)
make test-coverage  # Run tests with coverage report
make lint           # Run gosec, go vet, staticcheck
make dep            # Tidy and download dependencies
make clean          # Remove previous build
```

Run a single test:
```bash
go test -short -run TestName ./src/service/
```

Run the app locally (defaults to in-memory store):
```bash
go run ./cmd/... --db.type=memory
```

## Architecture

The app follows a strict three-layer architecture: `web` → `service` → `store`.

- **`cmd/main.go`** — Entry point. Parses config (CLI flags or `GOPB_*` env vars), wires together the web server with a chosen store backend, and handles graceful shutdown.

- **`src/store/`** — Storage abstraction. `store.go` defines the `Store` interface. Three implementations: `memory.go` (thread-safe maps), `postgres.go` (GORM + PostgreSQL), `disk.go` (diskv file store). Core types `Paste` and `User` are defined here.

- **`src/service/`** — Business logic. Sits between web and store. Handles paste expiration parsing, bcrypt password hashing, burner paste deletion on read, view count increment, and privacy/user validation. Custom error types (`ErrPasteNotFound`, `ErrUserNotFound`, etc.) are defined here.

- **`src/web/`** — HTTP server. Uses standard library `http.ServeMux`. Authentication is handled via `go-pkgz/auth/v2` (JWT + OAuth: GitHub, Google, Twitter). `page/` renders HTML templates. `routes.go` contains all handlers.

- **`templates/`** and **`assets/`** — HTML templates and static assets served by the web layer.

## Development Conventions

- Prefer the standard library over external dependencies.
- Document all exported symbols with concise comments; no emojis in comments.
- New features go on separate branches and merge to `main` when complete.
- Lint with `gosec`, `go vet`, and `staticcheck` before submitting changes.
