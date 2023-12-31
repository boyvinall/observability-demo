.PHONY: all # Default target - lint, build and start the stack
all: lint docker-build start

define PROMPT
	@echo
	@echo "**********************************************************"
	@echo "*"
	@echo "*   $(1)"
	@echo "*"
	@echo "**********************************************************"
	@echo
endef

.PHONY: help # Show help
help:
	@echo
	@echo "Usage: make [target...]"
	@echo
	@grep -E '^(.PHONY|#####)' Makefile | \
		sed \
			-e 's/\.PHONY: \(.*\) # \(.*\)/  \1\t\2/' \
			-e 's/^#####* \(.*\) #*$$/\n\1:\n/' | \
		expand -t25

############# Application targets #############

GRPC_PROTO=\
	pkg/boomer/boomer.proto

.PHONY: build # Build the application code
build: generate
	$(call PROMPT,$@)
	CGO_ENABLED=0 go build -o bin/boomer ./cmd/boomer-server/

.PHONY: lint # Run linter on the application code
lint: generate
	$(call PROMPT,$@)
	golangci-lint run

.PHONY: docker-build # Build local docker images
docker-build:
	$(call PROMPT,$@)
	docker compose build

.PHONY: generate # Generate code from proto files
generate: \
	$(GRPC_PROTO:.proto=.pb.go) \
	$(GRPC_PROTO:.proto=_grpc.pb.go)

%_grpc.pb.go: %.proto Makefile
	$(call PROMPT,$@)
	protoc --go-grpc_out=paths=source_relative:. $<

%.pb.go: %.proto Makefile
	$(call PROMPT,$@)
	protoc --go_out=paths=source_relative:. $<

.PHONY: clean # Delete build/generated artifacts
clean:
	$(call PROMPT,$@)
	rm -rf bin
	rm -rf pkg/boomer/*.pb.go
	rm -rf test/*.png

.PHONY: purge # More aggressive clean .. delete anything not in git
purge: clean
	$(call PROMPT,$@)
	git clean -fdx

############# Stack targets #############

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

.PHONY: browser-tests # Run browser tests against the stack
browser-tests:
	$(call PROMPT,$@)
	cd test && go run main.go

############# Tooling targets #############

.PHONY: download-dependencies # Download go dependencies
download-dependencies:
	$(call PROMPT,$@)
	go mod download -x

.PHONY: install-tools # Install required build tools
install-tools: download-dependencies
	$(call PROMPT,$@)
	@cat tools.go | grep _ | awk -F'"' '{print $$2}' | xargs -tI % go install %
