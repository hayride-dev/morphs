.PHONY: all test build

default: all

all: build

test: ; tinygo test  -target wasip2 --wit-package ./wit/ --wit-world llm ./.

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world client