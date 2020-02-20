TEST?=$$(go list ./...)
PKG_LIST := $(shell go list ./...)

.PHONY: setup
setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.11
	go get -u github.com/golang/mock/mockgen

.PHONY: modules
modules:
	go mod download

.PHONY: generate
generate:
	mkdir -p resource/mocks
	go generate

.PHONY: build
build:
	go build

.PHONY: test
test: ## Run unit tests
	go clean -testcache ${PKG_LIST}
	go test -v -p 1 -short -race ${PKG_LIST}

.PHONY: test-all
test-all: ## Run tests (including acceptance and integration tests)
	go clean -testcache ${PKG_LIST}
	./bin/go-acc ${PKG_LIST} -- -v -p 1 -race -failfast -timeout 30m

.PHONY: lint
lint: generate
	./bin/golangci-lint run
