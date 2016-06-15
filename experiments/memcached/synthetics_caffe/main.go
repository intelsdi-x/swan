package main

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/utils/os"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
	"log"
	"os/user"
	"time"
)

const (
	defaultSSHPort       = 22
	memcachedThreads     = 10
	sloPercentile        = "99"
	tuningTime           = 1
	loadDuration         = 1
	slo                  = 500
	numberOfRepetitions  = 10
	defaultSnapAddress   = "http://127.0.0.1:8181"
	defaultCassandraHost = "127.0.0.1"
	defaultMutilateHost  = "127.0.0.1"
	defaultMemcachedHost = "127.0.0.1"
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
	hpIsolation.Create()
	defer hpIsolation.Clean()

	beIsolation, err := cgroup.NewCPUSet("be", beCpus, numaZero, true, false)
	beIsolation.Create()
	defer beIsolation.Clean()

	// Create local executors with given isolation.
	localHPIsolated := executor.NewLocalIsolated(hpIsolation)
	localBEIsolated := executor.NewLocalIsolated(beIsolation)

	// Initialize Memcached Launcher.
	conf := memcached.DefaultMemcachedConfig()
	conf.NumThreads = memcachedThreads
	memcachedLauncher := memcached.New(localHPIsolated, conf)

	// Initialize Mutilate Launcher.
	percentile, err := decimal.NewFromString(sloPercentile)
	if err != nil {
		panic(err)
		log.Fatalf("Retrieving decimal from given percentile %s ended with error %s", sloPercentile, err)
	}

	memcachedHost := os.GetEnvOrDefault("SWAN_MEMCACHED_HOST", defaultMemcachedHost)
	mutilateHost := os.GetEnvOrDefault("SWAN_MUTILATE_HOST", defaultMutilateHost)
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     memcachedHost,
		LatencyPercentile: percentile,
		TuningTime:        tuningTime * time.Second,
	}

	// Create ssh config and remote executor with this config.
	sshUserName := os.GetEnvOrDefault("SWAN_SSH_USER", "root")
	sshUser, err := user.Lookup(sshUserName)
	if err != nil {
		panic(err)
		log.Fatalf("Looking for a user %s ended with error %s", sshUserName, err)

	}
	sshConfig, err := executor.NewSSHConfig(mutilateHost, defaultSSHPort, sshUser)
	if err != nil {
		panic(err)
		log.Fatalf("Creating ssh config ended with error %s", err)
	}
	remote := executor.NewRemote(*sshConfig)

	// Create mutilate.
	mutilateLoadGenerator := mutilate.New(remote, mutilateConfig)

	// Create connection with Snap.
	snapAddress := os.GetEnvOrDefault("SWAN_SNAP_ADDRESS", defaultSnapAddress)
	logrus.Debug("Connecting to Snap at the address %s", snapAddress)
	snapConnection, err := client.New(snapAddress, "v1", true)
	if err != nil {
		panic(err)
		log.Fatalf("Connecting to snap at the address %s ended with error %s", snapAddress, err)

	}

	// Define publisher.
	publisher := wmap.NewPublishNode("cassandra", 2)
	if publisher == nil {
		panic("Failed to create Publish Node for cassandra")
	}

	cassandraHostName := os.GetEnvOrDefault("SWAN_CASSANDRA_HOST", defaultCassandraHost)
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
		SLO:             slo,
		LoadDuration:    loadDuration * time.Second,
		LoadPointsCount: numberOfRepetitions,
		Repetitions:     numberOfRepetitions,
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
