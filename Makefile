help:
	@echo "Usage:"
	@echo "  make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  lint              Run golangci-lint with auto-fix"
	@echo "  test              Run unit tests with race detector in short mode"

lint:
	golangci-lint run --fix

test:
	go test -race -short -v ./...

.PHONY: lint test
