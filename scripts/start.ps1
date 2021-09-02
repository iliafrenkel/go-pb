& .\.env.ps1

go run ./cmd/main.go --web-log-mode=debug

# docker run --rm -v "${PWD}:/src" returntocorp/semgrep --lang=go --config=p/ci
# docker run --rm -v "${PWD}:/src" returntocorp/semgrep --lang=go --config=p/security-audit