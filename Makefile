all:
	$(MAKE) deps
	$(MAKE) test

deps:
	./scripts/deps.sh

lint:
	golint ./pkg/...

test:
	go test ./pkg/...
