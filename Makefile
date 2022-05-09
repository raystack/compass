NAME="github.com/odpf/compass"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/compass.coverprofile"
PROTON_COMMIT := "9230539d922aaccaf4233ec771b4669c70b17ed0"

.PHONY: all build test clean install proto

all: build

build:
	go build -ldflags "-X cmd.Version=${VERSION}" ${NAME}

clean:
	rm -rf compass dist/

test:
	go test -race ./... -coverprofile=coverage.txt

test-coverage: test
	go tool cover -html=coverage.txt -o cover.html

e2e-test:
	go test ./test/... --tags=e2e

generate:
	go generate ./...

dist:
	@bash ./scripts/build.sh

lint:
	golangci-lint run
	
proto: ## Generate the protobuf files
	@echo " > generating protobuf from odpf/proton"
	@echo " > [info] make sure correct version of dependencies are installed using 'make install'"
	@buf generate https://github.com/odpf/proton/archive/${PROTON_COMMIT}.zip#strip_components=1 --template buf.gen.yaml --path odpf/compass -v
	@echo " > protobuf compilation finished"

install: ## install required dependencies
	@echo "> installing dependencies"
	go mod tidy
	go get github.com/vektra/mockery/v2@v2.10.4
	go get google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get google.golang.org/protobuf/proto@v1.27.1
	go get google.golang.org/grpc@v1.45.0
	go get google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.8.0
	go get github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.8.0
	go get github.com/bufbuild/buf/cmd/buf@v1.3.1
	go get github.com/envoyproxy/protoc-gen-validate@v0.6.7
