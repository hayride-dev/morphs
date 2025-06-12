.PHONY: all gen build

default: all

all: gen build

gen:
	wit-bindgen-go generate --world hayride:agents/default --out ./internal/gen ./wit

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world default
