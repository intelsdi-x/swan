all:
	$(MAKE) deps
	$(MAKE) test

deps:
	./scripts/deps.sh

lint:
	./scripts/lint.sh

test:
	./scripts/test.sh
