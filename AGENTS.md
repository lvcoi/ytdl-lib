# AGENTS.md

Go library and CLI (`youtubedr`) for downloading YouTube videos.

Module: `github.com/lvcoi/ytdl-lib/v2` — requires Go 1.23+.

## Commands

```bash
# Install dependencies
go mod tidy && go mod verify
# or
make deps

# Build (uses goreleaser)
make build

# Format
make format

# Lint (installs golangci-lint if missing)
make lint

# Unit tests
make test-unit

# Integration tests (requires ffmpeg, writes to output/)
make test-integration

# Run a single test
go test -v -run TestDownload ./downloader/

# Clean build artifacts
make clean
```

## CI Checks (GitHub Actions)

On push and PR (`.github/workflows/go.yaml`):
- **lint** — `make lint` on Go 1.23.x and latest, ubuntu-24.04
- **test** — `make test-integration` on Go 1.23.x and latest, ubuntu-24.04

Both must pass before merge.

## PR Requirements

- Link related issues (`#123`)
- Run `make format`, `make lint`, and `make test-integration` before submitting
- See `.github/pull_request_template.md`

## Key Directories

| Path | Description |
|---|---|
| `/` (root `.go` files) | Core library: client, video, playlist, transcript, decipher, formats |
| `cmd/youtubedr/` | CLI application entry point and subcommands |
| `downloader/` | Download logic, file utilities, progress reporting |
| `.github/workflows/` | CI: `go.yaml` (push/PR), `schedule.yaml` (daily integration tests) |
