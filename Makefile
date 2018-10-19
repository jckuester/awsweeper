TEST?=$$(go list ./... | grep -v 'vendor')
PKG_LIST := $(shell go list ./...)
GOFMT_FILES?=$$(find . -name '*.go' | grep -v vendor)

default: build

testacc:
	TF_ACC=1 go test $(TEST) -v $(TESTARGS) -timeout 120m

vet:
	@echo "go vet ."
	@go vet $$(go list ./... | grep -v vendor/) ; if [ $$? -eq 1 ]; then \
		echo ""; \
		echo "Vet found suspicious constructs. Please check the reported constructs"; \
		echo "and fix them if necessary before submitting the code for review."; \
		exit 1; \
	fi

fmt:
	gofmt -w $(GOFMT_FILES)

modules:
	go mod tidy

generate:
	go generate

build:
	go build

.PHONY: test
test: generate
	go clean -testcache ${PKG_LIST}
	go test -short --race ${PKG_LIST}