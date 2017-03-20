package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/caffe"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
)

var (
	qpsFlag = conf.NewStringFlag("nice_qps", "Comma-separated list of QpS to iterate over", "375000")
	appName = os.Args[0]
)

func main() {
	// Preparing application - setting name, help, aprsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// This very experiment needs to be run on K8s
	if !sensitivity.RunOnKubernetesFlag.Value() {
		logrus.Errorf("The experiment HAS to be run on Kubernetes!")
		os.Exit(experiment.ExUsage)
	}

	// Generate an experiment ID and start the metadata session.
	uid, err := uuid.NewV4()
	errutil.CheckWithContext(err, "Cannot generate experiment ID")

	// Connect to metadata database
	metadata := experiment.NewMetadata(uid.String(), experiment.MetadataConfigFromFlags())
	err = metadata.Connect()
	errutil.CheckWithContext(err, "Cannot connect to metadata database")

	logrus.Info("Starting Experiment ", appName, " with uid ", uid.String())
	fmt.Println(uid.String())

	// Write configuration as metadata.
	err = metadata.RecordFlags()
	errutil.CheckWithContext(err, "Cannot save flags to metadata database")

	// Store SWAN_ environment configuration.
	err = metadata.RecordEnv(conf.EnvironmentPrefix)
	errutil.CheckWithContext(err, "Cannot save environment metadata")

	// Create experiment directory
	experimentDirectory, logFile, err := experiment.CreateExperimentDir(uid.String(), appName)
	errutil.CheckWithContext(err, "Cannot create experiment logs directory")

	// Setup logging set to both output and logFile.
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.100"})
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	// Validate preconditions.
	validate.OS()

	// Read configuration.
	stopOnError := sensitivity.StopOnErrorFlag.Value()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	// Read QpS flag and convert to integers
	qps := qpsFlag.Value()
	var qpsList []int
	for _, v := range strings.Split(qps, ",") {
		vInt, err := strconv.Atoi(strings.TrimSpace(v))
		errutil.CheckWithContext(err, fmt.Sprintf("Failed converting %s to integer", v))
		qpsList = append(qpsList, vInt)
	}

	// Record metadata.
	records := map[string]string{
		"command_arguments": strings.Join(os.Args, ","),
		"experiment_name":   appName,
		"qps_set":           qps,
		"repetitions":       "1",
		"load_duration":     loadDuration.String(),
	}
	err = metadata.RecordMap(records)
	errutil.CheckWithContext(err, "Cannot save metadata")

	// We need to count fully executed aggressor loops to render progress bar correctly.
	var beIteration, totalIteration int

	//for _, beLauncher := range beLaunchers {
	for _, aggressorName := range sensitivity.AggressorsFlag.Value() {
		for _, qps := range qpsList {
			for niceness := -20; niceness <= 19; niceness++ {
				l1Isolation, llcIsolation, hpIsolation := prepareIsolation(niceness)
				logrus.Debugf("HP isolation: %+v, BE isolation: %+v", hpIsolation, l1Isolation)

				// Create executors with cleanup function.
				hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
				errutil.CheckWithContext(err, "Cannot create executors")
				defer cleanup()

				// Create BE workloads.
				beLauncher := createLauncherSessionPair(aggressorName, l1Isolation, llcIsolation, beExecutorFactory)

				// Create HP workload.
				memcachedConfig := memcached.DefaultMemcachedConfig()
				hpLauncher := executor.ServiceLauncher{memcached.New(hpExecutor, memcachedConfig)}

				// Create load generator.
				loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
				errutil.CheckWithContext(err, "Cannot create load generator")

				// Create snap session launcher
				snapSession, err := mutilatesession.NewSessionLauncherDefault()
				errutil.CheckWithContext(err, "Cannot create snap session")

				var phaseName string
				// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
				aggressorName := fmt.Sprintf("None - %d QpS", qps)
				if beLauncher.Launcher != nil {
					aggressorName = fmt.Sprintf("%s - %d QpS", beLauncher.Launcher.Name(), qps)
				}

				phaseName = fmt.Sprintf("Aggressor %s - BE priority %d", aggressorName, niceness*-1)
				// We need to collect all the TaskHandles created in order to cleanup after repetition finishes.
				var processes []executor.TaskHandle

				// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
				// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.
				executeRepetition := func() error {
					// Start BE job (and its session if it exists)
					beHandle, err := beLauncher.Launcher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch aggressor %s in %s", beLauncher.Launcher.Name(), phaseName)
					}
					processes = append(processes, beHandle)

					// Majority of LauncherSessionPairs do not use Snap.
					if beLauncher.SnapSessionLauncher != nil {
						logrus.Debugf("starting snap session: ")
						aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, beLauncher.Launcher.Name())
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor snap session for %s", phaseName)
						}
						defer func() {
							aggressorSnapHandle.Stop()
						}()
					}

					logrus.Infof("Starting %s", phaseName)

					err = experiment.CreateRepetitionDir(experimentDirectory, phaseName, 0)
					if err != nil {
						return errors.Wrapf(err, "cannot create repetition log directory in %s", phaseName)
					}

					hpHandle, err := hpLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s", phaseName)
					}
					processes = append(processes, hpHandle)

					err = loadGenerator.Populate()
					if err != nil {
						return errors.Wrapf(err, "cannot populate memcached in %s", phaseName)
					}

					snapTags := fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s,%s:%d,%s:%d",
						experiment.ExperimentKey, uid.String(),
						experiment.PhaseKey, phaseName,
						experiment.RepetitionKey, 0,
						experiment.LoadPointQPSKey, niceness*-1,
						experiment.AggressorNameKey, aggressorName,
						"niceness", niceness,
						"qps", qps,
					)

					logrus.Debugf("Launching Load Generator with load point %d", niceness)
					loadGeneratorHandle, err := loadGenerator.Load(qps, loadDuration)
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s", phaseName)
					}
					loadGeneratorHandle.Wait(0)

					snapHandle, err := snapSession.LaunchSession(loadGeneratorHandle, snapTags)
					if err != nil {
						return errors.Wrapf(err, "cannot launch mutilate Snap session in %s", phaseName)
					}
					defer func() {
						// It is ugly but there is no other way to make sure that data is written to Cassandra as of now.
						time.Sleep(5 * time.Second)
						snapHandle.Stop()
					}()

					exitCode, err := loadGeneratorHandle.ExitCode()
					if exitCode != 0 {
						return errors.Errorf("executing Load Generator returned with exit code %d in %s", exitCode, phaseName)
					}

					return nil
				}
				// Call repetition function.
				err = executeRepetition()

				// Collecting all the errors that might have been encountered.
				errColl := &errcollection.ErrorCollection{}
				errColl.Add(err)
				for _, th := range processes {
					errColl.Add(th.Stop())
					errColl.Add(th.Clean())
				}

				// If any error was found then we should log details and terminate the experiment if stopOnError is set.
				err = errColl.GetErrIfAny()
				if err != nil {
					logrus.Errorf("Experiment failed (%s): %+v", phaseName, err)
					if stopOnError {
						os.Exit(experiment.ExSoftware)
					}
				}
				totalIteration++
			}
			beIteration++
		}
	}
	logrus.Infof("Ended experiment %s with uid %s in %s", appName, uid.String(), time.Since(experimentStart).String())
}

func prepareIsolation(niceness int) (isolation.Decorator, isolation.Decorator, isolation.Decorator) {
	systemTopology := topology.NewManualTopology(topology.HpSetsFlag.Value(), topology.BeSetsFlag.Value(), topology.BeL1SetsFlag.Value(), topology.HpCPUExclusiveFlag.Value(), topology.BeCPUExclusiveFlag.Value())
	logrus.Debugf("Topology: %+v", systemTopology)
	l1Isolation := isolation.Decorators{isolation.Nice{Niceness: niceness}, isolation.Numactl{PhyscpubindCPUs: systemTopology.BeL1CPUs}}
	llcIsolation := isolation.Decorators{isolation.Nice{Niceness: niceness}, isolation.Numactl{PhyscpubindCPUs: systemTopology.BeCPUs}}
	hpIsolation := isolation.Decorators{isolation.Nice{Niceness: niceness}, isolation.Numactl{PhyscpubindCPUs: systemTopology.HpCPUs}}

	return l1Isolation, llcIsolation, hpIsolation
}

func createLauncherSessionPair(aggressorName string, l1Isolation, llcIsolation isolation.Decorator, beExecutorFactory sensitivity.ExecutorFactoryFunc) (beLauncher sensitivity.LauncherSessionPair) {
	aggressorFactory := sensitivity.NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)
	aggressorPair, err := aggressorFactory.Create(aggressorName, beExecutorFactory)
	errutil.CheckWithContext(err, "Cannot create aggressor pair")

	switch aggressorName {
	case caffe.ID:
		caffeSession, err := caffeinferencesession.NewSessionLauncher(caffeinferencesession.DefaultConfig())
		errutil.CheckWithContext(err, "Cannot create Caffee session launcher")
		beLauncher = sensitivity.NewMonitoredLauncher(aggressorPair, caffeSession)
	case l1data.ID, l1instruction.ID, memoryBandwidth.ID, l3data.ID, stream.ID:
		beLauncher = sensitivity.NewLauncherWithoutSession(aggressorPair)
	default:
		logrus.Warnf("Unknown aggressor: %q", aggressorName)
	}

	return
}
