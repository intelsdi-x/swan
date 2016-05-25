package main

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	viewer          = kingpin.New("SensitivityViewer", "Simple command-line tool for viewing Sensitivity experiment results.")
	cassandraServer = viewer.Flag("cassandra_host", "IP and Port of Cassandra DB with results.").String()

	listExperimentsCmd = viewer.Command("list", "List all experiment UUIDs")

	showExperimentDataCmd = viewer.Command("show", "Get Experiment Results for specific experiment UUID")
	experimentID          = showExperimentDataCmd.Arg("experiment_uuid", "Experiment UUID").Required().String()
)

func listExperiments() {
	// TODO
	fmt.Println("LIST")
}

func showExperiment(experimentID string) {
	// TODO
	fmt.Println("SHOW")
}

// Run via: go run scripts/sensitivity_viewer/main.go
func main() {
	switch kingpin.MustParse(viewer.Parse(os.Args[1:])) {
	// List experiments.
	case listExperimentsCmd.FullCommand():
		listExperiments()

	// Show experiment data for specified experimentID.
	case showExperimentDataCmd.FullCommand():
		showExperiment(*experimentID)
	}
}
