package main

import (
	"fmt"
	"io"
	"os"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
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
)

func main() {
	// Preparing application - setting name, help, aprsing flags etc.
	experimentStart := time.Now()
	conf.SetAppName("memcached-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)`)
	err := conf.ParseFlags()
	if err != nil {
		logrus.Errorf("Cannot parse flags: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(64)
	}
	logrus.SetLevel(conf.LogLevel())

	// Validate preconditions.
	validate.OS()

	// Create isolations.
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Create executors with cleanup function.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	if err != nil {
		logrus.Errorf("Cannot create executors: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(70)
	}
	defer cleanup()

	// Create BE workloads.
	beLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	if err != nil {
		logrus.Errorf("Cannot create BE tasks: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(70)
	}
	// Zero-value sensitivity.LauncherSessionPair represents baselining.
	beLaunchers = append([]sensitivity.LauncherSessionPair{sensitivity.LauncherSessionPair{}}, beLaunchers...)

	// Create HP workload.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := memcached.New(hpExecutor, memcachedConfig)

	// Load generator.
	loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	if err != nil {
		logrus.Errorf("Cannot create load generator: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(70)
	}

	snapSession, err := common.PrepareSnapMutilateSessionLauncher()
	if err != nil {
		logrus.Errorf("Cannot create snap session: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(70)
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		logrus.Errorf("Cannot generate experiment ID: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(70)
	}
	logrus.Info("Starting Experiment ", conf.AppName(), " with uuid ", uuid.String())
	fmt.Println(uuid.String())

	experimentDirectory, logFile, err := common.CreateExperimentDir(uuid.String())
	if err != nil {
		logrus.Errorf("IO error: %q", err.Error())
		// All the exit code values are based on /usr/include/sysexits.h
		os.Exit(74)
	}

	// Setup logging set to both output and logFile.
	logrus.SetLevel(conf.LogLevel())
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	// Retrieve peak load from flags and overwrite it when required.
	load := sensitivity.PeakLoadFlag.Value()
	if sensitivity.PeakLoadFlag.Value() == sensitivity.RunTuningPhase {
		load, err = common.GetPeakLoad(hpLauncher, loadGenerator, sensitivity.SLOFlag.Value())
		if err != nil {
			logrus.Errorf("Cannot retrieve peak load (using tuning): %q", err.Error())
			// All the exit code values are based on /usr/include/sysexits.h
			os.Exit(70)
		}
		logrus.Infof("Run tuning and achieved load of %d", load)
	} else {
		logrus.Infof("Skipping Tunning phase, using peakload %d", load)
	}

	// Initialiaze progress bar when log level is error.
	var bar *pb.ProgressBar
	totalPhases := sensitivity.LoadPointsCountFlag.Value() * sensitivity.RepetitionsFlag.Value() * len(beLaunchers)
	if conf.LogLevel() == logrus.ErrorLevel {
		bar = pb.StartNew(totalPhases)
		bar.ShowCounters = false
		bar.ShowTimeLeft = true
		defer bar.Finish()
	}

	// We need to count fully executed aggressor loops to render progress bar correctly.
	var beIteration int

	stopOnError := sensitivity.StopOnErrorFlag.Value()
	for _, beLauncher := range beLaunchers {
		for loadPoint := 0; loadPoint < sensitivity.LoadPointsCountFlag.Value(); loadPoint++ {
			// Calculate number of QPS in phase.
			phaseQPS := int(int(load) / sensitivity.LoadPointsCountFlag.Value() * (loadPoint + 1))
			// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
			aggressorName := "None"
			if beLauncher.Launcher != nil {
				aggressorName = beLauncher.Launcher.Name()
			}
			phaseName := fmt.Sprintf("Aggressor %s, load point %d", aggressorName, loadPoint)
			for repetition := 0; repetition < sensitivity.RepetitionsFlag.Value(); repetition++ {
				// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
				// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.
				err := func() error {
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

					// Launch BE tasks when we are not in baseline.
					if beLauncher.Launcher != nil {
						beHandle, err := beLauncher.Launcher.Launch()
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor %q, in %s repetition %d", beLauncher.Launcher.Name(), phaseName, repetition)
						}
						defer func() {
							beHandle.Stop()
							beHandle.Clean()
						}()
						// Majority of LauncherSessionPairs do not use Swan.
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

					logrus.Debugf("Launching Load Generator with load point %d", loadPoint)
					loadGeneratorHandle, err := loadGenerator.Load(phaseQPS, sensitivity.LoadDurationFlag.Value())
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s, repetition %d.", phaseName, repetition)
					}
					loadGeneratorHandle.Wait(0)

					snapTags := fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s",
						ExperimentKey, uuid.String(),
						PhaseKey, phaseName,
						RepetitionKey, repetition,
						LoadPointQPSKey, loadPoint,
						AggressorNameKey, aggressorName,
					)

					snapHandle, err := snapSession.LaunchSession(loadGeneratorHandle, snapTags)
					if err != nil {
						return errors.Wrapf(err, "cannot launch mutilate Snap session in %s, repetition %d", phaseName, repetition)
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
				}()
				if err != nil {
					logrus.Errorf("Experiment failed (%s, repetition %d): %q", phaseName, repetition, err.Error())
					if stopOnError {
						// All the exit code values are based on /usr/include/sysexits.h
						os.Exit(70)
					}
				}
			}
		}
		beIteration++
	}
	logrus.Infof("Ended experiment %s with uuid %s in %s", conf.AppName(), uuid.String(), time.Since(experimentStart).String())
}
