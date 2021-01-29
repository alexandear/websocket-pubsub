MAKEFILE_PATH := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
PATH := $(MAKEFILE_PATH):$(PATH)

export GOBIN := $(MAKEFILE_PATH)/bin
export GOFLAGS = -mod=vendor

.PHONY: default
default: build lint

.PHONY: build
build:
	@echo build
	@go build -o ./bin/pubsub .

.PHONY: vendor
vendor:
	@echo vendor
	@-rm -rf vendor/
	@go mod vendor

.PHONY: lint
lint:
	@echo lint
	@go install github.com/golangci/golangci-lint/cmd/golangci-lint
	@$(GOBIN)/golangci-lint run

.PHONY: run-server
run-server:
	@echo run-server
	@go run -race . server --broadcast 2s

.PHONY: run-client
run-client:
	@echo run-client
	@go run -race . client --clients 4
