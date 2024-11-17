.PHONY: build test

build:
	go build -o bin/docker-machine-driver-utm ./cmd/docker-machine-driver-utm

test:
	go clean -testcache
	go test -v ./...
