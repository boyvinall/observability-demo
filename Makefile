define PROMPT
	@echo
	@echo "**********************************************************"
	@echo "*"
	@echo "*   $(1)"
	@echo "*"
	@echo "**********************************************************"
	@echo
endef

.PHONY: all
all: lint build

GRPC_PROTO=\
	pkg/boomer/boomer.proto

.PHONY: build
build: generate
	$(call PROMPT,$@)
	CGO_ENABLED=0 go build -o bin/boomer ./cmd/boomer-server/

.PHONY: generate
generate: \
	$(GRPC_PROTO:.proto=.pb.go) \
	$(GRPC_PROTO:.proto=_grpc.pb.go)

.PHONY: download-dependencies
download-dependencies:
	$(call PROMPT,$@)
	@go mod download -x

.PHONY: install-tools
install-tools: download-dependencies
	$(call PROMPT,$@)
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: clean
clean:
	$(call PROMPT,$@)
	rm -rf bin
	rm -rf pkg/boomer/*.pb.go

.PHONY: lint
lint:
	$(call PROMPT,$@)
	golangci-lint run

%_grpc.pb.go: %.proto Makefile
	$(call PROMPT,$@)
	protoc --go-grpc_out=paths=source_relative:. $<

%.pb.go: %.proto Makefile
	$(call PROMPT,$@)
	protoc --go_out=paths=source_relative:. $<

.PHONY: start
start:
	$(call PROMPT,$@)
	docker compose up -d --build --remove-orphans

.PHONY: stop
stop:
	$(call PROMPT,$@)
	docker compose down -v --remove-orphans

.PHONY: run-client
run-client:
	$(call PROMPT,$@)
	go run ./cmd/boomer-cli/main.go
