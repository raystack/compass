NAME="github.com/odpf/compass"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/compass.coverprofile"
PROTON_COMMIT := "4d2fb0f0b145c31c02ccd65fb4a83510d58712e2"

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
	go install github.com/vektra/mockery/v2@v2.12.2
	go install google.golang.org/protobuf/cmd/protoc-gen-go@v1.28.0
	go get google.golang.org/protobuf/proto@v1.28.0
	go get google.golang.org/grpc@v1.46.0
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.9.0
	go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.9.0
	go install github.com/bufbuild/buf/cmd/buf@v1.4.0
	go install github.com/envoyproxy/protoc-gen-validate@v0.6.7
