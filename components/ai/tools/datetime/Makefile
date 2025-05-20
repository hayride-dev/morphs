.PHONY: all build gen

default: all

all: build

build: ; cargo build --target wasm32-wasip2 --release

gen-go: ; wit-bindgen-go generate --world hayride:datetime/imports --out ./pkg/internal/gen ./wit

gen-rust-exports: ; cargo component bindings

gen-rust-imports:
	cd crates/datetime-imports && \
	cargo component bindings

gen: gen-go gen-rust-exports gen-rust-imports
