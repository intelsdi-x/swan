package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	//"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils/os"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	//"os"
	//"path"
	//"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	"os/user"
	"time"
)

const (
	defaultSSHPort            = 22
	defaultMemcachedThreads   = 10
	defaultMutilatePercentile = "99"
	defaultTuningTime         = 1
	defaultLoadDuration       = 1
	defaultSLO                = 500
	defaultRepetitionsNumber  = 10
)

// Check README.md for details of this experiment.
func main() {
	logLevel := logrus.InfoLevel
	logrus.SetLevel(logLevel)

	numaZero := isolation.NewIntSet(0)
	hpCpus := isolation.NewIntSet(0, 1, 2, 3, 4, 20, 21, 22, 23, 24)
	beCpus := isolation.NewIntSet(5, 6, 7, 8, 9, 25, 26, 27, 28, 29)

	// Create new Cpu sets.
	hpIsolation, err := cgroup.NewCPUSet("hp", hpCpus, numaZero, true, false)
	beIsolation, err := cgroup.NewCPUSet("be", beCpus, numaZero, true, false)
	hpIsolation.Create()
	defer hpIsolation.Clean()
	beIsolation.Create()
	defer beIsolation.Clean()

	// Create local executors with given isolation.
	localHPIsolated := executor.NewLocalIsolated(hpIsolation)
	localBEIsolated := executor.NewLocalIsolated(beIsolation)

	// Initialize Memcached Launcher.
	conf := memcached.DefaultMemcachedConfig()
	conf.NumThreads = defaultMemcachedThreads
	memcachedLauncher := memcached.New(localHPIsolated, conf)

	// Initialize Mutilate Launcher.
	percentile, err := decimal.NewFromString(defaultMutilatePercentile)
	if err != nil {
		panic(err)
	}

	memcachedHost := os.GetEnvOrDefault("SWAN_MEMCACHED_HOST", "127.0.0.1")
	mutilateHost := os.GetEnvOrDefault("SWAN_MUTILATE_HOST", "127.0.0.1")
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     memcachedHost,
		LatencyPercentile: percentile,
		TuningTime:        defaultTuningTime * time.Second,
	}

	// Get current user for ssh config.
	user, err := user.Current()
	if err != nil {
		panic(err)
	}

	// Create ssh config and remote executor with this config.
	sshConfig, err := executor.NewSSHConfig(mutilateHost, defaultSSHPort, user)
	if err != nil {
		panic(err)
	}
	remote := executor.NewRemote(*sshConfig)

	// Create mutilate.
	mutilateLoadGenerator := mutilate.New(remote, mutilateConfig)

	// Create connection with Snap.
	logrus.Debug("Connecting to Snap")
	snapAddress := os.GetEnvOrDefault("SWAN_SNAP_ADDRESS", "http://127.0.0.1:8181")
	snapConnection, err := client.New(snapAddress, "v1", true)
	if err != nil {
		panic(err)
	}

	// Define publisher.
	publisher := wmap.NewPublishNode("cassandra", 2)
	if publisher == nil {
		panic("Failed to create Publish Node for cassandra")
	}

	cassandraHostName := os.GetEnvOrDefault("SWAN_CASSANDRA_HOST", "127.0.0.1")
	publisher.AddConfigItem("server", cassandraHostName)

	// Initialize Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		fs.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Initialize aggressors.
	llcAggressorLauncher := l3data.New(localBEIsolated, l3data.DefaultL3Config())
	memBwAggressorLauncher := memoryBandwidth.New(localBEIsolated, memoryBandwidth.DefaultMemBwConfig())
	l1iAggressorLauncher := l1instruction.New(localBEIsolated, l1instruction.DefaultL1iConfig())
	lidAggressorLauncher := l1data.New(localBEIsolated, l1data.DefaultL1dConfig())
	caffeAggressorLauncher := caffe.New(localBEIsolated, caffe.DefaultConfig())

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             defaultSLO,
		LoadDuration:    defaultLoadDuration * time.Second,
		LoadPointsCount: defaultRepetitionsNumber,
		Repetitions:     defaultRepetitionsNumber,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		"MemcachedWithLocalMutilateToCassandra",
		logLevel,
		configuration,
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession),
		[]sensitivity.LauncherSessionPair{
			sensitivity.NewLauncherWithoutSession(l1iAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(lidAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(llcAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(memBwAggressorLauncher),
			sensitivity.NewLauncherWithoutSession(caffeAggressorLauncher),
		},
	)

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		panic(err)
	}
}
