.PHONY: all build gen

default: all

all: build

build: ; cargo build --target wasm32-wasip2

gen: ; wit-bindgen-go generate --world hayride:datetime/imports --out ./pkg/internal/gen ./wit

