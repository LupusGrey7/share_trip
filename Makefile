# Переменные
GO := go
GO_PKG := ./...
APP_NAME=sharetrip
BUILD_DIR=./build
MAIN_FILE=cmd/sharetrip/main.go
DB_DSN = "postgres://postgres:password@localhost:6543/share_trip?sslmode=disable"
MIGRATIONS_DIR = ./migrations
DEPLOY_DIR := ./deploy
DC := $(DEPLOY_DIR)/docker-compose.yml

# Версия и ldflags (для вшивания версии в бинарник)
VERSION=1.0.0
LDFLAGS=-ldflags "-X main.Version=${VERSION}"

# Цель по умолчанию
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  deps        	  - Install tools or check their availability"
	@echo "  fmt         	  - Code formattingа"
	@echo "  lint        	  - Running the linter"
	@echo "  test        	  - Run all tests"
	@echo "  build           - Build a binary file"
	@echo "  run         	  - Running an application locally"
	@echo "  e2e         	  - End To End check an application locally"
	@echo "  up          	  - Raise app infrastructure docker image"
	@echo "  strat        	  - Start app docker image"
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
	@echo "  all         	  - Run lint, tests and coverage"
	@echo "  help        	  - Show this help"

# Задача - Подготовка окружения (установка инструментов)
# Если не хотите устанавливать локально(к примеру повторно), закомментируйте команды внутри
.PHONY: deps
deps:
	#$(GO) install github.com/golangci/golangci-lint/cmd/golangci-lint@v1.64.5 # install go linter
	#$(GO) install github.com/pressly/goose/v3/cmd/goose@latest) # install Goose

# Задача - Форматирует исходный код
.PHONY: fmt
fmt:
	$(GO) fmt $(GO_PKG)

# Проверка кода с помощью golangci-lint
.PHONY: lint
lint:
	@if ! command -v golangci-lint >/dev/null 2>&1; then \
		echo "❌ golangci-lint is not installed. Please install it:"; \
		echo "   https://golangci-lint.run/usage/install/"; \
		exit 1; \
	fi
	golangci-lint run

# Задача - Запуск тестов
.PHONY: test
test:
	$(GO) test -v $(GO_PKG)

# Задача - Очистка билдов (Удаляет скомпилированные файлы)
.PHONY: clean
clean:
	@echo "Cleaning..."
	$(GO) rm -rf $(BUILD_DIR)

# Компилирует исходный код на Go в бинарный файл
.PHONY: build
build:
	@echo "Building..."
	$(GO) build -o $(BUILD_DIR)/$(APP_NAME) $(MAIN_FILE)

# Задача - Локальный запуск приложения
.PHONY: run
run:
	$(GO) run $(MAIN_FILE)

# Задача - Локальная проверка работоспособности приложения (включая добавить проверку тела ответа:)
.PHONY: e2e
e2e:
	curl -f http://localhost:8080/api/ready | grep -q "OK"

# Задача - Поднять инфраструктуру (PostgreSQL в Docker)
# Предполагается, что у вас есть docker-compose.yml в каталоге ./deploy/docker-compose.yml
.PHONY: up down restart start stop logs clean
up:
	docker-compose -f $(DC) --project-name $(APP_NAME) up -d

# Остановить инфраструктуру (контейнеры останутся, но остановятся)
down:
	docker-compose -f $(DC) --project-name $(APP_NAME) down

# Перезапустить все сервисы проекта
restart:
	docker-compose -f $(DC) --project-name $(APP_NAME) restart

# Запустить все сервисы проекта (если они были остановлены)
start:
	docker-compose -f $(DC) --project-name $(APP_NAME) start

# Остановить все сервисы проекта
stop:
	docker-compose -f $(DC) --project-name $(APP_NAME) stop

# Посмотреть логи всех сервисов
logs:
	docker-compose -f $(DC) --project-name $(APP_NAME) logs -f

# Очистить всё: контейнеры, сети и volumes (осторожно!)
clean-image:
	docker-compose -f $(DC) --project-name $(APP_NAME) down -v
	rm -rf $(BUILD_DIR)

# Apply all pending migrations
.PHONY: migrate-up
migrate-up:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) up

# Roll back the last migration
.PHONY: migrate-down
migrate-down:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) down

# Check migration status
.PHONY: migrate-status
migrate-status:
	goose -dir $(MIGRATIONS_DIR) postgres $(DB_DSN) status

# Полный прогон, как в CI: форматирование, линтер, тесты
.PHONY: check
check: fmt lint test coverage

# Генерация отчёта о покрытии в формате HTML
.PHONY: coverage
coverage:
	$(GO) test -coverprofile=coverage.out $(GO_PKG)
	$(GO) tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: file://$(shell pwd)/coverage.html"

# Вывод покрытия в терминал (опционально)
.PHONY: cover-report
cover-report:
	$(GO) test -cover $(GO_PKG)

# По умолчанию - Запуск всех проверок из перечня
.PHONY: all
all: lint test coverage