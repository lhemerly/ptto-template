SHELL := /bin/bash

.PHONY: dev run fmt tidy

dev:
	@bash -c 'set -euo pipefail; \
		if ! command -v templ >/dev/null 2>&1; then echo "templ CLI is required: go install github.com/a-h/templ/cmd/templ@latest"; exit 1; fi; \
		if ! command -v tailwindcss >/dev/null 2>&1; then echo "tailwindcss CLI is required and must be on PATH"; exit 1; fi; \
		if ! command -v air >/dev/null 2>&1; then echo "air is required: go install github.com/air-verse/air@latest"; exit 1; fi; \
		mkdir -p ./web/static; \
		trap "kill 0" EXIT INT TERM; \
		templ generate --watch & \
		tailwindcss -i ./assets/input.css -o ./web/static/app.css --watch & \
		air --build.cmd "go build -o /tmp/ptto-dev ./cmd/server" --build.bin "/tmp/ptto-dev" --build.delay "100";'

run:
	go run ./cmd/server

fmt:
	go fmt ./...

tidy:
	go mod tidy
