run: build
	go run .
.PHONY: run

test:
	go test ./...
.PHONY: test

build:
	go build .
.PHONY: build
