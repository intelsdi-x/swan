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
	// Zero-value sensitivity.LauncherSessionPair represents baselining.
	beLaunchers = append([]sensitivity.LauncherSessionPair{sensitivity.LauncherSessionPair{}}, beLaunchers...)

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

	experimentDirectory, logFile, err := createExperimentDir(uuid.String())
	if err != nil {
		return errors.Wrap(err, "could not create experiment directory")
	}

	// Setup logging set to both output and logFile.
	logrus.SetLevel(conf.LogLevel())
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	load := sensitivity.PeakLoadFlag.Value()
	if sensitivity.PeakLoadFlag.Value() == sensitivity.RunTuningPhase {
		load, err = getPeakLoad(hpLauncher, loadGenerator, sensitivity.SLOFlag.Value())
		if err != nil {
			return errors.Wrap(err, "tuning failed - cannot determine peak load")
		}
		logrus.Infof("Run tuning and achieved load of %d", load)
	} else {
		logrus.Infof("Skipping Tunning phase, using peakload %d", load)
	}

	var bar *pb.ProgressBar
	totalPhases := sensitivity.LoadPointsCountFlag.Value() * sensitivity.RepetitionsFlag.Value() * len(beLaunchers)
	if conf.LogLevel() == logrus.ErrorLevel {
		bar = pb.StartNew(totalPhases)
		bar.ShowCounters = false
		bar.ShowTimeLeft = true
	}

	var launcherIteration int
	for _, beLauncher := range beLaunchers {
		for loadPoint := 0; loadPoint < sensitivity.LoadPointsCountFlag.Value(); loadPoint++ {
			phaseQPS := int(int(load) / sensitivity.LoadPointsCountFlag.Value() * (loadPoint + 1))
			var phaseName string
			if beLauncher.Launcher != nil {
				phaseName = fmt.Sprintf("%s, load point %d", beLauncher.Launcher.Name(), loadPoint)
			} else {
				phaseName = fmt.Sprintf("Baseline, load point %d", loadPoint)
			}
			for repetition := 0; repetition < sensitivity.RepetitionsFlag.Value(); repetition++ {
				err := func() error {
					if conf.LogLevel() == logrus.ErrorLevel {
						completedPhases := launcherIteration * sensitivity.LoadPointsCountFlag.Value() * sensitivity.RepetitionsFlag.Value()
						prefix := fmt.Sprintf("[%02d / %02d] %s, repetition %d ", completedPhases+loadPoint+repetition+1, totalPhases, phaseName, repetition)
						bar.Prefix(prefix)
						// Changes to progress bar should be applied immediately
						bar.AlwaysUpdate = true
						bar.Update()
						bar.AlwaysUpdate = false
					}

					logrus.Infof("Starting %s repetition %d", phaseName, repetition)

					_, err := createRepetitionDir(experimentDirectory, phaseName, repetition)
					if err != nil {
						return errors.Wrapf(err, "cannot create repetition log directory in %s, repetition %d", phaseName, repetition)
					}

					hpHandle, err := hpLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s repetition %d", phaseName, repetition)
					}
					defer func() {
						hpHandle.Stop()
						hpHandle.Clean()
					}()

					err = loadGenerator.Populate()
					if err != nil {
						return errors.Wrapf(err, "cannot populate memcached in %s, repetition %d", phaseName, repetition)
					}

					if beLauncher.Launcher != nil {
						beHandle, err := beLauncher.Launcher.Launch()
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor %q, in %s repetition %d", beLauncher.Launcher.Name(), phaseName, repetition)
						}
						defer func() {
							beHandle.Stop()
							beHandle.Clean()
						}()
						if beLauncher.SnapSessionLauncher != nil {
							aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, beLauncher.Launcher.Name())
							if err != nil {
								return errors.Wrapf(err, "cannot launch aggressor snap session for %s, repetition %d", phaseName, repetition)
							}
							defer func() {
								aggressorSnapHandle.Stop()
							}()
						}

					}

					logrus.Debugf("Launching Load Generator with load point %d.", loadPoint)
					loadGeneratorHandle, err := loadGenerator.Load(phaseQPS, sensitivity.LoadDurationFlag.Value())
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s, repetition %d.", phaseName, repetition)
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
						return errors.Wrapf(err, "cannot launch mutilate Snap session in %s, repetition %d", phaseName, repetition)
					}
					defer func() {
						time.Sleep(5 * time.Second)
						snapHandle.Stop()
					}()

					exitCode, err := loadGeneratorHandle.ExitCode()
					if exitCode != 0 {
						return errors.Errorf("executing Load Generator returned with exit code %d in %s, repetition %d", exitCode, phaseName, repetition)
					}

					if conf.LogLevel() == logrus.ErrorLevel {
						bar.Add(1)
					}

					return nil
				}()
				if err != nil {
					return errors.Wrap(err, "Experiment failed")
				}
			}
		}
		launcherIteration++
	}

	return nil
}

func createRepetitionDir(experimentDirectory, phaseName string, repetition int) (string, error) {
	repetitionDir := path.Join(experimentDirectory, phaseName, strconv.Itoa(repetition))
	err := os.MkdirAll(repetitionDir, 0777)
	if err != nil {
		return "", errors.Wrapf(err, "could not create dir %q", repetitionDir)
	}

	err = os.Chdir(repetitionDir)
	if err != nil {
		return "", errors.Wrapf(err, "could not change to dir %q", repetitionDir)
	}

	return repetitionDir, nil
}

func createExperimentDir(uuid string) (experimentDirectory string, logFile *os.File, err error) {
	experimentDirectory = path.Join(os.TempDir(), conf.AppName(), uuid)
	err = os.MkdirAll(experimentDirectory, 0777)
	if err != nil {
		return "", &os.File{}, errors.Wrap(err, "cannot create experiment directory")
	}
	err = os.Chdir(experimentDirectory)
	os.Chdir(os.TempDir())
	if err != nil {
		return "", &os.File{}, errors.Wrap(err, "cannot chdir to experiment directory")
	}

	masterLogFilename := path.Join(experimentDirectory, "master.log")
	logFile, err = os.OpenFile(masterLogFilename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "could not open log file %q", masterLogFilename)
	}

	return experimentDirectory, logFile, nil
}

func getPeakLoad(hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator, slo int) (int, error) {
	prTask, err := hpLauncher.Launch()
	if err != nil {
		return 0, errors.Wrap(err, "cannot launch memcached")
	}
	defer func() {
		prTask.Stop()
		prTask.Clean()
	}()

	err = loadGenerator.Populate()
	if err != nil {
		return 0, errors.Wrap(err, "cannot populate memcached")
	}

	load, _, err := loadGenerator.Tune(slo)
	if err != nil {
		return 0, errors.Wrap(err, "tuning failed")
	}

	return int(load), nil
}
