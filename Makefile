NAME="github.com/odpf/columbus"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/columbus.coverprofile"
PROTON_COMMIT := "712a5a1ae39c6dbbd5a1e152c2c082dde18b5e79"

.PHONY: all build test clean install proto

all: build

build:
	go build -ldflags "-X cmd.Version=${VERSION}" ${NAME}

clean:
	rm -rf columbus dist/

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
	go get -d google.golang.org/protobuf/cmd/protoc-gen-go@v1.27.1
	go get google.golang.org/protobuf/proto@v1.27.1
	go get google.golang.org/grpc@v1.45.0
	go get -d google.golang.org/grpc/cmd/protoc-gen-go-grpc@v1.2.0
	go get -d github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@v2.8.0
	go get -d github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-openapiv2@v2.8.0
	go get -d github.com/bufbuild/buf/cmd/buf@v1.1.0
	go get -d github.com/envoyproxy/protoc-gen-validate@v0.6.7
