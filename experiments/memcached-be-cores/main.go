package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	dockerclient "github.com/fsouza/go-dockerclient"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"gopkg.in/cheggaaa/pb.v1"
)

var (
	qpsFlag        = conf.NewStringFlag("cpus_qps", "Comma-separated list of QpS to iterate over", "375000")
	numberOfBECPUs = conf.NewIntFlag("cpus_be", "Number of CPUs available to BE job", 8)
	appName        = os.Args[0]
)

func main() {
	// Preparing application - setting name, help, aprsing flags etc.
	experimentStart := time.Now()
	errorLevelEnabled := experiment.Configure()

	// This very experiment needs to be run on K8s
	if !sensitivity.RunOnKubernetesFlag.Value() {
		logrus.Errorf("The experiment HAS to be run on Kubernetes!")
		os.Exit(experiment.ExUsage)
	}

	// Generate an experiment ID and start the metadata session.
	uuid, err := uuid.NewV4()
	errutil.CheckWithContext(err, "Cannot generate experiment ID")

	// Connect to metadata database
	metadata := experiment.NewMetadata(uuid.String(), experiment.MetadataConfigFromFlags())
	err = metadata.Connect()
	errutil.CheckWithContext(err, "Cannot connect to metadata database")

	logrus.Info("Starting Experiment ", appName, " with uuid ", uuid.String())
	fmt.Println(uuid.String())

	// Write configuration as metadata.
	err = metadata.RecordFlags()
	errutil.CheckWithContext(err, "Cannot save flags to metadata database")

	// Store SWAN_ environment configuration.
	err = metadata.RecordEnv(conf.EnvironmentPrefix)
	errutil.CheckWithContext(err, "Cannot save environment metadata")

	// Create experiment directory
	experimentDirectory, logFile, err := experiment.CreateExperimentDir(uuid.String(), appName)
	errutil.CheckWithContext(err, "Cannot create experiment logs directory")

	// Setup logging set to both output and logFile.
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.100"})
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	// Validate preconditions.
	validate.OS()

	// Create isolations.stdout_file
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Get isolation.Numactl from isolation.Decorator and terminate experiment if it can't be done.
	beThreads, ok := llcIsolation.(isolation.Numactl)
	if !ok {
		logrus.Fatal("Unable to cast isolation.Decorator to isolation.Numactl")
	}

	// Create executors with cleanup function.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	errutil.CheckWithContext(err, "Cannot create executors")
	defer cleanup()

	// Create BE workloads.
	beLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	errutil.CheckWithContext(err, "Cannot create BE tasks")

	// Create HP workload.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := executor.ServiceLauncher{memcached.New(hpExecutor, memcachedConfig)}

	// Create load generator.
	loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	errutil.CheckWithContext(err, "Cannot create load generator")

	// Create snap session launcher
	snapSession, err := mutilatesession.NewSessionLauncherDefault()
	errutil.CheckWithContext(err, "Cannot create snap session")

	// Read configuration.
	stopOnError := sensitivity.StopOnErrorFlag.Value()
	beCPUsCount := numberOfBECPUs.Value()
	repetitions := sensitivity.RepetitionsFlag.Value()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	// Read QpS flag and convert to integers
	qps := qpsFlag.Value()
	var qpsList []int
	for _, v := range strings.Split(qps, ",") {
		vInt, err := strconv.Atoi(strings.TrimSpace(v))
		errutil.CheckWithContext(err, fmt.Sprintf("Failed converting %s to integer", v))
		qpsList = append(qpsList, vInt)
	}

	// Initialiaze progress bar when log level is error.
	var bar *pb.ProgressBar
	totalPhases := beCPUsCount - repetitions*len(beLaunchers)*len(qpsList)
	if errorLevelEnabled {
		bar = pb.StartNew(totalPhases)
		bar.ShowCounters = false
		bar.ShowTimeLeft = true
		defer bar.Finish()
	}

	// Record metadata.
	records := map[string]string{
		"command_arguments":            strings.Join(os.Args, ","),
		"experiment_name":              appName,
		"qps_set":                      qps,
		"number_of_cores_combinations": strconv.Itoa(beCPUsCount),
		"repetitions":                  strconv.Itoa(repetitions),
		"load_duration":                loadDuration.String(),
	}
	err = metadata.RecordMap(records)
	errutil.CheckWithContext(err, "Cannot save metadata")

	// We need to count fully executed aggressor loops to render progress bar correctly.
	var beIteration, totalIteration int

	for _, beLauncher := range beLaunchers {
		for _, qps := range qpsList {
			for numberOfCores := 1; numberOfCores < beCPUsCount; numberOfCores++ {
				var phaseName string
				// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
				aggressorName := fmt.Sprintf("None - %d QpS", qps)
				if beLauncher.Launcher != nil {
					aggressorName = fmt.Sprintf("%s - %d QpS", beLauncher.Launcher.Name(), qps)
				}

				phaseName = fmt.Sprintf("Name of phase unavailable - %d QpS", qps)
				// We need to collect all the TaskHandles created in order to cleanup after repetition finishes.
				var processes []executor.TaskHandle
				// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
				// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.

				executeRepetition := func() error {
					// Calculate cores that BE job should really use. We want BE job to start with all the cores that it can use and limit them later.
					logrus.Debugf("CPUs from isolation.Numactl: %v", beThreads.PhyscpubindCPUs)
					allCores := isolation.NewIntSet(beThreads.PhyscpubindCPUs...)
					logrus.Debugf("IntSet with all BE cores: %v", allCores)
					cores, err := allCores.Take(beCPUsCount - numberOfCores)
					coresRange := cores.AsRangeString()
					logrus.Debugf("Substracted %d cores and got: %v", numberOfCores, coresRange)
					if err != nil {
						return errors.Wrapf(err, "unable to substract cores for load point %d QpS %d", numberOfCores, qps)
					}
					phaseName = fmt.Sprintf("Aggressor %s - %d cores;", aggressorName, len(cores.AsSlice()))

					// Start BE job (and its session if it exists)
					beHandle, err := beLauncher.Launcher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch aggressor %q, in %s QpS %d", beLauncher.Launcher.Name(), phaseName, qps)
					}
					processes = append(processes, beHandle)

					// Majority of LauncherSessionPairs do not use Snap.
					if beLauncher.SnapSessionLauncher != nil {
						logrus.Debugf("starting snap session: ")
						aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, beLauncher.Launcher.Name())
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor snap session for %s, QpS %d", phaseName, qps)
						}
						defer func() {
							aggressorSnapHandle.Stop()
						}()
					}

					// Set cores for BE container
					metadata.Record("docker_cpus_"+phaseName, coresRange)
					err = setContainerCores(phaseName, coresRange, qps)
					if err != nil {
						return err
					}

					// Make progress bar to display current repetition.
					if errorLevelEnabled {
						prefix := fmt.Sprintf("[%02d / %02d] %s", totalIteration+1, totalPhases, phaseName)
						bar.Prefix(prefix)
						// Changes to progress bar should be applied immediately
						bar.AlwaysUpdate = true
						bar.Update()
						bar.AlwaysUpdate = false
						defer bar.Add(1)
					}

					logrus.Infof("Starting %s", phaseName, qps)

					err = experiment.CreateRepetitionDir(experimentDirectory, phaseName, 0)
					if err != nil {
						return errors.Wrapf(err, "cannot create repetition log directory in %s, QpS %d", phaseName, qps)
					}

					hpHandle, err := hpLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s QpS %d", phaseName, qps)
					}
					processes = append(processes, hpHandle)

					err = loadGenerator.Populate()
					if err != nil {
						return errors.Wrapf(err, "cannot populate memcached in %s, QpS %d", phaseName, qps)
					}

					snapTags := fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s,%s:%s,%s:%d,%s:%d",
						experiment.ExperimentKey, uuid.String(),
						experiment.PhaseKey, phaseName,
						experiment.RepetitionKey, 0,
						experiment.LoadPointQPSKey, beCPUsCount-numberOfCores,
						experiment.AggressorNameKey, aggressorName,
						"cores", strings.Replace(coresRange, ",", ".", -1),
						"number_of_cores", beCPUsCount-numberOfCores,
						"qps", qps,
					)

					logrus.Debugf("Launching Load Generator with load point %d", numberOfCores)
					loadGeneratorHandle, err := loadGenerator.Load(qps, loadDuration)
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s, QpS %d.", phaseName, qps)
					}
					loadGeneratorHandle.Wait(0)

					snapHandle, err := snapSession.LaunchSession(loadGeneratorHandle, snapTags)
					if err != nil {
						return errors.Wrapf(err, "cannot launch mutilate Snap session in %s, QpS %d", phaseName, qps)
					}
					defer func() {
						// It is ugly but there is no other way to make sure that data is written to Cassandra as of now.
						time.Sleep(5 * time.Second)
						snapHandle.Stop()
					}()

					exitCode, err := loadGeneratorHandle.ExitCode()
					if exitCode != 0 {
						return errors.Errorf("executing Load Generator returned with exit code %d in %s, QpS %d", exitCode, phaseName, qps)
					}

					return nil
				}
				// Call repetition function.
				err := executeRepetition()

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
					logrus.Errorf("Experiment failed (%s, QpS %d): %+v", phaseName, qps, err)
					if stopOnError {
						os.Exit(experiment.ExSoftware)
					}
				}
				totalIteration++
			}
			beIteration++
		}
	}
	logrus.Infof("Ended experiment %s with uuid %s in %s", appName, uuid.String(), time.Since(experimentStart).String())
}

func setContainerCores(phaseName, coresRange string, qps int) error {
	docker, err := dockerclient.NewClient("unix:///var/run/docker.sock")
	if err != nil {
		return errors.Wrapf(err, "unable to create Docker client for %s QpS %d", phaseName, qps)
	}
	containers, err := docker.ListContainers(dockerclient.ListContainersOptions{})
	if err != nil {
		return errors.Wrapf(err, "unable to retrive list of containers for %s QpS %d", phaseName, qps)
	}
	err = docker.UpdateContainer(containers[0].ID, dockerclient.UpdateContainerOptions{
		CpusetCpus: coresRange,
	})
	if err != nil {
		return errors.Wrapf(err, "unable to update container %s cpuset to %s %s QpS %d", containers[0].ID, coresRange, phaseName, qps)
	}

	return nil
}
