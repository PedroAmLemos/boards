# Makefile for building, running, and cleaning the Go project

# check if in windows to make boards.exe
ifeq ($(OS),Windows_NT)
	BINARY_NAME=boards.exe
else
	BINARY_NAME=boards
endif

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
