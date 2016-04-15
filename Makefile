all:
	$(MAKE) deps
	$(MAKE) test

deps:
	./scripts/deps.sh

lint:
	fgt golint ./pkg/...

test:
	go test ./pkg/...
