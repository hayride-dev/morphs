.PHONY: all test build

default: all

all: build

gen:
	wit-bindgen-go generate --world hayride:contexts/in-memory --out ./internal/gen ./wit

build: ; tinygo build -target wasip2 --wit-package ./wit/ --wit-world in-memory inmemory.go
