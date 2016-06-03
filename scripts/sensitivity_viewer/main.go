package main

import (
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"gopkg.in/alecthomas/kingpin.v2"
	"os"
)

var (
	viewer          = kingpin.New("SensitivityViewer", "Simple command-line tool for viewing Sensitivity experiment results.")
	cassandraServer = viewer.Flag("cassandra_host", "host for Cassandra DB with results.").Default("127.0.0.1").String()

	listExperimentsCmd = viewer.Command("list", "List all experiment UUIDs")

	showExperimentDataCmd     = viewer.Command("show", "Get experiment results for specific experiment UUID")
	showSensitivityProfileCmd = viewer.Command("sensitivity", "Draw sensitivity profile for specific experiment UUID")
	showExperimentID          = showExperimentDataCmd.Arg("experiment_uuid", "Experiment UUID").Required().String()
	sensitivityExperimentID   = showSensitivityProfileCmd.Arg("experiment_uuid", "Experiment UUID").Required().String()
)

func listExperiments() {
	err := experiment.List(*cassandraServer)
	if err != nil {
		panic(err)
	}
}

// TODO(ala) sort table based on timestamp.
func showExperiment() {
	err := experiment.Draw(*showExperimentID, *cassandraServer)
	if err != nil {
		panic(err)
	}
}

// TODO(ala) create also CSV exporter of Sensitivity Profile.
func showSensitivityProfile() {
	err := sensitivity.Draw(*sensitivityExperimentID, *cassandraServer)
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

	// Show sensitivity profile for specific experimentID.
	case showSensitivityProfileCmd.FullCommand():
		showSensitivityProfile()
	}
}
