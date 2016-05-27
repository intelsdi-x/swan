package main

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
)

// Check README.md for details of this experiment.
func main() {
	logLevel := logrus.ErrorLevel
	logrus.SetLevel(logLevel)

	local := executor.NewLocal()
	// Init Memcached Launcher.
	memcachedLauncher := memcached.New(local,
		memcached.DefaultMemcachedConfig())
	// Init Mutilate Launcher.
	percentile, _ := decimal.NewFromString("99.9")

	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     "127.0.0.1",
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}

	mutilateLoadGenerator := mutilate.New(local, mutilateConfig)

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             1000, // TODO: make this variable precise (us?)
		LoadDuration:    5 * time.Second,
		LoadPointsCount: 1,
		Repetitions:     1,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		"MemcachedWithLocalMutilateNoCollection",
		logLevel,
		configuration,
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewLoadGeneratorWithoutSession(mutilateLoadGenerator),
		[]sensitivity.LauncherSessionPair{},
	)

	// Run Experiment.
	err := sensitivityExperiment.Run()
	if err != nil {
		panic(err)
	}
}
