.PHONY: build $(SUBDIRS) register compose register-datetime register-default-agent register-llama register-inmemory register-cli compose-cli compose-server

SUBDIRS := $(shell find . -mindepth 1 -maxdepth 4 -type d -exec test -f '{}/Makefile' \; -print)

all: build

build:
	@for dir in $(SUBDIRS); do \
		echo "==> Building in $$dir"; \
		$(MAKE) -C $$dir build; \
	done

register-datetime:
	hayride register --bin ./components/ai/tools/datetime/target/wasm32-wasip2/release/datetime.wasm --package hayride:datetime@0.0.1

register-default-agent:
	hayride register --bin ./components/ai/agents/default.wasm --package hayride:default-agent@0.0.1

register-llama:
	hayride register --bin ./components/ai/models/llama-3.1.wasm --package hayride:llama31@0.0.1

register-inmemory:
	hayride register --bin ./components/ai/contexts/inmemory.wasm --package hayride:inmemory@0.0.1

register-cli:
	hayride register --bin ./components/ai/runners/cli.wasm --package hayride:cli@0.0.1

register: register-datetime register-default-agent register-llama register-inmemory register-cli

compose-cli:
	hayride wac compose --path ./compositions/default-agent-cli.wac --out ./compositions/composed-cli-agent.wasm

compose-server:
	hayride wac compose --path ./compositions/default-agent-server.wac --out ./compositions/composed-server-agent.wasm

compose: compose-cli compose-server
