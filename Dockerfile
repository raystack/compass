FROM alpine:latest

COPY compass /usr/bin/compass
RUN apk update
RUN apk add ca-certificates

CMD ["compass"]
