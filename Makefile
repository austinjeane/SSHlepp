# SSHlepp Makefile

# Variables
BINARY_NAME=sshlepp
MAIN_PATH=./cmd/sshlepp
BUILD_DIR=build

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOMOD=$(GOCMD) mod

.PHONY: all build clean test coverage deps help install dev

all: build

## Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) $(MAIN_PATH)

## Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -rf $(BUILD_DIR)

## Run tests
test:
	$(GOTEST) -v ./...

## Run tests with coverage
coverage:
	$(GOTEST) -coverprofile=coverage.out ./...
	$(GOCMD) tool cover -html=coverage.out

## Download dependencies
deps:
	$(GOMOD) download
	$(GOMOD) tidy

## Install the application
install: build
	sudo mv $(BINARY_NAME) /usr/local/bin/

## Development build with race detection
dev:
	$(GOBUILD) -race -o $(BINARY_NAME) $(MAIN_PATH)

## Build for multiple platforms
build-all: build-linux build-darwin build-windows

build-linux:
	mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-darwin:
	mkdir -p $(BUILD_DIR)
	GOOS=darwin GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-windows:
	mkdir -p $(BUILD_DIR)
	GOOS=windows GOARCH=amd64 $(GOBUILD) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

## Run the application
run: build
	./$(BINARY_NAME)

## Format code
fmt:
	$(GOCMD) fmt ./...

## Lint code
lint:
	golangci-lint run

## Show help
help:
	@echo 'Usage:'
	@sed -n 's/^##//p' ${MAKEFILE_LIST} | column -t -s ':' |  sed -e 's/^/ /'
