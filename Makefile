NAME="github.com/odpf/columbus"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/columbus.coverprofile"

.PHONY: all build test clean

all: build

build:
	go build -ldflags "-X cmd.Version=${VERSION}" ${NAME}

clean:
	rm -rf columbus dist/

test:
	go test ./... -coverprofile=coverage.out

test-coverage: test
	go tool cover -html=coverage.out

dist:
	@bash ./scripts/build.sh

lint:
	golangci-lint run

