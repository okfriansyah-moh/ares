# ADR-0001 — Use Go for the ARES CLI

**Status:** Accepted
**Date:** 2026-06-04

## Context

ARES needs a CLI tool (`ars`) that teams can install and run on developer machines and in CI pipelines without any runtime dependency. The tool must be cross-platform (macOS, Linux, Windows) and usable via `go install`, Homebrew, and `npx`-style wrappers. Distribution simplicity is a first-class concern because adoption depends on a frictionless install experience.

## Decision

Implement the CLI in Go 1.26.

## Alternatives

| Option | Why not chosen |
|---|---|
| Node.js | Requires Node runtime; `npx`-based distribution is possible but slower and heavier than a native binary |
| Python | Requires Python runtime and pip install; not suitable as a standalone binary without PyInstaller, which adds complexity |
| Rust | Produces excellent static binaries but the toolchain and build times are significantly higher; the team is not fluent in Rust |

## Tradeoffs

**Gained:**
- Single static binary (`CGO_ENABLED=0`) — no runtime, no dynamic linker
- `go install github.com/ars-standard/ars/cmd/ars@latest` works out of the box
- Cross-compilation is trivial with `GOOS`/`GOARCH`
- Strong stdlib for file I/O, path manipulation, and YAML
- Fast compile times

**Given up:**
- Marginally more verbose error handling than Python or Rust
- No npm ecosystem for distribution wrappers (must build separately)

## Consequences

- All CLI code lives under `cmd/ars/` and `internal/`
- CGO must remain disabled; no cgo dependencies permitted
- The build command is `CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ars ./cmd/ars`
- The Dockerfile uses `golang:1.26-alpine` as the builder stage and discards it in the final image
