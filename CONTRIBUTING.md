# ![Swan logo](/images/swan-logo-48.png) Swan

* [Contributing Code](#contributing-code)
* [Contributing Experiments](#contributing-experiments)
* [Integration Tests](#integration-tests)
* [Thank You](#thank-you)

## Contributing Code
**_IMPORTANT_**: Whenever you contribute to the project by creating pull request, sending patches or doing other activities that will change code or documentation in repository you agree to publish your work on Apache 2.0 license.

**_IMPORTANT_**: We encourage contributions to the project from the community. We ask that you keep the following guidelines in mind when planning your contribution.

* Whether your contribution is for a bug fix or a feature request, **create an [Issue](https://github.com/intelsdi-x/swan/issues)** and let us know what you are thinking.
* **For bugs**, if you have already found a fix, feel free to submit a Pull Request referencing the Issue you created. Include the `Fixes #` syntax to link it to the issue you're addressing.
* **For feature requests**, we want to improve upon Swan incrementally which means small changes at a time. In order to ensure your PR can be reviewed in a timely manner, please keep PRs small, e.g. <10 files and <1000 lines changed. If you think this is unrealistic, then mention that within the issue and we can discuss it.

Once you're ready to contribute code back to this repo, start with these steps:

* Fork the appropriate sub-projects that are affected by your change.
* Clone the fork to `$GOPATH/src/github.com/intelsdi-x/`:

```
$ cd "${GOPATH}/src/github.com/intelsdi-x/"
$ git clone https://github.com/intelsdi-x/swan.git
```

* Create a topic branch for your change and checkout that branch:

```
$ git checkout -b some-topic-branch
```

* Make your changes to the code and add tests to cover contributed code.
* Run `make` to validate it will not break current functionality.
* Commit your changes and push them to your fork.
* Open a pull request for the appropriate project.
* Contributors will review your pull request, suggest changes, run integration tests and eventually merge or close the request.

## Integration Tests

Due to them taking a significant amount of time to run, integration tests are manually triggered. Once your code has been reviewed and you are ready to move forward, you or a maintainer can start an integration test by commenting on the pull request with `run integration tests`.

![Running an Integration Test](https://cloud.githubusercontent.com/assets/1744971/25538578/24197b8a-2bf8-11e7-90c1-05e76b45ce77.png)

## Contributing Experiments
The most immediately helpful way you can benefit this project is by adding your own experiments. The current list is in the [experiments/ folder](/experiments/). We would be glad to help you share these, so please open an Issue if you run into any challenges doing so.

## Thank You
And **thank you!** Your contribution, through code and participation, is appreciated.
