MAKEFILE_PATH := $(abspath $(dir $(abspath $(lastword $(MAKEFILE_LIST)))))
PATH := $(MAKEFILE_PATH):$(PATH)

export GOBIN := $(MAKEFILE_PATH)/bin
export GOFLAGS = -mod=vendor

PATH := $(GOBIN):$(PATH)

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

.PHONY: test
test:
	@echo test
	@go test -race -v -count=1 ./...

.PHONY: server
server:
	@echo server
	@go run -race . server --broadcast 2s

.PHONY: client
client:
	@echo client
	@go run -race . client --clients 4

.PHONY: generate
generate: mock

.PHONY: mock
mock:
	@echo mock
	@go install github.com/golang/mock/mockgen
	@go generate ./...
