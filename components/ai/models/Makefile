.PHONY: all test build

default: all

all: build

test: ; go test

build: ; tinygo build -tags=llama3 -target wasip2 --wit-package ./wit/ --wit-world llm -o llama31.wasm .
