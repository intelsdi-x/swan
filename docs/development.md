<!--
 Copyright (c) 2017 Intel Corporation

 Licensed under the Apache License, Version 2.0 (the "License");
 you may not use this file except in compliance with the License.
 You may obtain a copy of the License at

      http://www.apache.org/licenses/LICENSE-2.0

 Unless required by applicable law or agreed to in writing, software
 distributed under the License is distributed on an "AS IS" BASIS,
 WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 See the License for the specific language governing permissions and
 limitations under the License.
-->

# ![Swan logo](/images/swan-logo-48.png) Swan

# Developer's guide

To run Vagrant development environment, the `$SWAN_DEVELOPEMENT_ENVIRONMENT` variable must be set.

```bash
export SWAN_DEVELOPEMENT_ENVIRONMENT=true
cd vagrant
vagrant up
```

If Vagrant machine is already provisioned through [Quick Start](/README.md#quick-start) procedure, it should be re-provisioned.

```bash
export SWAN_DEVELOPEMENT_ENVIRONMENT=true
cd vagrant
vagrant provision
vagrant ssh
```

Also, thou [Quick Start](/README.md#quick-start) procedure proposes `git clone` of repository, for development purposes it is recommended that repository should be put in `$GOPATH/github.com/intelsdi-x/swan`.

[Golang](https://golang.org/dl/) 1.7.5+ is required for Swan development.


## Development using Makefile

Before sending or updating pull requests, make sure to run:

test & build & run
```bash
$ make build_and_test_all
```

### Vagrant (Virtualbox) development environment

Instead of building Swan in your own environment, you can build and run it in a development VM.
Follow the [Vagrant instructions](../misc/dev/vagrant/singlenode/README.md) to
create a Linux virtual machine pre-configured for developing Swan.

### Detailed options for tests
```bash
$ make test TEST_OPT="-v -run <specific test>"
```

## Development using go binaries

As an alternative to using our Makefiles, you can test, lint and build manually:

```bash
go test ./pkg/...
golint ./pkg/...
```

## Dependency management

Dependency management in Swan is handled by [glide](https://github.com/Masterminds/glide). Please refer to the glide documentation for further dependency managenement.

*Please note* that the <swan>/vendor directory shall not be commited to the repository.

## Tests

Pull requests should ship with tests which exercise the proposed code.
Swan use [goconvey](https://github.com/smartystreets/goconvey) as a framework for behavioral tests.

### Integration tests

For Swan integration tests see [here](testing.md) file for instructions.

### Mock generation

Mock generation is done by Mockery tool.
Sometimes Mockery is not able to resolve all imports in file correctly.
Developer needs to use it manually, that's why we are vendoring our mocks.

To generate mocks go to desired package and ```mockery -name=".*" -inpkg -case=underscore```

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
 - Keep pull requests small and split them into several pull requests. If there are dependencies, name them `<NAME>: [1/N] <short info>`, `<NAME>: [2/N] <short info>`, and so on.

### Continuous integration

We use Jenkins as a pre check for our pull requests. By default, linters checks and unit tests will be run. For linters check we use [golint](https://github.com/golang/lint) and [go vet](https://golang.org/cmd/vet/) tools.
Add comment `run integration test` in PR to run integration tests.
All tests must pass to merge PR.

### Reviewing

For general guidelines around code reviews, please refer this [document](http://kevinlondon.com/2015/05/05/code-review-best-practices.html).
For best practices for Go lang coding style and reviews, please refer to this [document](https://github.com/golang/go/wiki/CodeReviewComments) and
this [document](https://golang.org/doc/effective_go.html#introduction) as a go to resource for idiomatic Go code.
Many code style issues should be caught by our linter and `gofmt`. We don't try to make exceptions which can't be automated.

### Merging

Before merging, the PR needs at least one 'LGTM' from a maintainer.

## Coding practices

### Error Handling

We use `github.com/pkg/errors` library for errors. Here are guidelines on how to use the package from [source](http://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package)

- In your own code, use errors.New or errors.Errorf at the point an error occurs.
```
    func parseArgs(args []string) error {
            if len(args) < 3 {
                    return errors.Errorf("not enough arguments, expected at least 3, got %d", len(args))
            }
            // ...
    }
```

- If you receive an error from another function, it is often sufficient to simply return it.
```
    if err != nil {
           return err
    }
```
- If you interact with a package from another repository, consider using errors.Wrap or errors.Wrapf to establish a stack trace at that point. This advice also applies when interacting with the standard library.
```
    f, err := os.Open(path)
    if err != nil {
            return errors.Wrapf(err, "failed to open %q", path)
    }
```
- Always return errors to their caller rather than logging them throughout your program.
    At the top level of your program, or worker goroutine, use %+v to print the error with sufficient detail.
```
    func main() {
            err := app.Run()
            if err != nil {
                    fmt.Printf("FATAL: %+v\n", err)
                    os.Exit(1)
            }
    }
```
- If you want to exclude some classes of error from printing, use errors.Cause to unwrap errors before inspecting them.

**Errod code style:**
- Start error message from lower case and do not end it with dot
- Error message should contain parameters given to function that has returned error
- Parameters should be printed with `%q`

**Sources:**
- http://dave.cheney.net/2016/06/12/stack-traces-and-the-errors-package
- http://dave.cheney.net/2016/04/27/dont-just-check-errors-handle-them-gracefully
