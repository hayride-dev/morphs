.PHONY: build $(SUBDIRS)

SUBDIRS := $(shell find . -mindepth 1 -maxdepth 3 -type d -exec test -f '{}/Makefile' \; -print)

all: build

build:
	@for dir in $(SUBDIRS); do \
		echo "==> Building in $$dir"; \
		$(MAKE) -C $$dir build; \
	done
