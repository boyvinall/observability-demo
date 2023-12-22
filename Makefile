define PROMPT
	@echo
	@echo "**********************************************************"
	@echo "*"
	@echo "*   $(1)"
	@echo "*"
	@echo "**********************************************************"
	@echo
endef

.PHONY: all # Default target - run lint and build
all: lint docker-build start

GRPC_PROTO=\
	pkg/boomer/boomer.proto

.PHONY: build # Build the application code
build: generate
	$(call PROMPT,$@)
	CGO_ENABLED=0 go build -o bin/boomer ./cmd/boomer-server/

.PHONY: docker-build # Build the docker image
docker-build:
	$(call PROMPT,$@)
	docker compose build

.PHONY: generate # Generate code from proto files
generate: \
	$(GRPC_PROTO:.proto=.pb.go) \
	$(GRPC_PROTO:.proto=_grpc.pb.go)

.PHONY: download-dependencies # Download go dependencies
download-dependencies:
	$(call PROMPT,$@)
	@go mod download -x

.PHONY: install-tools # Install required build tools
install-tools: download-dependencies
	$(call PROMPT,$@)
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %

.PHONY: clean # Delete build artifacts
clean:
	$(call PROMPT,$@)
	rm -rf bin
	rm -rf pkg/boomer/*.pb.go

.PHONY: lint # Run linter on the application code
lint:
	$(call PROMPT,$@)
	golangci-lint run

%_grpc.pb.go: %.proto Makefile
	$(call PROMPT,$@)
	protoc --go-grpc_out=paths=source_relative:. $<

%.pb.go: %.proto Makefile
	$(call PROMPT,$@)
	protoc --go_out=paths=source_relative:. $<

.PHONY: start # Start the docker-compose stack
start:
	$(call PROMPT,$@)
	docker compose up -d --remove-orphans

.PHONY: stop # Stop the docker-compose stack and remove volumes
stop:
	$(call PROMPT,$@)
	docker compose down -v --remove-orphans -t2

.PHONY: run-client # Run boomer client, requires the stack to be running
run-client:
	$(call PROMPT,$@)
	go run ./cmd/boomer-cli/main.go

.PHONY: help # Show help
help:
	@echo "Usage: make [target]"
	@echo
	@grep ^.PHONY Makefile | sed 's/\.PHONY: \(.*\) # \(.*\)/  \1\t\2/' | expand -t25
