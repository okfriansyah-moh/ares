# ADR-0003 — Use Distroless as the Container Base Image

**Status:** Accepted
**Date:** 2026-06-04

## Context

The ARES CLI is distributed as a container image for CI/CD use cases (e.g., running `ars compose` in a GitHub Actions job). Container images must have zero known CVEs and must not provide a shell or package manager that could be exploited if the container is compromised.

Three base image options were evaluated for the final stage of the multi-stage build.

## Decision

Use `gcr.io/distroless/static-debian12:nonroot` as the final stage base image.

## Alternatives

| Option | Why not chosen |
|---|---|
| `alpine:3` | Contains a shell, package manager (apk), and libc — larger attack surface; CVEs appear regularly in Alpine base packages |
| `scratch` | Truly minimal but lacks CA certificates (needed for future HTTPS support) and the `nonroot` user convention; harder to extend |
| `debian:slim` | Contains apt, shell, and many OS packages — defeats the purpose of hardening |

## Tradeoffs

**Gained:**
- No shell (`/bin/sh`, `/bin/bash`) — no shell injection surface
- No package manager — no `apt`/`apk` CVEs
- No setuid binaries
- Runs as UID 65532 (`nonroot`) by convention — no root process
- Contains CA certificates — HTTPS works if needed in future versions
- Smaller than Alpine for static binaries (fewer files overall)

**Given up:**
- Cannot `docker exec` into the container interactively — intentional
- Debugging must happen at the build stage or via volume mounts, not inside the running container
- Adding OS-level packages requires switching base image entirely

## Consequences

- Final stage is always `FROM gcr.io/distroless/static-debian12:nonroot`
- The binary must be fully static: `CGO_ENABLED=0`
- `USER nonroot:nonroot` is set in the Dockerfile
- CI must pin the distroless image by digest for reproducible builds
- `govulncheck ./...` covers Go dependencies; no OS package scanner needed because there are no OS packages
- The container should be run with `--read-only` to enforce a read-only root filesystem
