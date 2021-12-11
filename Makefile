SHELL   := /bin/bash
APP_PORT := ${or ${APP_PORT}, "8080"}
FRONTEND_DIR := ${or ${FRONTEND_DIR}, ../nascalay-frontend}

build:
	@go build -v ./...

run:
	@APP_PORT=${APP_PORT} go run github.com/cosmtrek/air@latest

.PHONY: lint
lint:
	@go run github.com/golangci/golangci-lint/cmd/golangci-lint@latest run --fix

go-gen:
	@go generate ./...

.PHONY: dev-with-client
dev-with-client:
	@cd ${FRONTEND_DIR} && yarn build
	@cp ${FRONTEND_DIR}/dist . -r
	@APP_PORT=${APP_PORT} go run main.go -b /api
