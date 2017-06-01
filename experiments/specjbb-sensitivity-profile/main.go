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
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/experiments/specjbb-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	"github.com/pkg/errors"
)

var (
	specjbbTxICountFlag    = conf.NewIntFlag("specjbb_transaction_injectors_count", "Number of Transaction injectors run in one group", 1)
	specjbbWorkerCountFlag = conf.NewIntFlag(
		"specjbb_worker_count",
		"Number of fork join worker threads (defaults to number of logical threads)",
		runtime.NumCPU())
	appName = os.Args[0]
)

func main() {
	experimentStart := time.Now()
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	uid := uuid.New() // Initialize logger.
	logger.Initialize(appName, uid)
	// Create metadata associated with experiment
	metadata, err := experiment.NewMetadata(uid, experiment.DefaultMetadataConfig())
	errutil.Check(err)

	err = metadata.RecordRuntimeEnv(experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment details to Cassandra metadata database.")

	// Validate preconditions: for SPECjbb we only check if CPU governor is set to performance.
	validate.CheckCPUPowerGovernor()

	// Prepare isolation for HighPriority job and default aggressors (L1 cache, Last Level Cache)
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Create executor for high priority job and for aggressors. Apply isolation to high priority task.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	errutil.Check(err)

	// On exit performa deferred cleanup.
	defer func() {
		if cleanup != nil {
			err := cleanup()
			if err != nil {
				logrus.Errorf("Cannot clean the environment: %q", err)
			}
		}
	}()

	// Prepare session launchers (including Snap session if necessary) for aggressors.
	aggressorSessionLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	errutil.Check(err)

	specjbbControllerAddress := specjbb.ControllerAddress.Value()
	// Create launcher for high priority task (in case of SPECjbb it is a backend).
	backendConfig := specjbb.DefaultSPECjbbBackendConfig()
	backendConfig.JVMHeapMemoryGBs = 8
	backendConfig.ParallelGCThreads = 4
	backendConfig.WorkerCount = specjbbWorkerCountFlag.Value()
	backendConfig.ControllerAddress = specjbbControllerAddress
	specjbbBackendLauncher := specjbb.NewBackend(hpExecutor, backendConfig)

	// Prepare load generator for hp task (in case of the specjbb it is a controller with transaction injectors).
	specjbbLoadGenerator, err := common.PrepareSpecjbbLoadGenerator(specjbbControllerAddress, specjbbTxICountFlag.Value())
	errutil.Check(err)

	// Note: DefaultConfig shall set SnaptelAddress.
	specjbbSnapSession, err := specjbbsession.NewSessionLauncherDefault()
	errutil.Check(err)

	specjbbLoadGeneratorSessionPair := sensitivity.NewMonitoredLoadGenerator(specjbbLoadGenerator, specjbbSnapSession)

	// Retrieve peak load from flags and overwrite it when required.
	load := sensitivity.PeakLoadFlag.Value()

	if load == sensitivity.RunTuningPhase {
		load, err = experiment.GetPeakLoad(specjbbBackendLauncher, specjbbLoadGeneratorSessionPair.LoadGenerator, sensitivity.SLOFlag.Value())
		errutil.Check(err)
		logrus.Infof("Ran tuning and achieved load of %d", load)
	} else {
		logrus.Infof("Skipping tuning phase, using peakload %d", load)
	}

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

	errutil.Check(metadata.RecordMap(records))

	// Iterate over aggressors
	for _, beLauncher := range aggressorSessionLaunchers {
		// For each aggressor iterate over defined loadpoints
		for loadPoint := 0; loadPoint < loadPoints; loadPoint++ {
			phaseQPS := int(int(load) / sensitivity.LoadPointsCountFlag.Value() * (loadPoint + 1))
			// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
			aggressorName := "Baselining"
			if beLauncher.Launcher != nil {
				aggressorName = beLauncher.Launcher.String()
			}
			phaseName := fmt.Sprintf("Aggressor %s; load point %d;", aggressorName, loadPoint)
			// Repeat measurement to check if it is consistent
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

					// Launch specjbb backend (high priority job)
					hpHandle, err := specjbbBackendLauncher.Launch()
					if err != nil {
						return errors.Wrapf(err, "cannot launch memcached in %s repetition %d", phaseName, repetition)
					}
					processes = append(processes, hpHandle)

					snapTags := make(map[string]interface{})
					snapTags[experiment.ExperimentKey] = uid
					snapTags[experiment.PhaseKey] = phaseName
					snapTags[experiment.RepetitionKey] = repetition
					snapTags[experiment.LoadPointQPSKey] = phaseQPS
					snapTags[experiment.AggressorNameKey] = aggressorName

					var beHandle executor.TaskHandle
					// Launch aggressor task(s) when we are not in baseline.
					if beLauncher.Launcher != nil {
						beHandle, err = beLauncher.Launcher.Launch()
						if err != nil {
							return errors.Wrapf(err, "cannot launch aggressor %q, in %s repetition %d", beLauncher.Launcher, phaseName, repetition)
						}
						processes = append(processes, beHandle)

						// In case of some aggressor we measure work done by them thus snaptel collector is needed.
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

					// After high priority job and aggressors are launched Load Generator may start it's job to stress HP
					logrus.Debugf("Launching Load Generator with load point %d", loadPoint)
					loadGeneratorHandle, err := specjbbLoadGeneratorSessionPair.LoadGenerator.Load(phaseQPS, loadDuration)
					if err != nil {
						return errors.Wrapf(err, "Unable to start load generation in %s, repetition %d.", phaseName, repetition)
					}
					loadGeneratorHandle.Wait(0)

					if beHandle != nil {
						err = beHandle.Stop()
						if err != nil {
							return errors.Wrapf(err, "best effort task has failed in phase %s", phaseName)
						}
					}

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
				}

				// If any error was found then we should log details and terminate the experiment if stopOnError is set.
				err = errColl.GetErrIfAny()
				errutil.Check(err)
			} // repetition
		} // loadpoints
	} // aggressors
}
