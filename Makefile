PACKAGE=github.com/ismtabo/time-left
BINARY_NAME=time-left
TARGET_DIR=target
VERSION?=$(shell git describe --tags --match "*.*.*" --candidates 1)
ARCH=amd64
OUTPUTS=$(foreach os,${OS},${TARGET_DIR}/${BINARY_NAME}_${os})
GO_FILES=$(shell find . -type f -name '*.go' -and -not -path "./vendor/*" -and -not -name main.go)
ifeq ($(OS),Windows_NT)
	DETECTED_OS := windows
else
	DETECTED_OS := $(shell sh -c 'uname 2>/dev/null || echo Unknown' | tr '[:upper:]' '[:lower:]')
endif
OS?=${DETECTED_OS}

os=$(subst ${TARGET_DIR}/${BINARY_NAME}_,,$@)
date=$(shell date -u --iso-8601=minutes)
$(OUTPUTS): main.go ${GO_FILES}
	@mkdir -p ${TARGET_DIR}
	@echo "Building ${BINARY_NAME} for ${GOARCH} ${os}..."
	GOARCH=${ARCH} GOOS=${os} go build \
		-ldflags "-X '${PACKAGE}/config.Version=${VERSION}' \
			-X '${PACKAGE}/config.BuildTime=${date}' \
			-X '${PACKAGE}/config.OS=${os}'" \
		-o $@ $<

.PHONY: help
help:
	@echo "Usage: make <target>"
	@echo ""
	@echo "Targets:"
	@echo "  build         Build the binary for all supported platforms"
	@echo "  install       Install dependencies"
	@echo "  run           Run the application"
	@echo "  dev 				   Run the application in development mode"
	@echo "  clean         Clean the project"
	@echo "  test          Run tests"
	@echo "  test_coverage Run tests with coverage"
	@echo "  dep           Download dependencies"
	@echo "  vet           Run go vet"
	@echo "  lint          Run golangci-lint"

.PHONY: build
build: $(OUTPUTS)
	@chmod +x ${TARGET_DIR}/${BINARY_NAME}_*

.PHONY: install
install:
	go mod tidy

.PHONY: run
run:
	${TARGET_DIR}/${BINARY_NAME}_${DETECTED_OS}

.PHONY: clean
clean:
	go clean
	rm -rf ${TARGET_DIR}

.PHONY: test
test:
	go test ./...

.PHONY: test_coverage
test_coverage:
	go test ./... -coverprofile=coverage.out

.PHONY: dep
dep:
	go mod download

.PHONY: vet
vet:
	go vet

.PHONY: lint
lint:
	golangci-lint run --enable-all
