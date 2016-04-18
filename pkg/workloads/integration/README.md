# Workload integration tests.

In this directory there all integration tests for each workload. These tests requires some
setup before execution like special packages and building the workload binary.

After setup you can run them in following manner:

`go test -tags=integration`

To create integration test file we use build tags, so you need to place

```
// +build integration
<newline>
package integration
```

(Make sure you place newline between package name and build flag)
