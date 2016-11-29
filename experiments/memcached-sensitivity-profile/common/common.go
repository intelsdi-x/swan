package common

import (
	"fmt"
	"io"
	"os"
	"path"
	"strconv"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	mutilateMasterFlagDefault = "local"
)

var (

	// Mutilate configuration.
	percentileFlag     = conf.NewStringFlag("percentile", "Tail latency Percentile", "99")
	mutilateMasterFlag = conf.NewIPFlag(
		"mutilate_master",
		"Mutilate master host for remote executor. In case of 0 agents being specified it runs in agentless mode."+
			"Use `local` to run with local executor.",
		"127.0.0.1")
	mutilateAgentsFlag = conf.NewSliceFlag(
		"mutilate_agent",
		"Mutilate agent hosts for remote executor. Can be specified many times for multiple agents setup.")
)

// PrepareSnapMutilateSessionLauncher prepare a SessionLauncher that runs mutilate collector and records that into storage.
// Note: SnapdHTTPEndpoint set to "none" will disable mutilate session completely.
// TODO: this should be put into athena:/pkg/snap
func prepareSnapMutilateSessionLauncher() (snap.SessionLauncher, error) {
	// NOTE: For debug it is convenient to disable snap for some experiment runs.
	if snap.SnapdHTTPEndpoint.Value() != "none" {
		// Create connection with Snap.
		logrus.Info("Connecting to Snapd on ", snap.SnapdHTTPEndpoint.Value())
		// TODO(bp): Make helper for passing host:port or only host option here.

		mutilateConfig := mutilatesession.DefaultConfig()
		mutilateConfig.SnapdAddress = snap.SnapdHTTPEndpoint.Value()
		mutilateSnapSession, err := mutilatesession.NewSessionLauncher(mutilateConfig)
		if err != nil {
			return nil, err
		}
		return mutilateSnapSession, nil
	}
	return nil, nil
}

// prepareMutilateGenerator create new LoadGenerator based on mutilate.
func prepareMutilateGenerator(memcacheIP string, memcachePort int) (executor.LoadGenerator, error) {
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcacheIP
	mutilateConfig.MemcachedPort = memcachePort
	mutilateConfig.LatencyPercentile = percentileFlag.Value()

	// Special case to have ability to use local executor for mutilate master load generator.
	// This is needed for docker testing.
	var masterLoadGeneratorExecutor executor.Executor
	masterLoadGeneratorExecutor = executor.NewLocal()
	if mutilateMasterFlag.Value() != mutilateMasterFlagDefault {
		var err error
		masterLoadGeneratorExecutor, err = sensitivity.NewRemote(mutilateMasterFlag.Value())
		if err != nil {
			return nil, err
		}
	}

	// Pack agents.
	agentsLoadGeneratorExecutors := []executor.Executor{}
	for _, agent := range mutilateAgentsFlag.Value() {
		remoteExecutor, err := sensitivity.NewRemote(agent)
		if err != nil {
			return nil, err
		}
		agentsLoadGeneratorExecutors = append(agentsLoadGeneratorExecutors, remoteExecutor)
	}
	logrus.Debugf("Added %d mutilate agent(s) to mutilate cluster", len(agentsLoadGeneratorExecutors))

	// Validate mutilate cluster executors and their limit of
	// number of open file descriptors. Sane mutilate configuration requires
	// more than default (1024) for mutilate cluster.
	validate.ExecutorsNOFILELimit(
		append(agentsLoadGeneratorExecutors, masterLoadGeneratorExecutor),
	)

	// Initialize Mutilate Load Generator.
	mutilateLoadGenerator := mutilate.NewCluster(
		masterLoadGeneratorExecutor,
		agentsLoadGeneratorExecutors,
		mutilateConfig)

	return mutilateLoadGenerator, nil
}

// noopSessionLauncherFactory is a factory of snap.SessionLauncher that returns nothing.
func noopSessionLauncherFactory(_ sensitivity.Configuration) snap.SessionLauncher {
	return nil
}

// RunExperiment is main entrypoint to prepare and run experiment.
func RunExperiment() error {
	return RunExperimentWithMemcachedSessionLauncher(noopSessionLauncherFactory)
}

// RunExperimentWithMemcachedSessionLauncher is preparing all the components necessary to run experiment but uses memcachedSessionLauncherFactory
// to create a snap.SessionLauncher that will wrap memcached (HP workload).
// Note: it includes parsing the environment to get configuration as well as preparing executors and eventually running the experiment.
func RunExperimentWithMemcachedSessionLauncher(memcachedSessionLauncherFactory func(sensitivity.Configuration) snap.SessionLauncher) error {
	//experimentStartingTime := time.Now()
	conf.SetAppName("memcached-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)`)
	err := conf.ParseFlags()
	if err != nil {
		return err
	}
	logrus.SetLevel(conf.LogLevel())

	// Validate preconditions.
	validate.OS()

	// Isolations.
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Executors.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	fmt.Printf("%v", beExecutorFactory)
	if err != nil {
		return err
	}

	// BE workloads.
	beLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	if err != nil {
		cleanup()
		return err
	}
	beLaunchers = append([]sensitivity.LauncherSessionPair{sensitivity.LauncherSessionPair{}}, beLaunchers...)

	// Prepare experiment configuration to be used by session launcher factory.
	configuration := sensitivity.DefaultConfiguration()

	// HP workload.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := memcached.New(hpExecutor, memcachedConfig)

	// Load generator.
	loadGenerator, err := prepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	if err != nil {
		return err
	}

	snapSession, err := prepareSnapMutilateSessionLauncher()
	if err != nil {
		return err
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return errors.Wrap(err, "could not create uuid")
	}

	logrus.Info("Starting Experiment ", conf.AppName(), " with uuid ", uuid.String())
	fmt.Println(uuid.String())

	experimentDirectory := path.Join(os.TempDir(), conf.AppName(), uuid.String())
	err = os.MkdirAll(experimentDirectory, 0777)
	if err != nil {
		return errors.Wrap(err, "cannot create experiment directory")
	}
	err = os.Chdir(experimentDirectory)
	os.Chdir(os.TempDir())
	if err != nil {
		return errors.Wrap(err, "cannot chdir to experiment directory")
	}

	masterLogFilename := path.Join(experimentDirectory, "master.log")
	logFile, err := os.OpenFile(masterLogFilename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return errors.Wrapf(err, "could not open log file %q", masterLogFilename)
	}

	// Setup logging set to both output and logFile.
	logrus.SetLevel(conf.LogLevel())
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	achievedSLI := float64(sensitivity.PeakLoadFlag.Value())
	if sensitivity.PeakLoadFlag.Value() == sensitivity.RunTuningPhase {
		prTask, err := hpLauncher.Launch()
		if err != nil {
			return errors.Wrap(err, "cannot launch memcached")
		}
		stop := func() {
			prTask.Stop()
			prTask.Clean()
		}

		err = loadGenerator.Populate()
		if err != nil {
			stop()
			return errors.Wrap(err, "cannot populate memcached")
		}

		load, sli, err := loadGenerator.Tune(configuration.SLO)
		if err != nil {
			stop()
			return errors.Wrap(err, "tuning failed")
		}
		achievedSLI = float64(sli)

		stop()
		logrus.Infof("Run tuning and achieved following values: load - %d and SLI - %f", load, achievedSLI)
	} else {
		logrus.Infof("Skipping Tunning phase, using peakload %d", configuration.PeakLoad)
	}

	var bar *pb.ProgressBar
	totalPhases := configuration.LoadPointsCount * configuration.Repetitions * len(beLaunchers)
	if conf.LogLevel() == logrus.ErrorLevel {
		bar = pb.StartNew(totalPhases)
		bar.ShowCounters = false
		bar.ShowTimeLeft = true
	}

	for _, beLauncher := range beLaunchers {
		for loadPoint := 1; loadPoint <= configuration.LoadPointsCount; loadPoint++ {
			phaseName := fmt.Sprintf("Baseline, load point %d", loadPoint)
			for repetition := 0; repetition < configuration.Repetitions; repetition++ {
				var cleanupStack []func()
				if conf.LogLevel() == logrus.ErrorLevel {
					prefix := fmt.Sprintf("[%02d / %02d] %s ", loadPoint*configuration.Repetitions+repetition, totalPhases, phaseName)
					bar.Prefix(prefix)
					bar.Add(1)
				}

				logrus.Infof("Starting %s repetition %d", phaseName, repetition)

				_, err := createDirs(experimentDirectory, phaseName, repetition)
				if err != nil {
					return errors.Wrapf(err, "cannot create repetition log directory")
				}

				hpHandle, err := hpLauncher.Launch()
				if err != nil {
					return errors.Wrapf(err, "cannot launch memcached in baseline, load point %d, repetition %d", loadPoint, repetition)
				}
				cleanupStack = append(cleanupStack, func() {
					hpHandle.Stop()
					hpHandle.Clean()
				})

				err = loadGenerator.Populate()
				if err != nil {
					cleanEnvironment(cleanupStack)
					return errors.Wrapf(err, "cannot populate memcached in baseline, load point %d, repetition %d", loadPoint, repetition)
				}

				if beLauncher.Launcher != nil {
					beHandle, err := beLauncher.Launcher.Launch()
					if err != nil {
						cleanEnvironment(cleanupStack)
						return errors.Wrapf(err, "cannot launch aggressor %q, load point %d, repetition %d", beLauncher.Launcher.Name(), loadPoint, repetition)
					}
					cleanupStack = append(cleanupStack, func() {
						beHandle.Stop()
						beHandle.Clean()
					})
					if beLauncher.SnapSessionLauncher != nil {
						aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, beLauncher.Launcher.Name())
						if err != nil {
							cleanEnvironment(cleanupStack)
							return errors.Wrapf(err, "cannot launch aggressor snap session for %q, load point %d, repetition %d", beLauncher.Launcher.Name(), loadPoint, repetition)
						}
						cleanupStack = append(cleanupStack, func() {
							aggressorSnapHandle.Stop()
						})
					}

				}

				logrus.Debugf("Launching Load Generator with load point %d.", loadPoint)
				loadGeneratorHandle, err := loadGenerator.Load(loadPoint, sensitivity.LoadDurationFlag.Value())
				if err != nil {
					cleanEnvironment(cleanupStack)
					return errors.Wrapf(err, "Unable to start load generation in baseline, load point %d, repetition %d.", loadPoint, repetition)
				}
				loadGeneratorHandle.Wait(0)

				snapTags := fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s",
					phase.ExperimentKey, uuid.String(),
					phase.PhaseKey, phaseName,
					phase.RepetitionKey, repetition,
					// TODO: Remove that when completing SCE-376
					phase.LoadPointQPSKey, loadPoint,
					phase.AggressorNameKey, "",
				)

				snapHandle, err := snapSession.LaunchSession(loadGeneratorHandle, snapTags)
				if err != nil {
					cleanEnvironment(cleanupStack)
					return errors.Wrapf(err, "cannot launch mutilate Snap session in baseline, loadpoint %d, repetition %d", loadPoint, repetition)
				}
				cleanupStack = append(cleanupStack, func() {
					time.Sleep(5 * time.Second)
					snapHandle.Stop()
				})

				exitCode, err := loadGeneratorHandle.ExitCode()
				cleanEnvironment(cleanupStack)
				if exitCode != 0 {
					return errors.Errorf("executing Load Generator returned with exit code %d in baseline, load point %d, repetition %d", exitCode, loadPoint, repetition)
				}
			}
		}
	}

	return nil
}

func createDirs(experimentDirectory, phaseName string, repetition int) (string, error) {
	phaseDir := path.Join(experimentDirectory, phaseName, strconv.Itoa(repetition))
	err := os.MkdirAll(phaseDir, 0777)
	if err != nil {
		return "", errors.Wrapf(err, "could not create dir %q", phaseDir)
	}

	err = os.Chdir(phaseDir)
	if err != nil {
		return "", errors.Wrapf(err, "could not change to dir %q", phaseDir)
	}

	return phaseDir, nil
}

func cleanEnvironment(cleanupStack []func()) {
	for _, v := range cleanupStack {
		v()
	}
}
