.PHONY: all test build

default: all

all: build

TAGS ?= cli

RUNNER ?= cli.go

WORLD ?= cli

build: ; tinygo build --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world $(WORLD) $(RUNNER)
