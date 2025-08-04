.PHONY: all test build

default: all

all: build

test: ; go test

TAGS ?= llama3

OUTPUT ?= llama31.wasm

build: ; tinygo build -tags=$(TAGS) -target wasip2 --wit-package ./wit/ --wit-world llm -o $(OUTPUT) .
