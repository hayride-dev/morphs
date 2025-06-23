.PHONY: all test build

default: all

all: build

gen:
	wit-bindgen-go generate --world hayride:models/llm --out ./internal/gen ./wit

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world llm llama-3.1.go
