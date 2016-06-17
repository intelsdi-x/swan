package main

import (
	"fmt"
	"os"
	"os/user"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/cassandra"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/shopspring/decimal"
)

// AggressorsFlag represents flag specifying aggressors to run experiment with.
var AggressorsFlag = conf.NewSliceFlag(
	"aggr", "Aggressor to run experiment with. "+
		"You can state as many as you want (--aggr=l1d --aggr=membw)")

// Check README.md for details of this experiment.
func main() {
	// Setup conf.
	conf.SetAppName("MemcachedWithMutilateToCassandra")
	conf.SetHelpPath(
		path.Join(fs.GetSwanExperimentPath(), "memcached", "llc_aggr_cassandra", "README.md"))

	// Parse CLI.
	err := conf.ParseFlagAndEnv()
	if err != nil {
		logrus.Fatal(err)
	}

	aggressorsSet := make(map[string]struct{})
	for _, aggr := range AggressorsFlag.Value() {
		aggressorsSet[aggr] = struct{}{}
	}
	logrus.SetLevel(conf.LogLevel())

	numaZero := isolation.NewIntSet(0)
	// Initialize Memcached Launcher with HP isolation.
	hpCpus, err := isolation.CPUSelect(4, isolation.ShareLLCButNotL1L2, isolation.NewIntSet())
	if err != nil {
		logrus.Fatal(err)
	}

	hpIsolation, err := cgroup.NewCPUSet("hp", hpCpus, numaZero, true, false)
	if err != nil {
		logrus.Fatal(err)
	}

	err = hpIsolation.Create()
	if err != nil {
		logrus.Fatal(err)
	}

	defer hpIsolation.Clean()

	localForHP := executor.NewLocalIsolated(hpIsolation)
	memcachedLauncher := memcached.New(localForHP, memcached.DefaultMemcachedConfig())

	// Special case to have ability to use local executor for load generator.
	// This is needed for docker testing.
	var loadGeneratorExecutor executor.Executor
	loadGeneratorExecutor = executor.NewLocal()

	if workloads.LoadGeneratorAddrFlag.Value() != "local" {
		// Initialize Mutilate Launcher.
		user, err := user.Current()
		if err != nil {
			logrus.Fatal(err)
		}

		sshConfig, err := executor.NewSSHConfig(
			workloads.LoadGeneratorAddrFlag.Value(), executor.DefaultSSHPort, user)
		if err != nil {
			logrus.Fatal(err)
		}

		loadGeneratorExecutor = executor.NewRemote(sshConfig)
	}

	percentile, _ := decimal.NewFromString("99.9")
	mutilateConfig := mutilate.Config{
		MutilatePath:      mutilate.GetPathFromEnvOrDefault(),
		MemcachedHost:     "127.0.0.1",
		LatencyPercentile: percentile,
		TuningTime:        1 * time.Second,
	}
	mutilateLoadGenerator := mutilate.New(loadGeneratorExecutor, mutilateConfig)

	// Initialize BE isolation.
	beCpus, err := isolation.CPUSelect(4, isolation.ShareLLCButNotL1L2, hpCpus)
	if err != nil {
		logrus.Fatal(err)
	}

	beIsolation, err := cgroup.NewCPUSet("be", beCpus, numaZero, true, false)
	if err != nil {
		logrus.Fatal(err)
	}

	err = beIsolation.Create()
	if err != nil {
		logrus.Fatal(err)
	}

	defer beIsolation.Clean()

	localForBE := executor.NewLocalIsolated(beIsolation)

	// Initialize aggressors with BE isolation.
	aggressors := []sensitivity.LauncherSessionPair{}

	// TODO(bp): Make a factory for aggressors and use it here.
	if _, ok := aggressorsSet[l1data.ID]; ok {
		// l1data.
		aggressors = append(aggressors,
			sensitivity.NewLauncherWithoutSession(
				l1data.New(localForBE, l1data.DefaultL1dConfig())))
	}

	if _, ok := aggressorsSet[l1instruction.ID]; ok {
		// l1instruction.
		aggressors = append(aggressors,
			sensitivity.NewLauncherWithoutSession(
				l1instruction.New(localForBE, l1instruction.DefaultL1iConfig())))
	}

	if _, ok := aggressorsSet[memoryBandwidth.ID]; ok {
		// memBW.
		aggressors = append(aggressors,
			sensitivity.NewLauncherWithoutSession(
				memoryBandwidth.New(localForBE, memoryBandwidth.DefaultMemBwConfig())))

	}

	if _, ok := aggressorsSet[caffe.ID]; ok {
		// caffe.
		aggressors = append(aggressors,
			sensitivity.NewLauncherWithoutSession(
				caffe.New(localForBE, caffe.DefaultConfig())))
	}

	if _, ok := aggressorsSet[l3data.ID]; ok {
		// llc.
		aggressors = append(aggressors,
			sensitivity.NewLauncherWithoutSession(
				l3data.New(localForBE, l3data.DefaultL3Config())))
	}

	// Create connection with Snap.
	logrus.Debug("Connecting to Snapd on ", snap.AddrFlag.Value())
	// TODO(bp): Make helper for passing host:port or only host option here.
	snapConnection, err := client.New(
		fmt.Sprintf("http://%s:%s", snap.AddrFlag.Value(), snap.DefaultDaemonPort),
		"v1",
		true,
	)
	if err != nil {
		logrus.Fatal(err)
	}

	// Load the snap cassandra publisher plugin if not yet loaded.
	// TODO(bp): Make helper for that.
	logrus.Debug("Checking if publisher cassandra is loaded.")
	plugins := snap.NewPlugins(snapConnection)
	loaded, err := plugins.IsLoaded("publisher", "cassandra")
	if err != nil {
		logrus.Fatal(err)
	}

	if !loaded {
		pluginPath := []string{path.Join(
			os.Getenv("GOPATH"), "bin", "snap-plugin-publisher-cassandra")}
		err = plugins.Load(pluginPath)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	// Define publisher.
	publisher := wmap.NewPublishNode("cassandra", 2)
	if publisher == nil {
		logrus.Fatal("Failed to create Publish Node for cassandra")
	}

	publisher.AddConfigItem("server", cassandra.AddrFlag.Value())

	// Initialize Mutilate Snap Session.
	mutilateSnapSession := sessions.NewMutilateSnapSessionLauncher(
		fs.GetSwanBuildPath(),
		1*time.Second,
		snapConnection,
		publisher)

	// Create Experiment configuration.
	configuration := sensitivity.Configuration{
		SLO:             500, // us
		LoadDuration:    10 * time.Second,
		LoadPointsCount: 10,
		Repetitions:     3,
	}

	sensitivityExperiment := sensitivity.NewExperiment(
		conf.AppName(),
		conf.LogLevel(),
		configuration,
		sensitivity.NewLauncherWithoutSession(memcachedLauncher),
		sensitivity.NewMonitoredLoadGenerator(mutilateLoadGenerator, mutilateSnapSession),
		aggressors,
	)

	// Run Experiment.
	err = sensitivityExperiment.Run()
	if err != nil {
		logrus.Fatal(err)
	}
}
