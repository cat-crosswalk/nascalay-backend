SHELL   := /bin/bash

build:
	@go build -v ./...

.PHONY: lint
lint:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

go-gen:
	@go generate ./...
