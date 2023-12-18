//go:build tools
// +build tools

package tools

// Track tools here until https://github.com/golang/go/issues/48429 is resolved
// https://go.googlesource.com/proposal/+/refs/changes/55/495555/7/design/48429-go-tool-modules.md

import (
	_ "github.com/golangci/golangci-lint/cmd/golangci-lint"
	_ "google.golang.org/grpc/cmd/protoc-gen-go-grpc"
	_ "google.golang.org/protobuf/cmd/protoc-gen-go"
)
