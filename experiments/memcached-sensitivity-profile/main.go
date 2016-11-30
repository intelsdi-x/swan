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

	// TODO(bp): Remove these below when completing SCE-376

	// LoadPointQPSKey defines the key for Snap tag.
	LoadPointQPSKey = "swan_loadpoint_qps"
	// AggressorNameKey defines the key for Snap tag.
	AggressorNameKey = "swan_aggressor_name"
)

func main() {
	conf.SetAppName("memcached-sensitivity-profile")
	conf.SetHelp(`Sensitivity experiment runs different measurements to test the performance of co-located workloads on a single node.
It executes workloads and triggers gathering of certain metrics like latency (SLI) and the achieved number of Request per Second (QPS/RPS)`)
	err := conf.ParseFlags()
	if err != nil {
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
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
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
	}
	defer cleanup()

	// BE workloads.
	beLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	if err != nil {
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
	}
	// Zero-value sensitivity.LauncherSessionPair represents baselining.
	beLaunchers = append([]sensitivity.LauncherSessionPair{sensitivity.LauncherSessionPair{}}, beLaunchers...)

	// HP workload.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := memcached.New(hpExecutor, memcachedConfig)

	// Load generator.
	loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	if err != nil {
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
	}

	snapSession, err := common.PrepareSnapMutilateSessionLauncher()
	if err != nil {
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
	}

	logrus.Info("Starting Experiment ", conf.AppName(), " with uuid ", uuid.String())
	fmt.Println(uuid.String())

	experimentDirectory, logFile, err := common.CreateExperimentDir(uuid.String())
	if err != nil {
		logrus.Errorf("Error occured: %q", err.Error())
		os.Exit(1)
	}

	// Setup logging set to both output and logFile.
	logrus.SetLevel(conf.LogLevel())
	logrus.SetFormatter(new(logrus.TextFormatter))
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	load := sensitivity.PeakLoadFlag.Value()
	if sensitivity.PeakLoadFlag.Value() == sensitivity.RunTuningPhase {
		load, err = common.GetPeakLoad(hpLauncher, loadGenerator, sensitivity.SLOFlag.Value())
		if err != nil {
			logrus.Errorf("Error occured: %q", err.Error())
			os.Exit(1)
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
	stopOnError := sensitivity.StopOnErrorFlag.Value()
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
						ExperimentKey, uuid.String(),
						PhaseKey, phaseName,
						RepetitionKey, repetition,
						// TODO: Remove that when completing SCE-376
						LoadPointQPSKey, loadPoint,
						AggressorNameKey, "",
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
					logrus.Errorf("Experiment failed (%s, repetition %d): %q", phaseName, repetition, err.Error())
					if stopOnError {
						os.Exit(1)
					}
				}
			}
		}
		launcherIteration++
	}
}
