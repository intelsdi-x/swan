package main

import (
	"github.com/intelsdi-x/swan/pkg/visualization"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	viewer          = kingpin.New("SensitivityViewer", "Simple command-line tool for viewing Sensitivity experiment results.")
	cassandraServer = viewer.Flag("cassandra_host", "host for Cassandra DB with results.").Default("127.0.0.1").String()

	listExperimentsCmd = viewer.Command("list", "List all experiment UUIDs")

	showExperimentDataCmd     = viewer.Command("show", "Get Experiment Results for specific experiment UUID")
	showSensitivityProfileCmd = viewer.Command("sensitivity", "Draw sensitivity profile for specific experiment UUID")
	showExperimentID          = showExperimentDataCmd.Arg("experiment_uuid", "Experiment UUID").Required().String()
	sensitivityExperimentID   = showSensitivityProfileCmd.Arg("experiment_uuid", "Experiment UUID").Required().String()
)

func listExperiments() {
	err := visualization.DrawList(*cassandraServer)
	if err != nil {
		panic(err)
	}
}

func showExperiment() {
	err := visualization.DrawTable(*showExperimentID, *cassandraServer)
	if err != nil {
		panic(err)
	}
}

func showSensitivityProfile() {
	err := visualization.DrawSensitivityProfile(*sensitivityExperimentID, *cassandraServer)
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

	//
	case showSensitivityProfileCmd.FullCommand():
		showSensitivityProfile()
	}
}
