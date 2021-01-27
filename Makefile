MAKEFILE_PATH := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
PATH := $(MAKEFILE_PATH):$(PATH)

export GOBIN := $(MAKEFILE_PATH)/bin
export GOFLAGS = -mod=vendor

.PHONY: build
build: build-client build-server

.PHONY: build-client
build-client:
	@echo build-client
	@go build -o ./bin/client client/*

.PHONY: build-server
build-server:
	@echo build-server
	@go build -o ./bin/server server/*

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
