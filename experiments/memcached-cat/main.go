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
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/caffe"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/pkg/errors"
)

var (
	qpsFlag                  = conf.NewStringFlag("cat_qps", "Comma-separated list of QpS to iterate over", "375000")
	maxCacheWaysToAssignFlag = conf.NewIntFlag("cat_max_cache_ways", "Mask representing maximum number of cache ways to assign to a job. It is assumed that cat_max_cache_ways and cat_min_cache_ways sum to number of all cache ways available.", 11)
	minCacheWaysToAssignFlag = conf.NewIntFlag("cat_min_cache_ways", "Mask representing minumim number of cache ways to assing to a job. It is assumed that cat_max_cache_ways and cat_min_cache_ways sum to number of all cache ways available.", 1)
	minNumberOfBECPUsFlag    = conf.NewIntFlag("cat_min_be_cpus", "Minimum number of CPUs available to BE job.", 1)
	maxNumberOfBECPUsFlag    = conf.NewIntFlag("cat_max_be_cpus", "Maximum number of CPUs available to BE job. If set to zero then all availabe cores will be used (taking isolation defined into consideration).", 0)
	includeBaselinePhaseFlag = conf.NewBoolFlag("baseline", "Run baseline phase (without aggressors)", true)
	appName                  = os.Args[0]
)

func main() {
	// Preparing application - setting name, help, parsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// This very experiment needs to be run on K8s
	if !sensitivity.RunOnKubernetesFlag.Value() {
		logrus.Errorf("The experiment HAS to be run on Kubernetes!")
		os.Exit(experiment.ExUsage)
	}

	// Generate an experiment ID and start the metadata session.
	uid := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, uid)

	// Connect to metadata database
	metadata, err := experiment.NewMetadata(uid, experiment.MetadataConfigFromFlags())
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Read configuration.
	stopOnError := sensitivity.StopOnErrorFlag.Value()
	maxCacheWaysToAssign := uint64(maxCacheWaysToAssignFlag.Value())
	minCacheWaysToAssign := uint64(minCacheWaysToAssignFlag.Value())
	loadDuration := sensitivity.LoadDurationFlag.Value()
	minBECPUsCount := minNumberOfBECPUsFlag.Value()
	maxBECPUsCount := maxNumberOfBECPUsFlag.Value()
	// Read QpS flag and convert to integers
	qps := qpsFlag.Value()
	var qpsList []int
	for _, v := range strings.Split(qps, ",") {
		vInt, err := strconv.Atoi(strings.TrimSpace(v))
		errutil.CheckWithContext(err, fmt.Sprintf("Failed converting %s to integer", v))
		qpsList = append(qpsList, vInt)
	}

	// Discover local CPU topology.
	beCores := topology.BeRangeFlag.Value()
	logrus.Debugf("BE CPU range: %+v", beCores)
	// If maximum number of cores is not provided - use all.
	if maxBECPUsCount == 0 {
		maxBECPUsCount = len(beCores)
	}

	// Record metadata.
	records := map[string]string{
		"command_arguments":            strings.Join(os.Args, ","),
		"experiment_name":              appName,
		"qps_set":                      qps,
		"repetitions":                  "1",
		"load_duration":                loadDuration.String(),
		"number_of_cores_combinations": strconv.Itoa(maxBECPUsCount - minBECPUsCount + 1),
		"max_be_cache_mask":            strconv.FormatUint(maxCacheWaysToAssign, 10),
		"min_be_cache_mask":            strconv.FormatUint(minCacheWaysToAssign, 10),
		"max_be_cpu_count":             strconv.Itoa(maxBECPUsCount),
		"min_be_cpu_count":             strconv.Itoa(minBECPUsCount),
	}
	err = metadata.RecordMap(records)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")
	logrus.Debugf("IntSet with all BE cores: %v", beCores)

	//We do not need to start Kubernetes on each repetition.
	cleanup, err := sensitivity.LaunchKubernetesCluster()
	errutil.CheckWithContext(err, "Cannot launch Kubernetes cluster")
	defer func() {
		if cleanup != nil {
			err := cleanup()
			if err != nil {
				logrus.Errorf("Kubernetes cleanup failed: %q", err)
			}
		}
	}()

	// Include baseline phase if necessary.
	aggressors := sensitivity.AggressorsFlag.Value()
	if includeBaselinePhaseFlag.Value() {
		aggressors = append(aggressors, "")
	}

	// We need to calculate mask for all cache ways to be able to calculate non-overlapping cache partitions.
	wholeCacheMask := 1<<(maxCacheWaysToAssign+minCacheWaysToAssign) - 1
	for _, aggressorName := range aggressors {
		logrus.Debugf("starting aggressor: %s", aggressorName)
		for _, qps := range qpsList {
			logrus.Debugf("starting QPS: %d", qps)
			// Initialize counters.
			var beIteration, totalIteration int
			for BECPUsCount := maxBECPUsCount; BECPUsCount >= minBECPUsCount; BECPUsCount-- {
				logrus.Debugf("staring cores: %d with limit of %d", BECPUsCount, minBECPUsCount)
				// Chose CPUs to be used
				cores, err := beCores.Take(BECPUsCount)
				errutil.CheckWithContext(err, fmt.Sprintf("unable to substract cores for aggressor %q, number of cores %d, QpS %d", aggressorName, BECPUsCount, qps))
				coresRange := cores.AsRangeString()
				logrus.Debugf("Substracted %d cores and got: %v", BECPUsCount, coresRange)

				for beCacheWays := maxCacheWaysToAssign; beCacheWays >= minCacheWaysToAssign; beCacheWays-- {
					// Calculate BE and HP cache masks
					beCacheMask := 1<<beCacheWays - 1
					hpCacheMask := wholeCacheMask &^ beCacheMask

					logrus.Debugf("Current L3 HP mask: %d, %b", hpCacheMask, hpCacheMask)
					logrus.Debugf("Current L3 BE mask: %d, %b", beCacheMask, beCacheMask)
					l1Isolation := isolation.Rdtset{Mask: beCacheMask, CPURange: topology.BeL1RangeFlag.Value().AsRangeString()}
					llcIsolation := isolation.Rdtset{Mask: beCacheMask, CPURange: coresRange}
					hpIsolation := isolation.Rdtset{Mask: hpCacheMask, CPURange: topology.HpRangeFlag.Value().AsRangeString()}
					logrus.Debugf("HP isolation: %+v, BE isolation: %+v", hpIsolation, llcIsolation)

					// Create executors with cleanup function.
					hpExecutor, err := sensitivity.CreateKubernetesHpExecutor(hpIsolation)
					errutil.CheckWithContext(err, "Cannot create executors")
					beExecutorFactory := sensitivity.DefaultKubernetesBEExecutorFactory

					// Clean any RDT assignments from previoues phases.
					pqosOutput, err := isolation.CleanRDTAssingments()
					logrus.Debugf("pqos -R has been run and produced following output: %q", pqosOutput)
					errutil.CheckWithContext(err, "pqos -R failed")

					// Create BE workloads.
					var beLauncher sensitivity.LauncherSessionPair
					if aggressorName != "" {
						beLauncher = createLauncherSessionPair(aggressorName, l1Isolation, llcIsolation, beExecutorFactory)
					}

					// Create HP workload.
					memcachedConfig := memcached.DefaultMemcachedConfig()
					hpLauncher := executor.ServiceLauncher{Launcher: memcached.New(hpExecutor, memcachedConfig)}

					// Create load generator.
					loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
					errutil.CheckWithContext(err, "Cannot create Mutilate load generator")

					// Create snap session launcher
					mutilateSnapSession, err := mutilatesession.NewSessionLauncherDefault()
					errutil.CheckWithContext(err, "Cannot create Mutilate snap session")

					// Generate name of the phase (taking zero-value LauncherSessionPair aka baseline into consideration).
					aggressorName := fmt.Sprintf("None - %d QpS", qps)
					if beLauncher.Launcher != nil {
						aggressorName = fmt.Sprintf("%s - %d QpS - %d cores", beLauncher.Launcher.Name(), qps, BECPUsCount)
					}

					phaseName := fmt.Sprintf("Aggressor %s - BE LLC %b", aggressorName, beCacheMask)
					// We need to collect all the TaskHandles created in order to cleanup after repetition finishes.
					var processes []executor.TaskHandle

					// Using a closure allows us to defer cleanup functions. Otherwise handling cleanup might get much more complicated.
					// This is the easiest and most golangish way. Deferring cleanup in case of errors to main() termination could cause panics.
					executeRepetition := func() error {
						// Building snap workload tags.
						snapTags := make(map[string]interface{})
						snapTags[experiment.ExperimentKey] = uid
						snapTags[experiment.PhaseKey] = phaseName
						snapTags[experiment.RepetitionKey] = 0
						snapTags[experiment.LoadPointQPSKey] = beCacheMask
						snapTags[experiment.AggressorNameKey] = aggressorName
						snapTags["be_l3_cache_size"] = beCacheMask
						snapTags["hp_l3_cache_size"] = hpCacheMask
						snapTags["qps"] = qps
						snapTags["number_of_cores"] = BECPUsCount

						logrus.Infof("Starting %s", phaseName)

						err = experiment.CreateRepetitionDir(appName, uid, phaseName, 0)
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

						var beHandle executor.TaskHandle
						// Start BE job (and its session if it exists)
						if beLauncher.Launcher != nil {
							beHandle, err = beLauncher.Launcher.Launch()
							if err != nil {
								return errors.Wrapf(err, "cannot launch aggressor %s in %s", beLauncher.Launcher.Name(), phaseName)
							}
							processes = append(processes, beHandle)
						}

						// Majority of LauncherSessionPairs do not use Snap.
						if beLauncher.SnapSessionLauncher != nil {
							logrus.Debugf("starting snap session: ")
							aggressorSnapHandle, err := beLauncher.SnapSessionLauncher.LaunchSession(beHandle, snapTags)
							if err != nil {
								return errors.Wrapf(err, "cannot launch aggressor snap session for %s", phaseName)
							}
							defer func() {
								aggressorSnapHandle.Stop()
							}()
						}

						logrus.Debugf("Launching Load Generator with BE cache mask: %b and HP cache mask: %b", beCacheMask, hpCacheMask)
						loadGeneratorHandle, err := loadGenerator.Load(qps, loadDuration)
						if err != nil {
							return errors.Wrapf(err, "Unable to start load generation in %s", phaseName)
						}
						if !loadGeneratorHandle.Wait(sensitivity.LoadGeneratorWaitTimeoutFlag.Value()) {
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
								return errors.Wrapf(err, "best effort task has failed in phase %s", phaseName)
							}
						}

						snapHandle, err := mutilateSnapSession.LaunchSession(loadGeneratorHandle, snapTags)
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
			}
			beIteration++
		}
	}
	logrus.Infof("Ended experiment %s with uid %s in %s", appName, uid, time.Since(experimentStart).String())
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
	case l1data.ID, l1instruction.ID, memoryBandwidth.ID, l3.ID, stream.ID:
		beLauncher = sensitivity.NewLauncherWithoutSession(aggressorPair)
	default:
		logrus.Fatalf("Unknown aggressor: %q", aggressorName)
	}

	return
}
