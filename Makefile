NAME="github.com/raystack/compass"
COMMIT := $(shell git rev-parse --short HEAD)
TAG := "$(shell git rev-list --tags --max-count=1)"
VERSION := "$(shell git describe --tags ${TAG})-next"
BUILD_DIR=dist
PROTON_COMMIT := "409f146"

.PHONY: all build clean test tidy vet proto setup format generate lint install

all: clean test build format lint

tidy:
	@echo "Tidy up go.mod..."
	@go mod tidy -v

install:
	@echo "Installing compass to ${GOBIN}..."
	@go install

format:
	@echo "Running go fmt..."
	@go fmt ./...

lint: ## Lint checker
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

clean: ## Clean the build artifacts
	@echo "Cleaning up build directories..."
	@rm -rf coverage.out ${BUILD_DIR}

test: ## Run the tests
	go test ./... -race -coverprofile=coverage.out

e2e: ## Run all e2e tests
	go test ./test/... --tags=e2e

coverage: test ## Print the code coverage
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out

build: ## Build the compass binary
	@echo "Building compass version ${VERSION}..."
	CGO_ENABLED=0 go build -ldflags "-X ${NAME}/cli.Version=${VERSION}"
	@echo "Build complete"

buildr: setup ## Build release snapshot
	goreleaser release --snapshot --skip=publish --clean

vet: ## Run go vet
	go vet ./...

download: ## Download go modules
	@go mod download

generate: ## Run all go generate in the code base
	@echo "Running go generate..."
	go generate ./...

config: ## Generate the sample config file
	@echo "Initializing sample server config..."
	@cp config/config.yaml config.yaml

proto: ## Generate the protobuf files
	@echo "Generating protobuf from raystack/proton"
	@echo " [info] make sure correct version of dependencies are installed using 'make setup'"
	@buf generate https://github.com/raystack/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path raystack/compass -v
	@echo "Protobuf compilation finished"

setup: ## Install required dependencies
	@echo "> Installing dependencies..."
	go mod tidy
	go install github.com/vektra/mockery/v2@latest
	go install github.com/bufbuild/buf/cmd/buf@latest

swagger-md:
	npx swagger-markdown -i proto/compass.swagger.yaml -o docs/docs/reference/api.md

clean-doc:
	@echo "> cleaning up auto-generated docs"
	@rm -rf ./docs/docs/reference/cli.md
	@rm -f ./docs/docs/reference/api.md

doc: clean-doc swagger-md ## Generate api and cli references
	@echo "> generate cli docs"
	@go run . reference --plain | sed '1 s,.*,# CLI,' > ./docs/docs/reference/cli.md

help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'