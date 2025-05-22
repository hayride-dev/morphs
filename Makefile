.PHONY: build $(SUBDIRS)

SUBDIRS := $(shell find . -mindepth 1 -maxdepth 4 -type d -exec test -f '{}/Makefile' \; -print)

all: build

build:
	@for dir in $(SUBDIRS); do \
		echo "==> Building in $$dir"; \
		$(MAKE) -C $$dir build; \
	done

compose: 
	wac plug ./components/ai/agents/default.wasm --plug ./components/ai/tools/datetime/target/wasm32-wasip2/release/datetime.wasm --plug ./components/ai/contexts/inmemory.wasm --plug ./components/ai/models/llama-3.1.wasm -o composed-agent.wasm 
	wac plug ./components/ai/runners/cli.wasm --plug composed-agent.wasm -o cli-agent.wasm