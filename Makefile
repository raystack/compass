NAME="github.com/raystack/compass"
COMMIT := $(shell git rev-parse --short HEAD)
TAG := "$(shell git rev-list --tags --max-count=1)"
VERSION := "$(shell git describe --tags ${TAG})-next"
BUILD_DIR=dist

.PHONY: all build clean test tidy vet proto setup format generate lint install

all: clean test build format lint

tidy: ## Tidy go.mod
	@go mod tidy -v

install: ## Install compass binary
	@go install

format: ## Format Go source files
	@go fmt ./...

lint: ## Run lint checks
	@golangci-lint run

clean: ## Clean build artifacts
	@rm -rf coverage.out ${BUILD_DIR}

test: ## Run tests with coverage
	@go test ./... -race -coverprofile=coverage.out

e2e: ## Run e2e tests
	@go test ./test/... --tags=e2e

coverage: test ## Generate coverage report
	@go tool cover -html=coverage.out

build: ## Build the compass binary
	@CGO_ENABLED=0 go build -ldflags "-X ${NAME}/cli.Version=${VERSION}"

buildr: setup ## Build release snapshot
	@goreleaser release --snapshot --skip=publish --clean

vet: ## Run go vet
	@go vet ./...

download: ## Download go modules
	@go mod download

generate: ## Run go generate
	@go generate ./...

config: ## Generate sample config file
	@cp internal/config/config.example.yaml config.yaml

proto: ## Generate protobuf files
	@buf generate

setup: ## Install required dependencies
	@go mod tidy
	@go install github.com/vektra/mockery/v2@latest
	@go install github.com/bufbuild/buf/cmd/buf@latest

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'
