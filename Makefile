NAME="github.com/raystack/compass"
COMMIT := $(shell git rev-parse --short HEAD)
TAG := "$(shell git rev-list --tags --max-count=1)"
VERSION := "$(shell git describe --tags ${TAG})-next"
BUILD_DIR=dist
PROTON_COMMIT := "ccbf219312db35a934361ebad895cb40145ca235"

.PHONY:  all build clean test tidy vet proto setup format generat

all: clean test build format lint

tidy:
	@echo "Tidy up go.mod..."
	@go mod tidy -v

install:
	@echo "Installing Guardian to ${GOBIN}..."
	@go install

format:
	@echo "Running go fmt..."
	@go fmt

lint: ## Lint checker
	@echo "Running lint checks using golangci-lint..."
	@golangci-lint run

clean: tidy ## Clean the build artifacts
	@echo "Cleaning up build directories..."
	@rm -rf $coverage.out ${BUILD_DIR}

test:  ## Run the tests
	go test ./... -race -coverprofile=coverage.out

e2e: ## Run all e2e tests
	go test ./test/... --tags=e2e

coverage: test ## Print the code coverage
	@echo "Generating coverage report..."
	@go tool cover -html=coverage.out

build: ## Build the compass binary
	@echo "Building guardian version ${VERSION}..."
	CGO_ENABLED=0 go build -ldflags "-X ${NAME}/cli.Version=${VERSION}"
	@echo "Build complete"

buildr: setup
	goreleaser --snapshot --skip-publish --rm-dist


vet:
	go vet ./...

download:
	@go mod download


generate: ## Run all go generate in the code base
	@echo "Running go generate..."
	go generate ./...

config: ## Generate the sample config file
	@echo "Initializing sample server config..."
	@cp config/config.yaml config.yaml


proto: ## Generate the protobuf files
	@echo "Generating protobuf from raystack/proton"
	@echo " [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/raystack/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path raystack/compass -v
	@echo "Protobuf compilation finished"

setup: ## Install required dependencies
	@echo "> Installing dependencies..."
	go mod tidy
	go install github.com/vektra/mockery/v2@v2.14.0
	go install google.golang.org/protobuf/proto@v1.28.0
	go install google.golang.org/grpc@v1.46.0
	go install github.com/bufbuild/buf/cmd/buf@v1.4.0

swagger-md:
	npx swagger-markdown -i proto/compass.swagger.yaml -o docs/docs/reference/api.md

clean-doc:
	@echo "> cleaning up auto-generated docs"
	@rm -rf ./docs/docs/reference/cli.md
	@rm -f ./docs/docs/reference/api.md

doc: clean-doc update-swagger-md ## Generate api and cli references
	@echo "> generate cli docs"
	@go run . reference --plain | sed '1 s,.*,# CLI,' > ./docs/docs/reference/cli.md
 
help: ## Display this help message
	@cat $(MAKEFILE_LIST) | grep -e "^[a-zA-Z_\-]*: *.*## *" | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'