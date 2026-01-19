# Developer Guide

This guide is designed to take you on a journey from running a simple, ephemeral instance of the application to configuring a full, production-like environment with a PostgreSQL database. We have structured this document by complexity, so you can dive as deep as your task requires.

---

## 🐣 Level 1: The Quick Start
**Goal:** Run the application instantly. No persistence, no extra dependencies.

If you just want to see the application running, this is the place to start. By default, `go-pb` uses **In-Memory** storage. This means it's incredibly fast to start, but all your pastes will vanish if you restart the server.

1.  **Get the code:**
    ```bash
    git clone https://github.com/iliafrenkel/go-pb.git
    cd go-pb
    ```

2.  **Get dependencies:**
    ```bash
    make dep
    ```

3.  **Run it:**
    ```bash
    go run ./cmd/... --debug
    ```

That's it! Open your browser to `http://localhost:8080`. The `--debug` is not technically necessary, but it will allow you to use the dev authenticator to login with different fake users.

---

## 💾 Level 2: Persistence (Disk Storage)
**Goal:** Save your work. Data survives restarts.

In-memory is great for testing UI changes, but if you're working on logic that requires data persistence, you'll want to switch to **Disk Storage**. This saves pastes as files in a directory of your choice.

This requires setting **Environment Variables**.

### Setting up the Environment

You can configure `go-pb` using command-line flags or environment variables. We recommend environment variables for local development.

**Linux / MacOS:**
```bash
# Set the storage type to disk
export GOPB_DB_TYPE=disk
# Tell it where to save the files
export GOPB_DISK_DATA_DIR=./my-pastes

# Run the app
go run ./cmd/...
```

**Windows (PowerShell):**
```powershell
$env:GOPB_DB_TYPE="disk"
$env:GOPB_DISK_DATA_DIR="./my-pastes"

go run ./cmd/...
```

> **Pro Tip:** You can create a `.env` file in the root directory to save these variables.
> *   **Linux/Mac:** Run `source .env` before starting the app.
> *   **Windows:** You'll need to set them manually or use a script, as PowerShell doesn't source `.env` files natively in the same way.

---

## 🐘 Level 3: The Full Setup (PostgreSQL)
**Goal:** A production-like environment. Requires Docker or Podman.

For backend development, performance testing, or working on database migrations, you need the real deal: a PostgreSQL database.

### 1. Start the Database
We have prepared helper scripts to spin up a database (and Adminer, a DB management UI) instantly. You don't need to install Postgres locally; you just need a container engine.

**If you use Docker:**
```bash
./scripts/start-postgres-docker.sh
```

**If you use Podman:**
```bash
./scripts/start-postgres-podman.sh
```

This starts:
*   **PostgreSQL** on port `5432` (User/Pass/DB: `iliaf`)
*   **Adminer** on `http://localhost:8888`

### 2. Connect the App
Now, tell `go-pb` to talk to that database.

**Linux / MacOS:**
```bash
export GOPB_DB_TYPE=postgres
export GOPB_DB_CONNECTION="host=localhost port=5432 user=iliaf password=iliaf dbname=iliaf sslmode=disable"

go run ./cmd/...
```

**Windows (PowerShell):**
```powershell
$env:GOPB_DB_TYPE="postgres"
$env:GOPB_DB_CONNECTION="host=localhost port=5432 user=iliaf password=iliaf dbname=iliaf sslmode=disable"

go run ./cmd/...
```

---

## 🛠️ Level 4: The Engineer's Workflow
**Goal:** Verify, Build, and Ship.

Once your features are coded, it's time to ensure quality and build the artifacts.

### Testing
Run the full suite of unit tests to ensure no regressions.

```bash
make test
```
To see how much code you've covered:
```bash
make test-coverage
```

### Building the Binary
Compile the application into a standalone executable.

**Linux / MacOS:**
```bash
make build
# Binary created at: build/go-pb
```

**Windows:**
If you don't have `make`, use the Go toolchain directly:
```powershell
go build -o build/go-pb.exe ./cmd/...
```

### Docker Image
To verify that your changes work inside a container (just like in production):

```bash
# Build the image
docker build -f Dockerfile.build -t go-pb:local .

# Run it
docker run -p 8080:8080 go-pb:local
```

### Creating a Release
If you want to simulate a full release (creating archives, checksums, and binaries for all platforms), use `goreleaser`.

```bash
goreleaser release --snapshot
```
Check the `dist/` folder for the results.

---

## 📚 Appendix: Configuration Reference

Here are the most common variables you might need to tweak.

| Category | Variable | Default | Description |
| :--- | :--- | :--- | :--- |
| **General** | `GOPB_DEBUG` | `false` | Enable verbose debug logs. |
| **Server** | `GOPB_WEB_PORT` | `8080` | The port the web server listens on. |
| **Server** | `GOPB_WEB_HOST` | `localhost` | The interface to bind to. |
| **Database** | `GOPB_DB_TYPE` | `memory` | Options: `memory`, `disk`, `postgres`. |
| **Auth** | `GOPB_AUTH_SECRET` | - | Secret for generating JWT tokens. |