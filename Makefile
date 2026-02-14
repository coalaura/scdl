check: tidy fmt lint test build

.PHONY: build
build:
	@go build ./...

.PHONY: test
test:
	@go test ./... > /dev/null 2>&1 || (echo "unit tests failed" && go test ./... && exit 1)

.PHONY: tidy
tidy:
	@go mod tidy

.PHONY: fmt
fmt:
	@go fmt ./...

.PHONY: lint
lint:
	@golangci-lint run
