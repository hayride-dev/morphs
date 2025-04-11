.PHONY: all test build

default: all

all: build

TAGS ?= basic

AGENT ?= basic.go

test: ; tinygo test --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world agent ./.

build: ; tinygo build --tags $(TAGS) -target wasip2 --wit-package ./wit/ --wit-world agent $(AGENT)
