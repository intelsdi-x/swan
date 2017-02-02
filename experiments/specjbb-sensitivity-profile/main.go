package main

import (
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/experiments/specjbb-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"gopkg.in/cheggaaa/pb.v1"
)

const (
	// ExperimentKey defines the key for Snap tag.
	ExperimentKey = "swan_experiment"
	// PhaseKey defines the key for Snap tag.
	PhaseKey = "swan_phase"
	// RepetitionKey defines the key for Snap tag.
	RepetitionKey = "swan_repetition"
	// LoadPointQPSKey defines the key for Snap tag.
	LoadPointQPSKey = "swan_loadpoint_qps"
	// AggressorNameKey defines the key for Snap tag.
	AggressorNameKey = "swan_aggressor_name"

	// See /usr/include/sysexits.h for refference regarding constants below

	// ExUsage reperense command line user error exit code
	ExUsage = 64
	// ExSoftware represents internal software error exit code
	ExSoftware = 70
	// ExIOErr represents input/output error exit code
	ExIOErr = 74
)

var (
	specjbbIP = conf.NewStringFlag(
		"specjbb_loadgenerator_ip",
		"IP of the SPECjbb Load Generator host",
		"127.0.0.1")
)

func main() {
	conf.SetAppName("specjbb-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
                     It executes workloads and triggers gathering of metrics like latency (SLI)`)
	err := conf.ParseFlags()
	if err != nil {
		return
	}
	logrus.SetLevel(conf.LogLevel())

	// Generate an experiment ID and start the metadata session.
	uuid, err := uuid.NewV4()
	if err != nil {
		logrus.Errorf("Cannot generate experiment ID: %q", err.Error())
		os.Exit(ExSoftware)
	}
	// Create metadata associated with experiment
	metadata := experiment.NewMetadata(uuid.String(), experiment.MetadataConfigFromFlags())
	err = metadata.Connect()
	if err != nil {
		logrus.Errorf("Cannot connect to metadata database %q", err.Error())
		os.Exit(ExSoftware)
	}

	logrus.Info("Starting Experiment ", conf.AppName(), " with uuid ", uuid.String())

	//By default print only UUID of the experiment and nothing more on the stdout
	fmt.Println(uuid.String())

	// Each experiment should have it's own directory to store logs and errors
	experimentDirectory, logFile, err := common.CreateExperimentDir(uuid.String())
	if err != nil {
		logrus.Errorf("IO error: %q", err.Error())
		os.Exit(ExIOErr)
	}

	// Setup logging set to both output and logFile.
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	// Validate preconditions: for SPECjbb we only check if CPU governor is set to performance.
	validate.CheckCPUPowerGovernor()

	// Prepare isolation for HighPriority job and default aggressors (L1 cache, Last Level Cache)
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Create executor for high priority job and for aggressors. Apply isolation to high priority task.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	if err != nil {
		return
	}
	// On exit performa deferred cleanup.
	defer cleanup()

	// Prepare session launchers (including Snap session if necessary) for aggressors.
	aggressorSessionLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	if err != nil {
		return
	}

	// Zero-value sensitivity.LauncherSessionPair represents baselining.
	aggressorSessionLaunchers = append([]sensitivity.LauncherSessionPair{sensitivity.LauncherSessionPair{}}, aggressorSessionLaunchers...)

	specjbbHost := specjbbIP.Value()
	// Create launcher for high priority task (in case of SPECjbb it is a backend).
	backendConfig := specjbb.DefaultSPECjbbBackendConfig()
	backendConfig.IP = specjbbHost
	specjbbBackendLauncher := specjbb.NewBackend(hpExecutor, backendConfig)

	// Prepare load generator for hp task (in case of the specjbb it is a controller with transaction injectors).
	specjbbLoadGenerator, err := common.PrepareSpecjbbLoadGenerator(specjbbHost)
	if err != nil {
		return
	}
	specjbbSnapSession, err := common.PrepareSnapSpecjbbSessionLauncher()
	if err != nil {
		return
	}
	specjbbLoadGeneratorSessionPair := sensitivity.NewMonitoredLoadGenerator(specjbbLoadGenerator, specjbbSnapSession)

	// Retrieve peak load from flags and overwrite it when required.
	load := sensitivity.PeakLoadFlag.Value()

	if load == sensitivity.RunTuningPhase {
		load, err = common.GetPeakLoad(specjbbBackendLauncher, specjbbLoadGeneratorSessionPair.LoadGenerator, sensitivity.SLOFlag.Value())
		if err != nil {
			logrus.Errorf("Cannot retrieve peak load (using tuning): %q", err.Error())
			os.Exit(ExSoftware)
		}
		logrus.Infof("Ran tuning and achieved load of %d", load)
	} else {
		logrus.Infof("Skipping Tunning phase, using peakload %d", load)
	}

	// Initialiaze progress bar when log level is error.
	var bar *pb.ProgressBar
	totalPhases := sensitivity.LoadPointsCountFlag.Value() * sensitivity.RepetitionsFlag.Value() * len(aggressorSessionLaunchers)
	if conf.LogLevel() == logrus.ErrorLevel {
		bar = pb.StartNew(totalPhases)
		bar.ShowCounters = false
		bar.ShowTimeLeft = true
		defer bar.Finish()
	}

	// We need to count fully executed aggressor loops to render progress bar correctly.
	var beIteration int

	// Read configuration.
	stopOnError := sensitivity.StopOnErrorFlag.Value()
	loadPoints := sensitivity.LoadPointsCountFlag.Value()
	repetitions := sensitivity.RepetitionsFlag.Value()
	loadDuration := sensitivity.LoadDurationFlag.Value()

	// Record metadata.
	records := map[string]string{
		"command_arguments": strings.Join(os.Args, ","),
		"experiment_name":   conf.AppName(),
		"peak_load":         strconv.Itoa(load),
		"load_points":       strconv.Itoa(loadPoints),
		"repetitions":       strconv.Itoa(repetitions),
		"load_duration":     loadDuration.String(),
	}
	// Store SWAN_ environment configuration.
	for _, env := range os.Environ() {
		fields := strings.SplitN(env, "=", 2)
		key, value := fields[0], fields[1]
		if strings.HasPrefix(key, "SWAN_") {
			records[key] = value
		}
	}

	err = metadata.RecordMap(records)
	if err != nil {
		logrus.Errorf("Cannot save metadata: %q", err.Error())
		os.Exit(ExSoftware)
	}

	// Iterate over aggressors
	for _, beLauncher := range aggressorSessionLaunchers {
		// For each aggressor iterate over defined loadpoints
		for loadPoint := 0; loadPoint < loadPoints; loadPoint++ {
			phaseQPS := int(int(load) / sensitivity.LoadPointsCountFlag.Value() * (loadPoint + 1))
			// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
			aggressorName := "Baselining"
			if beLauncher.Launcher != nil {
				aggressorName = beLauncher.Launcher.Name()
			}
			phaseName := fmt.Sprintf("Aggressor %s; load point %d;", aggressorName, loadPoint)
			// Repeat measurement to check if it is cosistent
			for repetition := 0; repetition < repetitions; repetition++ {
				// We need to collect all the TaskHandles created in order to cleanup after repetition finishes.
				var processes []executor.TaskHandle
				// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
				// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.
				executeRepetition := func() error {
					// Make progress bar to display current repetition.
					if conf.LogLevel() == logrus.ErrorLevel {
						completedPhases := beIteration * sensitivity.LoadPointsCountFlag.Value() * sensitivity.RepetitionsFlag.Value()
						prefix := fmt.Sprintf("[%02d / %02d] %s, repetition %d ", completedPhases+loadPoint+repetition+1, totalPhases, phaseName, repetition)
						bar.Prefix(prefix)
						// Changes to progress bar should be applied immediately
						bar.AlwaysUpdate = true
						bar.Update()
						bar.AlwaysUpdate = false
						defer bar.Add(1)
					}

					logrus.Infof("Starting %s repetition %d", phaseName, repetition)

					_, err := common.CreateRepetitionDir(experimentDirectory, phaseName, repetition)
					if err != nil {
						return errors.Wrapf(err, "cannot create repetition log directory in %s, repetition %d", phaseName, repetition)
					}

					// Launch specjbb backend (high priority job)
					hpHandle, err := specjbbBackendLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s repetition %d", phaseName, repetition)
					}
					processes = append(processes, hpHandle)

					// Launch specjbb Load Generator to populate data
					err = specjbbLoadGeneratorSessionPair.LoadGenerator.Populate()
					if err != nil {
						return errors.Wrapf(err, "cannot populate memcached in %s, repetition %d", phaseName, repetition)
					}

					snapTags := fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s",
						ExperimentKey, uuid.String(),
						PhaseKey, phaseName,
						RepetitionKey, repetition,
						LoadPointQPSKey, phaseQPS,
						AggressorNameKey, aggressorName,
					)

					// Launch aggressor task(s) when we are not in baseline.
					if beLauncher.Launcher != nil {
						beHandle, err := beLauncher.Launcher.Launch()
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor %q, in %s repetition %d", beLauncher.Launcher.Name(), phaseName, repetition)
						}
						processes = append(processes, beHandle)

						// In case of some aggressor we measure work done by them thus snaptel collector is needed.
						if beLauncher.SnapSessionLauncher != nil {
							logrus.Debugf("starting snap session: ")
							aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, beLauncher.Launcher.Name())
							if err != nil {
								return errors.Wrapf(err, "cannot launch aggressor snap session for %s, repetition %d", phaseName, repetition)
							}
							defer func() {
								aggressorSnapHandle.Stop()
							}()
						}

					}

					// After high priority job and aggressors are launched Load Generator may start it's job to stress HP
					logrus.Debugf("Launching Load Generator with load point %d", loadPoint)
					loadGeneratorHandle, err := specjbbLoadGeneratorSessionPair.LoadGenerator.Load(phaseQPS, loadDuration)
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s, repetition %d.", phaseName, repetition)
					}
					loadGeneratorHandle.Wait(0)

					// Grap results from Load Generator
					snapHandle, err := specjbbLoadGeneratorSessionPair.SnapSessionLauncher.LaunchSession(loadGeneratorHandle, snapTags)
					if err != nil {
						return errors.Wrapf(err, "cannot launch specjbb load generator Snap session in %s, repetition %d", phaseName, repetition)
					}
					defer func() {
						// It is ugly but there is no other way to make sure that data is written to Cassandra as of now.
						time.Sleep(5 * time.Second)
						snapHandle.Stop()
					}()

					exitCode, err := loadGeneratorHandle.ExitCode()
					if exitCode != 0 {
						return errors.Errorf("executing Load Generator returned with exit code %d in %s, repetition %d", exitCode, phaseName, repetition)
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
					logrus.Errorf("Experiment failed (%s, repetition %d): %q", phaseName, repetition, err.Error())
					if stopOnError {
						os.Exit(ExSoftware)
					}
				}
			} // repetition
		} // loadpoints
		beIteration++
	} // aggressors
}
