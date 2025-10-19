.PHONY: help generate test clean run

help: ## ヘルプを表示
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

generate: ## モックコードを生成
	@echo "Generating mocks..."
	@cd apps/api-server && go generate ./...
	@echo "Mocks generated successfully!"

test: ## テストを実行
	@echo "Running tests..."
	@cd apps/api-server && go test ./... -v

test-coverage: ## カバレッジ付きでテストを実行
	@echo "Running tests with coverage..."
	@cd apps/api-server && go test ./... -v -coverprofile=coverage.out
	@cd apps/api-server && go tool cover -html=coverage.out -o coverage.html
	@echo "Coverage report generated: apps/api-server/coverage.html"

run: ## ローカルでAPIサーバーを起動
	@echo "Starting API server..."
	@cd apps/api-server && go run cmd/main.go

clean: ## 生成されたファイルをクリーンアップ
	@echo "Cleaning up generated files..."
	@find apps/api-server -type d -name "mock" -exec rm -rf {} + 2>/dev/null || true
	@rm -f apps/api-server/coverage.out apps/api-server/coverage.html
	@echo "Clean complete!"

build: ## バイナリをビルド
	@echo "Building binary..."
	@cd apps/api-server && go build -o ../../bin/api-server cmd/main.go
	@echo "Binary built: bin/api-server"
