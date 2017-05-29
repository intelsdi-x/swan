// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	_ "github.com/intelsdi-x/swan/pkg/utils/unshare"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/pkg/errors"
)

var (
	appName = os.Args[0]
)

func main() {
	// Preparing application - setting name, help, aprsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	uid := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, uid)

	metadata, err := experiment.NewMetadata(uid, experiment.DefaultMetadataConfig())
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Create isolations.
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Create executors with cleanup function.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	errutil.CheckWithContext(err, "cannot prepare executors")
	defer func() {
		if cleanup != nil {
			err := cleanup()
			if err == nil {
				logrus.Errorf("Cannot clean the environment: %q", err)
			}
		}
	}()

	// Create BE workloads.
	beLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	errutil.CheckWithContext(err, "cannot prepare aggressors")

	// Create HP workload.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := executor.ServiceLauncher{Launcher: memcached.New(hpExecutor, memcachedConfig)}

	// Load generator.
	loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	errutil.CheckWithContext(err, "cannot prepare load generator")

	snapSession, err := mutilatesession.NewSessionLauncherDefault()
	errutil.CheckWithContext(err, "cannot create mutilate snap session")

	// Retrieve peak load from flags and overwrite it when required.
	load := sensitivity.PeakLoadFlag.Value()
	if load == sensitivity.RunTuningPhase {
		logrus.Info("Tuning phase...")
		load, err = experiment.GetPeakLoad(hpLauncher, loadGenerator, sensitivity.SLOFlag.Value())
		errutil.CheckWithContext(err, "cannot retrieve peak load during tuning")
		logrus.Infof("Ran tuning and achieved load of %d", load)
	} else {
		logrus.Infof("Skipping tuning phase, using peakload %d", load)
	}

	// Read configuration.
	stopOnError := sensitivity.StopOnErrorFlag.Value()
	loadPoints := sensitivity.LoadPointsCountFlag.Value()
	repetitions := sensitivity.RepetitionsFlag.Value()
	loadDuration := sensitivity.LoadDurationFlag.Value()

	// Record metadata.
	records := map[string]string{
		"command_arguments": strings.Join(os.Args, ","),
		"experiment_name":   appName,
		"peak_load":         strconv.Itoa(load),
		"load_points":       strconv.Itoa(loadPoints),
		"repetitions":       strconv.Itoa(repetitions),
		"load_duration":     loadDuration.String(),
	}

	err = metadata.RecordMap(records)
	errutil.CheckWithContext(err, "cannot save metadata")

	for _, beLauncher := range beLaunchers {
		for loadPoint := 0; loadPoint < loadPoints; loadPoint++ {
			// Calculate number of QPS in phase.
			phaseQPS := int(int(load) / sensitivity.LoadPointsCountFlag.Value() * (loadPoint + 1))
			// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
			aggressorName := sensitivity.NoneAggressorID
			if beLauncher.Launcher != nil {
				aggressorName = beLauncher.Launcher.String()
			}
			phaseName := fmt.Sprintf("Aggressor %s; load point %d;", aggressorName, loadPoint)
			for repetition := 0; repetition < repetitions; repetition++ {
				// We need to collect all the TaskHandles created in order to cleanup after repetition finishes.
				var processes []executor.TaskHandle
				// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
				// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.
				executeRepetition := func() error {
					logrus.Infof("Starting %s repetition %d", phaseName, repetition)

					err := experiment.CreateRepetitionDir(appName, uid, phaseName, repetition)
					if err != nil {
						return errors.Wrapf(err, "cannot create repetition log directory in %s, repetition %d", phaseName, repetition)
					}

					hpHandle, err := hpLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s repetition %d", phaseName, repetition)
					}
					processes = append(processes, hpHandle)

					err = loadGenerator.Populate()
					if err != nil {
						return errors.Wrapf(err, "cannot populate memcached in %s, repetition %d", phaseName, repetition)
					}

					snapTags := make(map[string]interface{})
					snapTags[experiment.ExperimentKey] = uid
					snapTags[experiment.PhaseKey] = phaseName
					snapTags[experiment.RepetitionKey] = repetition
					snapTags[experiment.LoadPointQPSKey] = phaseQPS
					snapTags[experiment.AggressorNameKey] = aggressorName

					// Launch BE tasks when we are not in baseline.
					var beHandle executor.TaskHandle
					if beLauncher.Launcher != nil {
						beHandle, err := beLauncher.Launcher.Launch()
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor %q, in %s repetition %d", beLauncher.Launcher, phaseName, repetition)
						}
						processes = append(processes, beHandle)

						// Majority of LauncherSessionPairs do not use Snap.
						if beLauncher.SnapSessionLauncher != nil {
							logrus.Debugf("starting snap session: ")
							aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, snapTags)
							if err != nil {
								return errors.Wrapf(err, "cannot launch aggressor snap session for %s, repetition %d", phaseName, repetition)
							}
							defer func() {
								aggressorSnapHandle.Stop()
							}()
						}

					}

					logrus.Debugf("Launching Load Generator with load point %d", loadPoint)
					loadGeneratorHandle, err := loadGenerator.Load(phaseQPS, loadDuration)
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s, repetition %d.", phaseName, repetition)
					}

					mutilateTerminated, err := loadGeneratorHandle.Wait(sensitivity.LoadGeneratorWaitTimeoutFlag.Value())
					if err != nil {
						logrus.Errorf("Mutilate cluster failed: %q", err)
						return errors.Wrap(err, "mutilate cluster failed")
					}
					if !mutilateTerminated {
						logrus.Warn("Mutilate cluster failed to stop on its own. Attempting to stop...")
						err := loadGeneratorHandle.Stop()
						if err != nil {
							logrus.Errorf("Stopping mutilate cluster errored: %q", err)
							return errors.Wrap(err, "stopping mutilate cluster errored")
						}
					}

					if beHandle != nil {
						err = beHandle.Stop()
						if err != nil {
							return errors.Wrapf(err, "best effort task has failed in phase %s, repetition %d", phaseName, repetition)
						}
					}

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
				}
				// Call repetition function.
				err := executeRepetition()

				// Collecting all the errors that might have been encountered.
				errColl := &errcollection.ErrorCollection{}
				errColl.Add(err)
				for _, th := range processes {
					errColl.Add(th.Stop())
				}

				// If any error was found then we should log details and terminate the experiment if stopOnError is set.
				err = errColl.GetErrIfAny()
				if err != nil {
					logrus.Errorf("Experiment failed (%s, repetition %d): %q", phaseName, repetition, err.Error())
					if stopOnError {
						os.Exit(experiment.ExSoftware)
					}
				}
			}
		}
	}
	logrus.Infof("Ended experiment %s with uid %s in %s", appName, uid, time.Since(experimentStart).String())
}
