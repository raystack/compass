FROM alpine:latest

COPY compass /usr/bin/compass
RUN apk update && apk add --no-cache ca-certificates && rm -rf /var/cache/apk/*

ENTRYPOINT ["compass"]
