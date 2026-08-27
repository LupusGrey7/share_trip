# ============================================================
# Check .env file
# ============================================================
ifneq (,$(wildcard ./.env))
    include .env
    export
endif

# ============================================================
# Check Operation System PC — единственная логика что остаётся в Makefile
# ============================================================
ifeq ($(OS),Windows_NT)
    DETECTED_OS := Windows
    YAML_CHECK_SCRIPT := scripts\yaml-check.bat
    RM_RF = if exist "$(BUILD_DIR)" rmdir /s /q "$(BUILD_DIR)"
else
    DETECTED_OS := $(shell uname -s)
    YAML_CHECK_SCRIPT := bash scripts/yaml-check.sh
    RM_RF = rm -rf
endif

# ============================================================
# Important Variables
# ============================================================
GO := go
GO_PKG := ./...
APP_NAME=sharetrip
BUILD_DIR=./build
MAIN_FILE=cmd/sharetrip/main.go
DB_DSN=${DB_DRIVER}://${DB_USER}:${DB_PASSWORD}@${DB_HOST}:${DB_PORT}/${DB_NAME}?sslmode=${DB_SSLMODE}
MIGRATIONS_DIR = ./migrations
DEPLOY_DIR := ./deploy
DC := $(DEPLOY_DIR)/docker-compose.yml

# Version and ldflags (for embedding the version into the binary)
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Task - Default target
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  deps        	  - Install tools or check their availability"
	@echo "  fmt         	  - Code formating"
	@echo "  lint        	  - Running the linter"
	@echo "  test        	  - Run all tests"
	@echo "  build           - Build a binary file"
	@echo "  run         	  - Running an application locally"
	@echo "  e2e         	  - End To End check an application locally"
	@echo "  up          	  - Raise app infrastructure docker image"
	@echo "  start        	  - Start app docker image"
	@echo "  stop        	  - Stop app infrastructure docker image"
	@echo "  restart         - Restart app infrastructure docker image"
	@echo "  clean-image     - Clean all app infrastructure docker image"
	@echo "  down        	  - Down app infrastructure docker image"
	@echo "  migrate-up  	  - Apply migrations"
	@echo "  migrate-down	  - Roll back the last migration"
	@echo "  migrate-status  - Check migration status"
	@echo "  check           - A full run, like in CI: formatting, linter, tests"
	@echo "  coverage    	  - Run tests and generate HTML coverage report"
	@echo "  cover       	  - Alias for coverage"
	@echo "  vulncheck        - Vulnerability detection tool"
	@echo "  all         	  - Run lint, tests and coverage"
	@echo "  yaml-check      - Run yaml check tool"
	@echo "  info            - Show information about the OS and yq"
	@echo "  help        	  - Show this help"

# Task - Prepare environment (installation of tools)
# If you don't want to install locally (for example, again), comment out the commands inside
# ============================================================
# Prepare Environment (installation of tools)
# ============================================================
.PHONY: deps
deps:
	#$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 # install go linter
	#$(GO) install github.com/pressly/goose/v3/cmd/goose@latest) # install Goose

# Task - Formats source code
.PHONY: fmt
fmt:
	$(GO) fmt $(GO_PKG)

# Task - Check code with golangci-lint (start with check OS),
.PHONY: lint
lint:
ifeq ($(OS),Windows_NT)
	golangci-lint run --enable=paralleltest ./...
else
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "❌ golangci-lint is not installed. Please install it:"; \
		echo "   https://golangci-lint.run"; \
		exit 1; \
	fi
	golangci-lint run --enable=paralleltest ./...
endif

# Task - Run tests (detailed output (test names and PASS/FAIL))
.PHONY: test
test:
	$(GO) test -v $(GO_PKG)

# Task - Clean builds (Deletes compiled files)
.PHONY: clean
clean:
	@echo "Cleaning..."
ifeq ($(OS),Windows_NT)
	$(RM_RF)
else
	$(RM_RF) $(BUILD_DIR)
endif

# Task - Compile Go code to a binary file
.PHONY: build
build:
	@echo "Building..."
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

# Task - Local application run
.PHONY: run
run:
	$(GO) run $(MAIN_FILE)

# Task - Local application check (including adding a check for the response body:)
.PHONY: e2e
e2e:
ifeq ($(OS),Windows_NT)
	powershell -NoProfile -Command "$$r = Invoke-WebRequest -Uri http://localhost:8080/ready -UseBasicParsing; if ($$r.Content -notmatch 'OK') { exit 1 }"
else
	curl -f http://localhost:8080/ready | grep -q "OK"
endif

# ============================================================
# Task - Raise infrastructure (PostgreSQL in Docker)
# It is assumed that you have a docker-compose.yml in the ./deploy/docker-compose.yml directory
# ============================================================
.PHONY: up down restart start stop logs clean-image
# Task - Start all services
up:
	docker-compose -f $(DC) --project-name $(APP_NAME) up -d
# Task - Start specific service
start:
	docker-compose -f $(DC) --project-name $(APP_NAME) start

# Task - Stop all services
stop:
	docker-compose -f $(DC) --project-name $(APP_NAME) stop

# Task - Show logs of all services
logs:
	docker-compose -f $(DC) --project-name $(APP_NAME) logs -f

# Task - Restart all services
restart:
	docker-compose -f $(DC) --project-name $(APP_NAME) restart

# Task - Down all services
down:
	docker-compose -f $(DC) --project-name $(APP_NAME) down

# Task - Clean all: containers, networks and volumes (carefully!) will be deleted ALL!
clean-image:
	docker-compose -f $(DC) --project-name $(APP_NAME) down -v
ifeq ($(OS),Windows_NT)
	$(RM_RF)
else
	$(RM_RF) $(BUILD_DIR)
endif

# Task - Apply all pending migrations
.PHONY: migrate-up
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) up

# Task - Roll back the last migration
.PHONY: migrate-down
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) down

#Task - Check migration status
.PHONY: migrate-status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) status

# Task - A full run, like in CI: formatting, linter, tests
.PHONY: check
check:
	$(MAKE) fmt
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) coverage

#Task - Vulnerability detection tool
.PHONY: vulncheck
vulncheck:
	$(GO)vulncheck $(GO_PKG)

# Task - Generate coverage report in HTML format
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out $(GO_PKG)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: file://$(shell pwd)/coverage.html"

# Task - Output coverage to the terminal (optional)
.PHONY: cover-report
cover-report:
	$(GO) test -cover $(GO_PKG)

# Task - Default - Run all checks from the list
.PHONY: all
all: 
	$(MAKE) lint
	$(MAKE) test
	$(MAKE) coverage

# ============================================================
# Task - One line call - the script does everything itself
# ============================================================
.PHONY: yaml-check
yaml-check:
	@echo "OS: $(DETECTED_OS)"
	$(YAML_CHECK_SCRIPT)

# ============================================================
# Task - Show information about the OS and yq
# ============================================================
.PHONY: info
info:
	@echo "OS: $(DETECTED_OS)"
ifeq ($(OS),Windows_NT)
	@where yq >nul 2>&1 && yq --version || echo yq: not installed
else
	@echo "yq: $$(yq --version 2>/dev/null || echo 'not installed')"
endif

# DETECTED_OS: Windows | Linux | Darwin — для yaml-check и info