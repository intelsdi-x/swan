.PHONY: build

all: lint test build  

# deps not covered by "vendor" folder (testing/developing env) rather than application (excluding convey)
deps:
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt # return exit, fgt runs any command for you and exits with exitcode 1

godep:
	go get github.com/tools/godep
	godep restore -v

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
lint:
	fgt golint ./pkg/...
	fgt golint ./cmds/...

test:
	go test ./pkg/...
	go test ./cmds/...

# building
build:
	mkdir -p build
	(cd build; go build ../cmds/...)

run: memcache

memcache:
	./memcache

	
