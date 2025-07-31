TEST_DIR ?= ./...

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

test-debug:
	go test -v -timeout 10s ${TEST_DIR} | tee test.log

gql:
	cd account && go generate ./...

gql-client:
	cd account/accountusecase/accountproxy && go generate

.PHONY: lint test
