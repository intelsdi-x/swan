.PHONY: build

all: lint unit_test build

# deps not covered by "vendor" folder (testing/developing env) rather than application (excluding convey)
deps:
	go get github.com/tools/godep
	go get github.com/golang/lint/golint
	go get github.com/GeertJohan/fgt # return exit, fgt runs any command for you and exits with exitcode 1
	go get github.com/stretchr/testify
	go get github.com/vektra/mockery/.../
	godep restore -v

# testing
## fgt: lint doesn't return exit code when finds something (https://github.com/golang/lint/issues/65)
lint:
	fgt golint ./pkg/...
	fgt golint ./cmds/...

unit_test:
	go test ./pkg/...
	go test ./cmds/...

test:
	go test -tags=integration ./pkg/...
	go test ./cmds/...

# building
build:
	mkdir -p build
	(cd build; go build ../cmds/...)

run: memcache

memcache:
	./memcache

