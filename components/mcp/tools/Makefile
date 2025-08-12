.PHONY: all test build

default: all

all: build

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world default tools.go
