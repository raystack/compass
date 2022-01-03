FROM alpine:latest

COPY columbus /usr/bin/columbus
RUN apk update
RUN apk add ca-certificates

CMD ["columbus"]
