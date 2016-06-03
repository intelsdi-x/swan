package main

import "github.com/codegangsta/cli"

const (
	// Experiment sets the Local/Remote executors, so he need to gather their addresses from environment
	loadGeneratorLocationKey = "SWAN_LOAD_GENERATOR_LOCATION"
	loadGeneratorLocationDefault = "127.0.0.1"
)

// Example usage of experiment:
// ./experiment --help
// ./experiment SWAN_CASSANDRA_ADDRESS=10.4.1.1
// We can also implement ./experiment cassandra_address=10.4.1.1
// or https://github.com/kelseyhightower/envconfig to reduce prefixes
func main() {
	cli.AddReadme("Some link to readme")
	cli.AddHelper(thisExperiment.Helper()) // Most imporant part
	dontRun := cli.ParseArgs(argv) // If our enviroment lack required variable, we throw error
	if (dontRun) { // don't run when user just wanted to see help
		return 0
	}

	// rest of experiment goes here

	// At this moment all Launchers are configured by themselves.
	// Experiment nor CLI does not directly changes any enviroment configuration in them.
	// Here, Environment Variables are my IOC Container that injects configs where it has to be injected.
	// Simple, and with no entangling
}

// Helper function informs user of enviroment variables that are used by this Experiment and underlying launchers
// It should be part of Experiment Interface
func (e Experiment) Helper() []cli.Helper {
	return {
		cli.Helper{loadGeneratorLocationKey, loadGeneratorLocationDefault, "Mutilate host", cli.OPTIONAL},
		e.load_generator.Helper(), // We can extract helpers from Experiment's launcher
		e.lc_workload.Helper(),
		e.be_workload.Helper(),
	}
}
// Helper returs list of used enviroment variables in format
// <KEY, DEFAULT_VALUE, Description for experiment user, Optional/required>
// Optional/required - if required key is non-existent, CLI will throw error
