package main

import (
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	viewer          = kingpin.New("SensitivityViewer", "Simple command-line tool for viewing Sensitivity experiment results.")
	cassandraServer = viewer.Flag("cassandra_host", "IP and Port of Cassandra DB with results.").Default("127.0.0.1").String()

	listExperimentsCmd = viewer.Command("list", "List all experiment UUIDs")

	showExperimentDataCmd = viewer.Command("show", "Get Experiment Results for specific experiment UUID")
	experimentID          = showExperimentDataCmd.Arg("experiment_uuid", "Experiment UUID").Required().String()
)

func listExperiments() {
	err := cassandra.DrawList(*cassandraServer)
	if err != nil {
		panic(err)
	}
}

func showExperiment() {
	err := cassandra.DrawTable(*experimentID, *cassandraServer)
	if err != nil {
		panic(err)
	}
}

// Run via: go run scripts/sensitivity_viewer/main.go
func main() {
	switch kingpin.MustParse(viewer.Parse(os.Args[1:])) {
	// List experiments.
	case listExperimentsCmd.FullCommand():
		listExperiments()

	// Show experiment data for specified experimentID.
	case showExperimentDataCmd.FullCommand():
		showExperiment()
	}
}
