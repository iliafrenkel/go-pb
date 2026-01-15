# GEMINI.md

This document provides a comprehensive overview of the `go-pb` project, designed to serve as a
detailed instructional context for future interactions with the Gemini CLI.

## Project Overview

`go-pb` is a self-hostable pastebin alternative written in Go. It allows users to share text
snippets via short URLs, with features like syntax highlighting, self-destructing "burner" pastes,
password protection, and user accounts for managing pastes.

### Main Technologies

*   **Backend:** Go
*   **HTTP Routing:** `gorilla/mux`
*   **Database ORM:** `gorm.io/gorm`
*   **Configuration:** `jessevdk/go-flags` for command-line flags and environment variables.
*   **Database Support:**
    *   In-memory
    *   PostgreSQL
    *   On-disk (using `peterbourgon/diskv/v3`)

### Architecture

The application is structured into three main packages:

*   `cmd`: Contains the main application entry point (`main.go`). It handles command-line argument
    parsing, configuration, and server initialization.
*   `src`:
    *   `web`: Manages HTTP routing, handlers, and serves the frontend assets.
    *   `store`: Provides a data storage abstraction with implementations for in-memory, PostgreSQL,
        and on-disk storage.
    *   `service`: Implements the core business logic of the application.
*   `templates`: Contains the HTML templates for the web interface.
*   `assets`: Holds static assets like CSS, JavaScript, and images.

## Building and Running

### Building

To build the application, use the `make build` command. This will create a binary in the `build/`
directory.

```bash make build ```

### Testing

To run the unit tests, use the `make test` command.

```bash make test ```

To get a test coverage report, use `make test-coverage`.

```bash make test-coverage ```

### Running

The application can be started by running the compiled binary. Configuration can be provided through
command-line flags or environment variables.

A simple way to run the application for development is to use the `go run` command:

```bash go run ./cmd/... ```

The application can be configured using a variety of flags. For a complete list, run:

```bash go run ./cmd/... --help ```

## Development Conventions

### General rules

- Keep external dependencies to a minimum. Prefer standar library over external dependency even if it
means slightly mode code.
- Document everything with comments. Keep the comments concise and do not use emojis.
- New features must use separate git branches and merged into main branch once complete.

### Coding Style

The project follows standard Go coding conventions. The `Makefile` includes a `lint` target that
uses `gosec`, `go vet`, and `staticcheck` to enforce code quality.

### Testing

The project has a suite of unit tests. All new features should be accompanied by corresponding
tests. Tests are located in `_test.go` files alongside the code they are testing.

### Contribution Guidelines

Contributions are welcome. Please refer to the `CONTRIBUTING.md` file for more information. All
contributors are expected to adhere to the `CODE_OF_CONDUCT.md`.
