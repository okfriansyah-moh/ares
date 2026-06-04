.PHONY: build test lint vuln docker-build clean

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o bin/ars ./cmd/ars

test:
	go test -race -count=1 ./...

lint:
	go vet ./...
	staticcheck ./...

vuln:
	govulncheck ./...

docker-build:
	docker build -t ars:dev .

clean:
	rm -rf bin/
