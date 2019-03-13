TEST?=$$(go list ./...)
PKG_LIST := $(shell go list ./...)

.PHONY: setup
setup:
	curl -sfL https://install.goreleaser.com/github.com/golangci/golangci-lint.sh | sh -s v1.11
	go get -u github.com/golang/mock/mockgen

.PHONY: modules
modules:
	go mod tidy

.PHONY: generate
generate:
	mkdir -p resource/mocks
	go generate

.PHONY: build
build:
	@rm -rf ./bin
	@mkdir bin
	@go build -o ./bin/app

.PHONY: generate
test: generate
	go clean -testcache ${PKG_LIST}
	go test --race -cover ${PKG_LIST}

.PHONY: testacc
testacc:
	TF_ACC=1 go test $(TEST) -cover -v $(TESTARGS) -timeout 120m

.PHONY: lint
lint: generate
	./bin/golangci-lint run
