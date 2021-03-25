FROM golang:1.13-stretch as base
WORKDIR /build/
COPY . .
RUN ["make"]

FROM alpine:latest
WORKDIR /opt/columbus
COPY --from=base /build/columbus /opt/columbus/bin/columbus
RUN ["apk", "update"]
EXPOSE 8080

# glibc compatibility library, since go binaries 
# don't work well with musl libc that alpine uses
RUN ["apk", "add", "libc6-compat"] 
ENTRYPOINT ["/opt/columbus/bin/columbus"]