NAME="github.com/odpf/compass"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/compass.coverprofile"
PROTON_COMMIT := "c7639b42da0679b2340a52155d2fe577b9d45aa2"
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
	go install github.com/vektra/mockery/v2@v2.14.0
	go get google.golang.org/protobuf/proto@v1.28.0
	go get google.golang.org/grpc@v1.46.0
	go install github.com/bufbuild/buf/cmd/buf@v1.4.0

update-swagger-md:
	npx swagger-markdown -i proto/compass.swagger.yaml -o docs/docs/reference/api.md
