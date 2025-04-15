.PHONY: all test build

default: all

all: build

TAGS ?= in_memory

CTX ?= inmemory.go

test: ; tinygo test --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world ctx ./.

build: ; tinygo build --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world ctx $(CTX)
