FROM --platform=amd64 golang:1.21.5-alpine AS builder

WORKDIR /app
ADD https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-x86_64.zip /tmp/protoc.zip
RUN \
  unzip /tmp/protoc.zip -d /usr/local
RUN apk add --no-cache make

