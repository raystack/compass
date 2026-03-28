# Release stage — used by goreleaser (binary pre-built)
FROM alpine:3.21 AS release
COPY compass /usr/bin/compass
RUN apk add --no-cache ca-certificates
ENTRYPOINT ["compass"]

# Build stage — compiles from source
FROM golang:1.25-alpine3.21 AS builder
WORKDIR /build/
COPY . .
RUN CGO_ENABLED=0 go build -ldflags "-X github.com/raystack/compass/cli.Version=${VERSION}" -o compass

# Dev stage — default target, builds from source
FROM alpine:3.21
COPY --from=builder /build/compass /usr/bin/compass
RUN apk add --no-cache ca-certificates libc6-compat
ENTRYPOINT ["compass"]
