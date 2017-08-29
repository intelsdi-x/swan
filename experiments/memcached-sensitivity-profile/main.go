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

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/metadata"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	_ "github.com/intelsdi-x/swan/pkg/utils/unshare"
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

	metaData, err := metadata.NewCassandra(uid, metadata.DefaultCassandraConfig())
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(metaData, experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Launch Kubernetes cluster.
	if experiment.ShouldLaunchKubernetesCluster() {
		handle, err := experiment.LaunchKubernetesCluster()
		errutil.CheckWithContext(err, "Could not launch Kubernetes cluster")
		defer handle.Stop()
	}

	tuningTags := make(map[string]interface{})
	tuningTags[experiment.ExperimentKey] = uid
	tuningTags[experiment.PhaseKey] = "tuning"

	factory := sensitivity.NewDefaultWorkloadFactory()

	hpLauncher, err := factory.BuildDefaultHighPriorityLauncher(sensitivity.Memcached, tuningTags)
	errutil.CheckWithContext(err, "cannot prepare memcached")

	// Load generator.
	loadGenerator, err := common.PrepareDefaultMutilateGenerator()
	errutil.CheckWithContext(err, "cannot prepare load generator")

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

	err = metaData.RecordMap(records, metadata.TypeEmpty)
	errutil.CheckWithContext(err, "cannot save metadata")

	bestEfforts := sensitivity.AggressorsFlag.Value()
	for _, bestEffortWorkloadName := range bestEfforts {
		for loadPoint := 0; loadPoint < loadPoints; loadPoint++ {
			// Calculate number of QPS in phase.
			phaseQPS := int(int(load) / sensitivity.LoadPointsCountFlag.Value() * (loadPoint + 1))

			for repetition := 0; repetition < repetitions; repetition++ {
				phaseName := fmt.Sprintf("Aggressor %s; load point %d; repetition %d", bestEffortWorkloadName, loadPoint, repetition)
				// We need to collect all the TaskHandles created in order to cleanup after repetition finishes.
				var processes []executor.TaskHandle
				// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
				// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.
				executeRepetition := func() error {
					logrus.Infof("Starting phase: %s", phaseName)

					snapTags := make(map[string]interface{})
					snapTags[experiment.ExperimentKey] = uid
					snapTags[experiment.PhaseKey] = phaseName
					snapTags[experiment.RepetitionKey] = repetition
					snapTags[experiment.LoadPointQPSKey] = phaseQPS
					snapTags[experiment.AggressorNameKey] = bestEffortWorkloadName

					err := experiment.CreateRepetitionDir(appName, uid, phaseName, repetition)
					if err != nil {
						return errors.Wrapf(err, "cannot create repetition log directory in phase %q", phaseName)
					}

					hpLauncher, err := factory.BuildDefaultHighPriorityLauncher(sensitivity.Memcached, snapTags)
					errutil.CheckWithContext(err, "cannot prepare memcached")
					hpHandle, err := hpLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s", phaseName)
					}
					processes = append(processes, hpHandle)

					err = loadGenerator.Populate()
					if err != nil {
						return errors.Wrapf(err, "cannot populate memcached in %s", phaseName)
					}

					beLauncher, err := factory.BuildDefaultBestEffortLauncher(bestEffortWorkloadName, snapTags)
					errutil.CheckWithContext(err, fmt.Sprintf("cannot prepare best effort workload %q", bestEffortWorkloadName))
					// Launch BE tasks when we are not in baseline.
					var beHandle executor.TaskHandle
					if beLauncher != nil {
						beHandle, err := beLauncher.Launch()
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor %q, in phase %q", beLauncher, phaseName)
						}
						processes = append(processes, beHandle)
					}

					logrus.Debugf("Launching Load Generator with load point %d", loadPoint)
					loadGeneratorHandle, err := loadGenerator.Load(phaseQPS, loadDuration)
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in phase %q", phaseName)
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
							return errors.Wrapf(err, "best effort task has failed in phase %q", phaseName)
						}
					}

					mutilateOutput, err := loadGeneratorHandle.StdoutFile()
					if err != nil {
						return errors.Wrapf(err, "cannot get mutilate stdout file")
					}
					defer mutilateOutput.Close()

					// Create snap session launcher
					mutilateSnapSession, err := mutilatesession.NewSessionLauncherDefault(
						mutilateOutput.Name(), snapTags)
					if err != nil {
						return errors.Wrapf(err, fmt.Sprintf("Cannot create Mutilate snap session during phase %q", phaseName))
					}

					snapHandle, err := mutilateSnapSession.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch mutilate Snap session in phase %s", phaseName)
					}
					defer func() {
						// It is ugly but there is no other way to make sure that data is written to Cassandra as of now.
						time.Sleep(5 * time.Second)
						snapHandle.Stop()
					}()

					exitCode, err := loadGeneratorHandle.ExitCode()
					if exitCode != 0 {
						return errors.Errorf("executing Load Generator returned with exit code %d in phase %q", exitCode, phaseName)
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
					logrus.Errorf("Experiment failed (%s): %q", phaseName, err.Error())
					if stopOnError {
						os.Exit(experiment.ExSoftware)
					}
				}
			}
		}
	}
	logrus.Infof("Experiment %s with uid %s has ended in %s", appName, uid, time.Since(experimentStart).String())
}
