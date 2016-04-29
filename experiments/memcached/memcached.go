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

	local := executor.NewLocal()
	// Init Memcached Launcher.
	memcachedLauncher := memcached.New(local,
		memcached.DefaultMemcachedConfig(fetchMemcachedPath()))
	// Init Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")
	mutilateConfig := mutilate.Config{
		MutilatePath:      fetchMutilatePath(),
		MemcachedHost:     "localhost",
		LatencyPercentile: percentile,
		TuningTime:        30 * time.Second,
	}
	mutilateLauncher := mutilate.New(local, mutilateConfig)

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             1,
		LoadDuration:    30 * time.Second,
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
