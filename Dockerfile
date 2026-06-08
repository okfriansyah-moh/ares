# Stage 1: Build
FROM golang:1.26-alpine AS builder

ARG VERSION=dev

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 \
    go build \
      -trimpath \
      -ldflags="-s -w -X github.com/okfriansyah-moh/ares/internal/version.Version=${VERSION}" \
      -o /ars \
      ./cmd/ars

# Stage 2: Final - no shell, no package manager, nonroot user
FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=builder /ars /ars

USER nonroot:nonroot

ENTRYPOINT ["/ars"]
