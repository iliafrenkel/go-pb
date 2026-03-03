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
# Developer Guide

This guide is designed to take you on a journey from running a simple, ephemeral instance of the
application to configuring a full, production-like environment with a PostgreSQL database. We have
structured this document by complexity, so you can dive as deep as your task requires.

---

## 🐣 Level 1: The Quick Start
**Goal:** Run the application instantly. No persistence, no extra dependencies.

If you just want to see the application running, this is the place to start. By default, `go-pb`
uses **In-Memory** storage. This means it's incredibly fast to start, but all your pastes will
vanish if you restart the server.

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

That's it! Open your browser to `http://localhost:8080`. The `--debug` is not technically necessary,
but it will allow you to use the dev authenticator to login with different fake users.

---

## 💾 Level 2: Persistence (Disk Storage)
**Goal:** Save your work. Data survives restarts.

In-memory is great for testing UI changes, but if you're working on logic that requires data
persistence, you'll want to switch to **Disk Storage**. This saves pastes as files in a directory of
your choice.

This requires setting **Environment Variables**.

### Setting up the Environment

You can configure `go-pb` using command-line flags or environment variables. We recommend
environment variables for local development.

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
> *   **Windows:** You'll need to set them manually or use a script, as PowerShell doesn't source
>     `.env` files natively in the same way.

---

## 🐘 Level 3: The Full Setup (PostgreSQL)
**Goal:** A production-like environment. Requires Docker or Podman.

For backend development, performance testing, or working on database migrations, you need the real
deal: a PostgreSQL database.

### 1. Start the Database
We have prepared helper scripts to spin up a database (and Adminer, a DB management UI) instantly.
You don't need to install Postgres locally; you just need a container engine.

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

## Appendix: Configuration Reference

 **Application options:**
 Each option has a reasonable default value which can be overridden by a corresponding environment
 variable. The environment varaible can in turn be overridden by providing command line parameter.

 ```text
      --debug                   [$GOPB_DEBUG]
      --log-file=               Full path to the log file, default is stdout. If
                                run in container, it is best to log to stdout and
                                let the containerastion system handle the logs.
                                Environment variable: $GOPB_LOG_FILE

timeout:
      --timeout-shutdown=       Server graceful shutdown timeout (default: 10s).
                                The maximum time server has to close all
                                connections and shutdown. There is rarely a need
                                to change this.
                                Environment variable: $GOPB_TIMEOUT_SHUTDOWN
      --timeout-http-read=      Duration for reading the entire request (default:
                                15s), This is to protect from very slow connections
                                to keep the connection open.
                                Environment variable: $GOPB_TIMEOUT_HTTP_READ
      --timeout-http-write=     Duration before timing out writes of the response
                                (default: 15s). Same as above, protect from slow
                                connections.
                                Environment variable: $GOPB_TIMEOUT_HTTP_WRITE
      --timeout-http-idle=      Amount of time to wait for the next request on a
                                persistent (keep-alive) connection. If set to 0,
                                read timeout will be used instead (default: 60s).
                                Environment variable: $GOPB_TIMEOUT_HTTP_IDLE

web:
      --web-host=               Hostname part of the Web server address (default:
                                localhost).
                                Environment variable: $GOPB_WEB_HOST
      --web-port=               Port part of the Web server address (default: 8080)
                                Environment variable: $GOPB_WEB_PORT
      --web-log-file=           Full path to the log file, default is stdout.
                                Environment variable: $GOPB_WEB_LOG_FILE
      --web-log-mode=           Log mode, can be 'debug' or 'production' (default:
                                production)
                                Environment variable: $GOPB_WEB_LOG_MODE
      --web-brand-name=         Brand name shown in the header of every page
                                (default: Go PB)
                                Environment variable: $GOPB_WEB_BRAND_NAME
      --web-brand-tagline=      Brand tagline shown below the brand name (default:
                                "A nice and simple pastebin alternative that you can
                                host yourself.")
                                Environment variable: $GOPB_WEB_BRAND_TAGLINE
      --web-assets=             Path to the assets folder (default: ./assets). This
                                is where all the CSS, JavaScripts, and icons are.
                                Environment variable: $GOPB_WEB_ASSETS
      --web-templates=          Path to the templates folder (default: ./templates).
                                This is where the HTML templates are.
                                Environment variable: $GOPB_WEB_TEMPLATES
      --web-bootstrap-theme=    Name of the bootstrap CSS theme to use. One of the
                                following: flatly, litera, materia, sandstone, yeti
                                or zephyr. (default: original)
                                Environment variable: $GOPB_WEB_BOOTSTRAP_THEME]
      --web-logo=               Name of the logo image file within the assets folder
                                (default: bighead.svg)
                                Environment variable: $GOPB_WEB_LOGO
      --web-max-body-size=      Maximum size for request's body in bytes (default:
                                10240). Be carefull not to set it too high. Bad
                                people can start using your server as a storage.
                                Environment variable: $GOPB_WEB_MAX_BODY_SIZE

db:
      --db-type=                Database type to use for storage. Can be one of:
                                memory - everything is stored in memory, everyting
                                will be gone after server shutdown. Main use for
                                development and debugging.
                                disk - use files on disk to store the data. Data is
                                persisted and survives restarts. Good for low use
                                setups. ⚠ Data on disk is not encrypted! ⚠
                                postgres - use Postgres database for persistence.
                                Best for high use cases.
                                (default: memory)
                                Environment variable: $GOPB_DB_TYPE
      --db-connection=          Database connection string for Postgres type.
                                Ignored for memory or disk storage types.
                                Environment variable: $GOPB_DB_CONNECTION

auth:
      --auth-secret=            Secret used for JWT token generation/verification.
                                Environment variable: $GOPB_AUTH_SECRET
      --auth-token-duration=    JWT token expiration. (default: 5m)
                                Environment variable: $GOPB_AUTH_TOKEN_DURATION
      --auth-cookie-duration=   Cookie expiration. (default: 24h)
                                Environment variable: $GOPB_AUTH_COOKIE_DURATION
      --auth-issuer=            App name used to oauth requests. (default: go-pb)
                                Environment variable: $GOPB_AUTH_ISSUER
      --auth-url=               Callback url for oauth requests. (default:
                                http://localhost:8080)
                                Environment variable: $GOPB_AUTH_URL
      --auth-github-cid=        Github client id used for oauth.
                                Environment variable: $GOPB_AUTH_GITHUB_CID
      --auth-github-csec=       Github client secret used for oauth.
                                Environment variable: $GOPB_AUTH_GITHUB_CSEC
      --auth-google-cid=        Google client id used for oauth.
                                Environment variable: $GOPB_AUTH_GOOGLE_CID
      --auth-google-csec=       Google client secret used for oauth.
                                Environment variable: $GOPB_AUTH_GOOGLE_CSEC

disk:
      --disk-data-dir=          Directory where pastes are stored. The directory
                                must exist. (default: ./data)
                                Environment variable: $GOPB_DISK_DATA_DIR
      --disk-cache-size=        File system storage cache size.
                                Environment variable: $GOPB_DISK_CACHE_SIZE
      --disk-dir-mode=          File mode for new directories.
                                Environment variable: $GOPB_DISK_DIR_MODE
```
