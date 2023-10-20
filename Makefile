# Makefile for building, running, and cleaning the Go project

# Variables
BINARY_NAME=boards

# Default target to build the binary
all: build

# Target to build the binary
build:
	go mod download
	go build -o $(BINARY_NAME)

# Target to clean the binary
clean:
	go clean

# Target to run tests
test:
	go test -v

# Target to install dependencies
deps:
	go mod download
