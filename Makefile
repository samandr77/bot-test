.PHONY: build run lint test check

# Сборка приложения
build:
	go build -o bin/bot ./cmd/bot

# Запуск приложения
run:
	go run cmd/bot/main.go

# Запуск линтера
lint:
	golangci-lint run ./...

# Запуск тестов
test:
	go test -v ./...

# Полная проверка качества (линтер + тесты)
check: lint test build
	@echo "All checks passed!"
