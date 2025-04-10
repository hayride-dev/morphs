.PHONY: all test build

default: all

all: build

TAGS ?= llama_3_1

MODEL ?= ./llama-3.1.go

test: ; tinygo test --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world llm ./.

build: ; tinygo build --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world llm ./$(MODEL)