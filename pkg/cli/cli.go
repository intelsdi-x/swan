package cli

import (
	"fmt"
	"github.com/codegangsta/cli"
	"github.com/intelsdi-x/snap/plugin/helper"
	"k8s.io/kubernetes/third_party/golang/go/doc/testdata"
	"os"
)

type Helper struct {
	key          string
	defaultValue string
	helperString string
	isRequired   Key // I'm not sure about this swich.
	// We have defaults in most places now,
	// but we can be open for changes.
}

type Key int

const (
	OPTIONAL Key = iota
	REQUIRED Key = iota
)

type Cli struct {
	helpers []Helper
	readme  string

	canRun bool
}

func (c Cli) AddReadme(readme string) {
	c.readme = readme
}

func (c Cli) AddHelper(helper Helper) {
	c.helpers = append(c.helpers, helper)
}

func (c Cli) CanRunExperiment() {
	return c.canRun
}

func (c Cli) ParseArgs(argv []string) {
	if argumentIsShowHelp(argv) {
		c.showHelp()
		c.canRun = false
	}

	parameters := parseParams(argv)

	for key, value := range parameters {
		// This way we can pass arguments from CLI to our deepest Launchers.
		// We can let them configure themselves.
		os.Setenv(key, value)
	}

	c.checkIfRequiredParamsAreSet() // throw panic if not
}

func argumentIsShowHelp(argv []string) {
	return len(argv) > 1 && (argv[1] == "--help" || argv[1] == "-h")
}

// Print all nested helpers
func (c Cli) showHelp() {
	fmt.Printf("%s\n", c.readme)
	for _, helper := range c.helpers {
		fmt.Printf("%s %s %s required: %v",
			helper.key,
			helper.helperString,
			helper.defaultValue,
			helper.isRequired)
	}
}

func parseParams(argv []string) map[string]string {
	var params map[string]string
	// Map CLI arguments like SWAN_CASSANDRA_ADDRESS=10.4.1.1 into key-value pairs
	return params
}

func (c Cli) checkIfRequiredParamsAreSet() {
	for _, param := range c.helpers {
		if param.isRequired {
			if os.Getenv(param.key) == "" {
				panic("Key " + param.key + " is not set!")
			}
		}
	}
}
