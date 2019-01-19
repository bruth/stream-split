FROM alpine:latest

RUN apk add --update curl && rm -rf /var/cache/apk/*

COPY ./dist/linux-amd64/stream-split /usr/bin
