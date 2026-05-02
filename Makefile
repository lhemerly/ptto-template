SHELL := /bin/bash

.PHONY: dev run fmt tidy clean

dev:
	@# 1. Dependency Checks
	@command -v templ >/dev/null 2>&1 || { echo "templ missing: go install github.com/a-h/templ/cmd/templ@latest"; exit 1; }
	@command -v tailwindcss >/dev/null 2>&1 || { echo "tailwindcss missing: install the standalone binary"; exit 1; }
	@command -v air >/dev/null 2>&1 || { echo "air missing: go install github.com/air-verse/air@latest"; exit 1; }
	
	@# 2. Prep directories
	@mkdir -p ./assets ./tmp
	@echo "🚀 Starting ptto dev server..."
	
	@# 3. Execution (Tailwind outputs to /assets where the Go router expects it)
	@trap 'kill 0' EXIT INT TERM; \
	templ generate --watch & \
	tailwindcss -i ./assets/input.css -o ./assets/app.css --watch & \
	air --build.cmd "go build -o ./tmp/main ./cmd/server" --build.bin "./tmp/main" --build.delay "100" --build.exclude_dir "assets,web"

run:
	go run ./cmd/server/main.go

fmt:
	go fmt ./...
	templ fmt .

tidy:
	go mod tidy

clean:
	rm -rf ./tmp ./assets/app.css