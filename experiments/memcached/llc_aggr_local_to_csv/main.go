package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/swan"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"path"
	"time"
)

// This Experiments contains:
// - memcached as LC task on localhost
// - mutilate as loadGenerator on localhost
// - Snap Session: Mutilate -> CSV Publisher
// - LLC aggressor
func main() {
	logLevel := logrus.DebugLevel
	logrus.SetLevel(logLevel)

	local := executor.NewLocal()

	// Initialize Memcached Launcher.
	memcachedLauncher := memcached.New(local, memcached.DefaultMemcachedConfig())

	// Initialize Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     "127.0.0.1",
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}

	mutilateLoadGenerator := mutilate.New(local, mutilateConfig)

	// Create connection with Snap.
	logrus.Debug("Connecting to Snap")
	snapConnection, err := client.New("http://127.0.0.1:8181", "v1", true)
	if err != nil {
		panic(err)
	}

	// Load the snap session test plugin if not yet loaded.
	// TODO(bp): Make helper for that.
	logrus.Debug("Checking if publisher session test is loaded.")
	plugins := snap.NewPlugins(snapConnection)
	loaded, err := plugins.IsLoaded("publisher", "session-test")
	if err != nil {
		panic(err)
	}

	if !loaded {
		pluginPath := []string{path.Join(
			swan.GetSwanBuildPath(), "snap-plugin-publisher-session-test")}
		err = plugins.Load(pluginPath)
		if err != nil {
			panic(err)
		}
	}

	// Define publisher.
	publisher := wmap.NewPublishNode("session-test", 1)
	if publisher == nil {
		panic("Failed to create Publish Node for session-test")
	}
	tmpFile, err := ioutil.TempFile("", "MemcachedWithLocalMutilateWithLLCAggr")
	if err != nil {
		panic(err)
	}
	tmpFile.Close()
	publisher.AddConfigItem("file", tmpFile.Name())
	logrus.Debug("Results should be available in publisher's file: ", tmpFile.Name())

	// Initialize Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		swan.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Initialize LLC aggressor.
	llcAggressorLauncher := l3data.New(local, l3data.DefaultL3Config())

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             500, // us
		LoadDuration:    10 * time.Second,
		LoadPointsCount: 10,
		Repetitions:     3,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		"MemcachedWithLocalMutilateToCSV",
		logLevel,
		configuration,
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession),
		[]sensitivity.LauncherSessionPair{
			sensitivity.NewLauncherWithoutSession(llcAggressorLauncher),
		},
	)

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		panic(err)
	}
}
