package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	"os"
	"path"
	"time"
)

const (
	defaultMemcachedPath = "workloads/data_caching/memcached/memcached-1.4.25/build/memcached"
	defaultMutilatePath  = "workloads/data_caching/memcached/mutilate/mutilate"
	swanPkg              = "github.com/intelsdi-x/swan"
)

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
		// If custom path does not exists use default path for built memcached.
		return path.Join(os.Getenv("GOPATH"), "src", swanPkg, defaultMutilatePath)
	}

	return mutilatePath
}

// This Experiments contains:
// - memcached as LC task on localhost
// - mutilate as loadGenerator on localhost
// - no aggressors so far
func main() {
	logLevel := logrus.DebugLevel
	logrus.SetLevel(logLevel)

	// Init Memcached Launcher.
	memcachedLauncher := memcached.New(executor.NewLocal(),
		memcached.DefaultMemcachedConfig(fetchMemcachedPath()))
	// Init Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99")
	mutilateConfig := mutilate.Config{
		MutilatePath:      fetchMutilatePath(),
		MemcachedHost:     "127.0.0.1",
		LatencyPercentile: percentile,
		TuningTime:        5 * time.Second,
	}

	local := executor.NewLocal()
	local.OutputPrefix = "mutilate"
	mutilateLauncher := mutilate.New(local, mutilateConfig)

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             1000, // TODO: make this variable precise (us?)
		LoadDuration:    5 * time.Second,
		LoadPointsCount: 1,
		Repetitions:     1,
	}

	// Init Experiment.
	sensitivityExperiment, err := sensitivity.InitExperiment(
		"MemcachedWithLocalMutilate",
		logLevel,
		configuration,
		memcachedLauncher,
		mutilateLauncher,
		[]workloads.Launcher{},
	)

	if err != nil {
		panic(err)
	}

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		panic(err)
	}

}
