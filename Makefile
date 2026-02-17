.PHONY: build clean install test run build-windows build-linux build-darwin build-all

BINARY_NAME=cipherhub
MAIN_PATH=./cmd/cipherhub
VERSION?=1.0.0

build:
	go build -o bin/$(BINARY_NAME) $(MAIN_PATH)

build-windows:
	GOOS=windows GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY_NAME)-windows-amd64.exe $(MAIN_PATH)

build-windows-arm64:
	GOOS=windows GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY_NAME)-windows-arm64.exe $(MAIN_PATH)

build-linux:
	GOOS=linux GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY_NAME)-linux-amd64 $(MAIN_PATH)

build-linux-arm64:
	GOOS=linux GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY_NAME)-linux-arm64 $(MAIN_PATH)

build-darwin:
	GOOS=darwin GOARCH=amd64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY_NAME)-darwin-amd64 $(MAIN_PATH)

build-darwin-arm64:
	GOOS=darwin GOARCH=arm64 go build -ldflags="-s -w -X main.version=$(VERSION)" -o bin/$(BINARY_NAME)-darwin-arm64 $(MAIN_PATH)

build-all: build-windows build-linux build-darwin

install:
	go install $(MAIN_PATH)

clean:
	rm -rf bin/
	go clean

test:
	go test -v ./...

run:
	go run $(MAIN_PATH)

deps:
	go mod download
	go mod tidy

lint:
	golangci-lint run ./...

fmt:
	go fmt ./...

.PHONY: all
all: deps fmt build test
