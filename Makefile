.PHONY: build $(SUBDIRS) register compose register-default-tools register-datetime register-default-agent register-llama register-inmemory register-cli compose-cli compose-server register-composed

SUBDIRS := $(shell find . -mindepth 1 -maxdepth 4 -type d -exec test -f '{}/Makefile' \; -print)

all: build

build:
	@for dir in $(SUBDIRS); do \
		echo "==> Building in $$dir"; \
		$(MAKE) -C $$dir build; \
	done

register-default-tools:
	hayride register --bin ./components/ai/tools/tools.wasm --package hayride:default-tools@0.0.1

register-datetime:
	hayride register --bin ./components/ai/tools/datetime/target/wasm32-wasip2/release/datetime.wasm --package hayride:datetime@0.0.1

register-default-agent:
	hayride register --bin ./components/ai/agents/default.wasm --package hayride:default-agent@0.0.1

register-llama:
	hayride register --bin ./components/ai/models/llama31.wasm --package hayride:llama31@0.0.1

register-inmemory:
	hayride register --bin ./components/ai/contexts/inmemory.wasm --package hayride:inmemory@0.0.1

register-runner:
	hayride register --bin ./components/ai/runners/default.wasm --package hayride:default-runner@0.0.1

register-cli:
	hayride register --bin ./components/examples/agents/cli.wasm --package hayride:cli@0.0.1

register-http:
	hayride register --bin ./components/examples/agents/http.wasm --package hayride:http@0.0.1

register: register-default-tools register-datetime register-default-agent register-llama register-inmemory register-runner register-cli register-http

compose-cli:
	hayride wac compose --path ./compositions/default-agent-cli.wac --out ./compositions/composed-cli-agent.wasm

compose-http:
	hayride wac compose --path ./compositions/default-agent-http.wac --out ./compositions/composed-http-agent.wasm

compose-server:
	hayride wac compose --path ./compositions/default-agent-server.wac --out ./compositions/composed-server-agent.wasm

compose: compose-cli compose-http compose-server

register-cli-agent:
	hayride register --bin ./compositions/composed-cli-agent.wasm --package hayride:composed-cli-agent@0.0.1

register-http-agent:
	hayride register --bin ./compositions/composed-http-agent.wasm --package hayride:composed-http-agent@0.0.1

register-server-agent:
	hayride register --bin ./compositions/composed-server-agent.wasm --package hayride:composed-server-agent@0.0.1

register-composed: register-cli-agent register-http-agent register-server-agent

