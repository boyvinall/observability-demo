FROM golang:1.21.5-alpine AS builder

WORKDIR /app
RUN \
  apk add --no-cache make && \
  wget -O/tmp/protoc.zip https://github.com/protocolbuffers/protobuf/releases/download/v25.1/protoc-25.1-linux-$(uname -m | sed 's,aarch64,aarch_64,').zip && \
  unzip /tmp/protoc.zip -d /usr/local

COPY tools.go go.mod go.sum Makefile /app/
RUN make install-tools

COPY . /app/
RUN make build

FROM alpine:3.14.2
RUN apk add --no-cache ca-certificates
COPY --from=builder /app/bin/boomer /boomer
ENTRYPOINT [ "/boomer" ]
