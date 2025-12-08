.PHONY: build run clean test help

# 預設目標
help:
	@echo "Available targets:"
	@echo "  build   - Build the application"
	@echo "  run     - Build and run the application"
	@echo "  clean   - Remove build artifacts"
	@echo "  test    - Run tests"
	@echo "  deps    - Download dependencies"

# 建置應用程式
build:
	@echo "Building APGo..."
	@go build -o bin/apgo ./cmd/main.go
	@echo "Build complete: bin/apgo"

# 執行應用程式
run: build
	@echo "Starting APGo..."
	@./bin/apgo

# 清理建置檔案
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf bin/
	@echo "Clean complete"

# 執行測試
test:
	@echo "Running tests..."
	@go test -v ./...

# 下載相依套件
deps:
	@echo "Downloading dependencies..."
	@go mod download
	@go mod tidy
	@echo "Dependencies updated"
