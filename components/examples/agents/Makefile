.PHONY: all test build

default: all

all: build

build-cli:
	tinygo build --tags cli -target wasip2 --wit-package ./wit/ --wit-world cli cli.go

build-http:
	tinygo build --tags http -target wasip2 --wit-package ./wit/ --wit-world http http.go

build: build-cli build-http