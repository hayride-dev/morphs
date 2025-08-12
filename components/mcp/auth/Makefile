.PHONY: all build

default: all

all: build

TAGS ?= ory

OUTPUT ?= ory-auth.wasm

build: ; tinygo build -tags=$(TAGS) -target wasip2 --wit-package ./wit/ --wit-world auth -o $(OUTPUT) .
