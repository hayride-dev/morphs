.PHONY: all test build

default: all

all: build build-rust

test: ; go test

TAGS ?= llama_3_1

OUTPUT ?= llama31.wasm

build: ; tinygo build -tags=$(TAGS) -target wasip2 --wit-package ./wit/ --wit-world llm -o $(OUTPUT) .

build-rust: ; cargo build --target wasm32-wasi --release