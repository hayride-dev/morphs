.PHONY: all test build

default: all

all: build

gen:
	wit-bindgen-go generate --world hayride:models/llm --out ./internal/gen ./wit

build: ; tinygo build -tags=llama3 -target wasip2 --wit-package ./wit/ --wit-world llm -o llama31.wasm .
