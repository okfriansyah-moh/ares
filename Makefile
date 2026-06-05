.PHONY: build test lint vuln docker-build docker-scan release release-dry release-checksums release-tag clean

VERSION ?= dev

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o bin/ars ./cmd/ars

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	staticcheck ./...

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

docker-build:
	docker build -t ars:dev .

docker-scan:
	docker run --rm aquasec/trivy:latest image ars:dev --exit-code 1 --severity HIGH,CRITICAL

release:
	mkdir -p bin
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o bin/ars-linux-amd64 ./cmd/ars
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o bin/ars-linux-arm64 ./cmd/ars
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o bin/ars-darwin-amd64 ./cmd/ars
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o bin/ars-darwin-arm64 ./cmd/ars
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o bin/ars-windows-amd64.exe ./cmd/ars

release-dry:
	mkdir -p dist
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o dist/ars-linux-amd64 ./cmd/ars
	CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o dist/ars-linux-arm64 ./cmd/ars
	CGO_ENABLED=0 GOOS=darwin GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o dist/ars-darwin-amd64 ./cmd/ars
	CGO_ENABLED=0 GOOS=darwin GOARCH=arm64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o dist/ars-darwin-arm64 ./cmd/ars
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 go build -trimpath -ldflags="-s -w -X github.com/ars-standard/ars/internal/version.Version=$(VERSION)" -o dist/ars-windows-amd64.exe ./cmd/ars
	@ls -lh dist/

release-checksums:
	@[ -d dist ] || (echo "dist/ does not exist; run make release-dry first"; exit 1)
	@cd dist && for f in ars-*; do \
		case "$$f" in *.sha256) continue ;; esac; \
		if command -v sha256sum >/dev/null 2>&1; then \
			sha256sum "$$f" > "$$f.sha256"; \
		else \
			shasum -a 256 "$$f" > "$$f.sha256"; \
		fi; \
	done
	@cat dist/*.sha256

release-tag:
	@[ "$(VERSION)" != "dev" ] || (echo "Usage: make release-tag VERSION=v1.2.3"; exit 1)
	git tag -a $(VERSION) -m "Release $(VERSION)"
	git push origin $(VERSION)

clean:
	rm -rf bin/ dist/
