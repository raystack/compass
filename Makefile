NAME="github.com/odpf/columbus"
VERSION=$(shell git describe --always --tags 2>/dev/null)
COVERFILE="/tmp/columbus.coverprofile"

.PHONY: all build test clean

all: build

build: 
	go build -ldflags "-X main.Version=${VERSION}" ${NAME}/cmd/columbus
	

unit-test:
	@go list ./... | grep -v cmd | grep -v es | xargs go test -count 1 -cover -race -timeout 30s -coverprofile ${COVERFILE}
	@go tool cover -func ${COVERFILE} | tail -1 | xargs echo test coverage:

test:
	@go list ./... | grep -v cmd | xargs go test -count 1 -cover -race -timeout 30s -coverprofile ${COVERFILE}
	@go tool cover -func ${COVERFILE} | tail -1 | xargs echo test coverage:

clean:
	rm -rf columbus
