.PHONY: all build build-basic build-vaunt gen wac wac-basic wac-vaunt

default: all

all: build wac

gen: ; wit-bindgen-go generate --world hayride:datetime/imports --out ./pkg/internal/gen ./wit

