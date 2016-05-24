package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	"io/ioutil"
	"os"
	"path"
	"time"
)

const (
	defaultMemcachedPath = "workloads/data_caching/memcached/memcached-1.4.25/build/memcached"
	defaultMutilatePath  = "workloads/data_caching/memcached/mutilate/mutilate"
	defaultL3dataPath    = "workloads/low-level-aggressors/l3"
	swanPkg              = "github.com/intelsdi-x/swan"
)

// TODO(bp): Create helper code for these fetchers.
func fetchMemcachedPath() string {
	// Get optional custom Memcached path from MEMCACHED_PATH.
	memcachedPath := os.Getenv("MEMCACHED_BIN")

	if memcachedPath == "" {
		// If custom path does not exists use default path for built memcached.
		return path.Join(os.Getenv("GOPATH"), "src", swanPkg, defaultMemcachedPath)
	}

	return memcachedPath
}

func fetchMutilatePath() string {
	// Get optional custom mutilate path from MUTILATE_BIN.
	mutilatePath := os.Getenv("MUTILATE_BIN")

	if mutilatePath == "" {
		// If custom path does not exists use default path for built mutilate.
		return path.Join(os.Getenv("GOPATH"), "src", swanPkg, defaultMutilatePath)
	}

	return mutilatePath
}

func fetchLLCPath() string {
	// Get optional custom mutilate path from LLC_BIN.
	llcPath := os.Getenv("LLC_BIN")

	if llcPath == "" {
		// llcPath custom path does not exists use default path for built l3.
		return path.Join(os.Getenv("GOPATH"), "src", swanPkg, defaultL3dataPath)
	}

	return llcPath
}

// This Experiments contains:
// - memcached as LC task on localhost
// - mutilate as loadGenerator on localhost
// - no aggressors so far
func main() {
	logLevel := logrus.DebugLevel
	logrus.SetLevel(logLevel)

	local := executor.NewLocal()
	// Init Memcached Launcher.
	memcachedLauncher := memcached.New(local,
		memcached.DefaultMemcachedConfig(fetchMemcachedPath()))
	// Init Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")
	mutilateConfig := mutilate.Config{
		MutilatePath:      fetchMutilatePath(),
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

	buildPath := path.Join(os.Getenv("GOPATH"), "src", swanPkg, "build")
	if !loaded {
		pluginPath := []string{path.Join(buildPath, "snap-plugin-publisher-session-test")}
		plugins.Load(pluginPath)

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

	// Init Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		buildPath, 1*time.Second, snapConnection, publisher)

	// Init LLC aggressor.
	llcAggressorLauncher := l3data.New(local, l3data.DefaultL3Config(fetchLLCPath()))

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             1000, // TODO: make this variable precise (us?)
		LoadDuration:    5 * time.Second,
		LoadPointsCount: 1,
		Repetitions:     1,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		"MemcachedWithLocalMutilateWithLLCggr",
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
