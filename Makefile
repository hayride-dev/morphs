.PHONY: build $(SUBDIRS) register compose register-default-tools register-datetime register-default-agent register-llama register-inmemory register-cli compose-cli compose-server register-composed

SUBDIRS := $(shell find . -mindepth 1 -maxdepth 4 -type d -exec test -f '{}/Makefile' \; -print)

all: build

build:
	@for dir in $(SUBDIRS); do \
		echo "==> Building in $$dir"; \
		$(MAKE) -C $$dir build; \
	done

register-default-tools:
	hayride register --bin ./components/mcp/tools/tools.wasm --package hayride:default-tools@0.0.1

register-datetime:
	hayride register --bin ./components/util/datetime/target/wasm32-wasip2/release/datetime.wasm --package hayride:datetime@0.0.1

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

register-mcp-server:
	hayride register --bin ./components/examples/mcp/http-server/mcp-server.wasm --package hayride:mcp-http-server@0.0.1

register-ory-auth:
	hayride register --bin ./components/mcp/auth/ory-auth.wasm --package hayride:mcp-ory-auth@0.0.1

register: register-default-tools register-datetime register-default-agent register-llama register-inmemory register-runner register-cli register-http register-mcp-server register-ory-auth

compose-cli:
	hayride wac compose --path ./compositions/default-agent-cli.wac --out ./compositions/composed-cli-agent.wasm

compose-http:
	hayride wac compose --path ./compositions/default-agent-http.wac --out ./compositions/composed-http-agent.wasm

compose-mcp-http-server:
	hayride wac compose --path ./compositions/mcp-server.wac --out ./compositions/composed-mcp-server.wasm

compose: compose-cli compose-http compose-mcp-http-server

register-cli-agent:
	hayride register --bin ./compositions/composed-cli-agent.wasm --package hayride:composed-cli-agent@0.0.1

register-http-agent:
	hayride register --bin ./compositions/composed-http-agent.wasm --package hayride:composed-http-agent@0.0.1

register-mcp-server-composed:
	hayride register --bin ./compositions/composed-mcp-server.wasm --package hayride:composed-mcp-server@0.0.1

register-composed: register-cli-agent register-http-agent register-mcp-server-composed
