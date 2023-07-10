export GO111MODULE=on

PACKAGE=ddm-admin-console/config

VERSION?=$(shell git describe --tags)
BUILD_DIR="."
BUILD_DATE=$(shell date -u +'%Y-%m-%d %H:%M:%S')
GIT_COMMIT=$(shell git rev-parse HEAD)
GIT_TAG=$(shell if [ -z "`git status --porcelain`" ]; then git describe --exact-match --tags HEAD 2>/dev/null; fi)

override LDFLAGS += \
  -w -s \
  -X ${PACKAGE}.version=${VERSION} \
  -X '${PACKAGE}.buildDate=${BUILD_DATE}' \
  -X ${PACKAGE}.gitCommit=${GIT_COMMIT} \

ifneq (${GIT_TAG},)
LDFLAGS += -X ${PACKAGE}.gitTag=${GIT_TAG}
endif

.PHONY: all
all:
	deps gen lint unit-test build

.PHONY: ci
ci:
	lint unit-test build

.PHONY: build
build:
	CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -a -installsuffix cgo -ldflags="${LDFLAGS}" -o ${BUILD_DIR}

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