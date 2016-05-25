# Swan developer guide

## Development using Makefile

Before sending or updating pull requests, make sure to run:

test & build & run
```
$ make deps
$ make              # lint unit_test build
$ make run
```

### Detailed options for tests
```
$ make test TEST_OPT="-v -run <specific test>"
```

## Development using go binaries

As an alternative to using our Makefiles, you can test, lint and build manually:

```
go test ./pkg/...
golint ./pkg/...
go build ./cmds/memcached
```

## Dependency management

Dependency management in Swan is handled by [godeps](https://github.com/tools/godep).

Before submitting pull requests, make sure that you have saved any dependencies with:

```
godep save ./...
```

## Tests

Pull requests should ship with tests which exercise the proposed code.
Swan use [goconvey](https://github.com/smartystreets/goconvey) as a framework for behavioral tests.

### Integration tests

For Swan integration tests see [here](testing.md) file for instructions.

### Mock generation

Mock generation is done by Mockery tool.
Sometimes Mockery is not able to resolve all imports in file correctly.
Developer needs to use it manually, that's why we are vendoring our mocks.

To generate mocks go to desired package and ```mockery -name ".*" -case underscore```

In case of error: `could not import <pkg> (can't find import: <pkg>)`
Do the `go install <pkg>`

In some cases `go install` won't help (eg. you do not want to install the project that you are working on). In such a scenario you should follow instructions from [know mockery issue](https://github.com/vektra/mockery/issues/81).

## Submitting Pull Requests

### Naming

For consistency, we aim to name the pull request by prepending the JIRA issue to the title. For example:

`SCE-236: Fixed race condition for the local executor`

Don't include details about the PR state in the title, but use some of the existing labels. For example, avoid naming a PR:

`[RDY FOR REVIEW] ...`

But mark the PR as `ready for review` instead. If we are missing any labels, please reach out to the repository administrators.

### Description

In the pull request description, remember to:

 - Motivate the change with _why_ your code acts as it does. What is the problem it is trying to solve and how can a reviewer be sure that the pull request indeed fixes the issue.
 - Include a testing strategy that you used to gain confidence in the correctness of the code. This also enables the reviewer to replicate the issue.
 - Keep pull requests small and split them into several pull requests. If there are dependencies, name them `SCE-XXX: [1/N] Foo`, `SCE-XXX: [2/N] Bar`, and so on.

### Continuous integration

We use Jenkins as a pre check for our pull requests.
This, currently, does not carry out integration tests.

### Reviewing

For general guidelines around code reviews, please refer this [document](http://kevinlondon.com/2015/05/05/code-review-best-practices.html).
For best practices for Go lang coding style and reviews, please refer to this [document](https://github.com/golang/go/wiki/CodeReviewComments) and
this [document](https://golang.org/doc/effective_go.html#introduction) as a go to resource for idiomatic Go code.
Many code style issues should be caught by our linter and `gofmt`. We don't try to make exceptions which can't be automated.

### Merging

Before merging, the PR needs at least one 'LGTM' from a maintainer.

## Lab environment

Currently, we use manual partitioning of internal Intel lab environments.
Each developer has access to a set of machines until a scheduling system is in place.
The details on the cluster can be found [here](https://intelsdi.atlassian.net/wiki/display/KOP/intel.sdi.us_west01).
Please contact the cluster owner to get machines assigned to you.
