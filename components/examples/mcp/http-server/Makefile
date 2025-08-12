.PHONY: all build

default: all

all: build

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world mcp-http-server -o mcp-server.wasm main.go
