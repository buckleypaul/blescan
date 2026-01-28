.PHONY: build run clean test install

BINARY_NAME=blescan
VERSION?=0.1.0
LDFLAGS=-ldflags "-s -w -X main.version=$(VERSION)"

build:
	go build $(LDFLAGS) -o $(BINARY_NAME) ./cmd/blescan

run: build
	./$(BINARY_NAME)

clean:
	rm -f $(BINARY_NAME)
	go clean

test:
	go test -v ./...

install: build
	mv $(BINARY_NAME) $(GOPATH)/bin/

deps:
	go mod download
	go mod tidy
