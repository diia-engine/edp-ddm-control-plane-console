export GO111MODULE=on

.PHONY: all
all:
	deps gen lint unit-test build

.PHONY: ci
ci:
	lint unit-test build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -mod=vendor -a -installsuffix cgo -ldflags="-w -s"

.PHONY: deps
deps:
	go mod tidy
	go mod download
	go mod vendor

.PHONY: unit-test
unit-test:
	go test -v -cover ./...

.PHONY: lint
lint:
	golangci-lint run