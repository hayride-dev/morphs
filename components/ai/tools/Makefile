.PHONY: all test build

default: all

all: build

gen:
	wit-bindgen-go generate --world hayride:tools/noop --out ./internal/gen ./wit

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world noop tools.go
