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
	"strconv"
	"strings"
	"time"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/isolation/topo"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/use"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	_ "github.com/intelsdi-x/swan/pkg/utils/unshare"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
)

var (
	appName             = os.Args[0]
	useCorePinningFlag  = conf.NewBoolFlag("use_core_pinning", "Enables core pinning of memcached threads", false)
	maxThreadsFlag      = conf.NewIntFlag("max_threads", "Scale memcached up to cores (default to number of physical cores).", 0)
	useUSECollectorFlag = conf.NewBoolFlag("use_USE_collector", "Collects USE (Utilization, Saturation, Errors) metrics.", false)
)

func main() {
	experimentStart := time.Now()

	// Preparing application - setting name, help, parsing flags etc.
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	uid := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, uid)

	// connect to metadata database
	metadata, err := experiment.NewMetadata(uid, experiment.DefaultMetadataConfig())
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Read configuration.
	loadDuration := sensitivity.LoadDurationFlag.Value()
	loadPoints := sensitivity.LoadPointsCountFlag.Value()
	useCorePinning := useCorePinningFlag.Value()
	peakLoad := sensitivity.PeakLoadFlag.Value()
	if peakLoad == 0 {
		logrus.Fatalf("peak load have to be != 0!")
	}

	// Record metadata.
	records := map[string]string{
		"command_arguments": strings.Join(os.Args, ","),
		"experiment_name":   appName,
		"repetitions":       "1",
		"load_duration":     loadDuration.String(),
		"load_points":       strconv.Itoa(loadPoints),
		"use_core_pinning":  strconv.FormatBool(useCorePinning),
		"peak_load":         strconv.Itoa(peakLoad),
	}
	err = metadata.RecordMap(records)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Discover CPU topology.
	topology, err := topo.Discover()
	errutil.CheckWithContext(err, "Cannot discover CPU topology")
	physicalCores := topology.AvailableCores()
	allSoftwareThreds := topology.AvailableThreads()

	maxThreads := maxThreadsFlag.Value()
	if maxThreads == 0 {
		maxThreads = len(physicalCores)
	}

	// Launch Kubernetes cluster if necessary.
	var cleanup func() error
	if sensitivity.RunOnKubernetesFlag.Value() && !sensitivity.RunOnExistingKubernetesFlag.Value() {
		cleanup, err = sensitivity.LaunchKubernetesCluster()
		errutil.CheckWithContext(err, "Cannot launch Kubernetes cluster")
		defer cleanup()
	}

	// Create mutilate snap session launcher.
	mutilateSnapSession, err := mutilatesession.NewSessionLauncherDefault()
	errutil.CheckWithContext(err, "Cannot create Mutilate snap session")

	// Create USE Collector session launcher.
	useUSECollector := useUSECollectorFlag.Value()
	var useSession snap.SessionLauncher
	if useUSECollector {
		useSession, err = use.NewSessionLauncherDefault()
		errutil.CheckWithContext(err, "Cannot create USE snap session")
	}

	// Calculate value to increase QPS by on every iteration.
	qpsDelta := int(peakLoad / loadPoints)
	logrus.Debugf("Increasing QPS by %d every iteration up to peak load %d to achieve %d load points", qpsDelta, peakLoad, loadPoints)

	// Iterate over all physical cores available.
	for numberOfThreads := 1; numberOfThreads <= maxThreads; numberOfThreads++ {
		// Iterate over load points that user requested.
		for qps := qpsDelta; qps <= peakLoad; qps += qpsDelta {
			func() {
				logrus.Infof("Running %d threads of memcached with load of %d QPS", numberOfThreads, qps)

				// Check if core pinning should be enabled and set phase name.
				var isolators isolation.Decorators
				phaseName := fmt.Sprintf("memcached -t %d", numberOfThreads)
				if useCorePinning {
					var threads isolation.IntSet
					if numberOfThreads > len(physicalCores) {
						threads, err = allSoftwareThreds.Take(numberOfThreads)
						errutil.PanicWithContext(err, "Cannot take %d software threads for memcached")
					} else {
						// We have enough physical threads - take them.
						threads, err = physicalCores.Take(numberOfThreads)
						errutil.PanicWithContext(err, "Cannot take %d hardware threads (cores) for memcached")
					}
					logrus.Infof("Threads pinning enabled, using threads %q", threads.AsRangeString())
					isolators = append(isolators, isolation.Taskset{CPUList: threads})
					phaseName = isolators.Decorate(phaseName)
				}
				logrus.Debugf("Running phase: %q", phaseName)

				// Create directory where output of all the tasks will be stored.
				err := experiment.CreateRepetitionDir(appName, uid, phaseName, 0)
				errutil.PanicWithContext(err, "Cannot create repetition directory")

				// Create memcached executor.
				var memcachedExecutor executor.Executor
				if sensitivity.RunOnKubernetesFlag.Value() {
					memcachedExecutor, err = sensitivity.CreateKubernetesHpExecutor(isolators)
					errutil.PanicWithContext(err, "Cannot create Kubernetes executor")
				} else {
					memcachedExecutor = executor.NewLocalIsolated(isolators)
				}

				// Create memcached launcher and start memcached
				memcachedConfiguration := memcached.DefaultMemcachedConfig()
				memcachedConfiguration.NumThreads = numberOfThreads
				memcachedLauncher := executor.ServiceLauncher{Launcher: memcached.New(memcachedExecutor, memcachedConfiguration)}
				memcachedTask, err := memcachedLauncher.Launch()
				errutil.PanicWithContext(err, "Memcached has not been launched successfully")
				defer memcachedTask.Stop()

				// Create mutilate load generator.
				loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfiguration.IP, memcachedConfiguration.Port)
				errutil.PanicWithContext(err, "Cannot create mutilate load generator")

				// Populate memcached.
				err = loadGenerator.Populate()
				errutil.PanicWithContext(err, "Memcached cannot be populated")

				// Create tags to be used on Snap metrics.
				phase := strings.Replace(phaseName, ",", "'", -1)
				aggressor := "No aggressor " + strings.Replace(phaseName, ",", "'", -1)

				snapTags := make(map[string]interface{})
				snapTags[experiment.ExperimentKey] = uid
				snapTags[experiment.PhaseKey] = phase
				snapTags[experiment.RepetitionKey] = 0
				snapTags[experiment.LoadPointQPSKey] = qps
				snapTags[experiment.AggressorNameKey] = aggressor
				snapTags["number_of_cores"] = numberOfThreads // For backward compatibility.
				snapTags["number_of_threads"] = numberOfThreads

				var useSessionHandle executor.TaskHandle
				// Start USE Collection.
				if useUSECollector {
					useSessionHandle, err = useSession.LaunchSession(nil, snapTags)
					errutil.PanicWithContext(err, "Cannot launch Snap USE Collection session")
					defer useSessionHandle.Stop()
				}

				// Start sending traffic from mutilate cluster to memcached.
				mutilateHandle, err := loadGenerator.Load(qps, loadDuration)
				errutil.PanicWithContext(err, "Cannot start load generator")
				mutilateClusterMaxExecution := sensitivity.LoadGeneratorWaitTimeoutFlag.Value()

				mutilateTerminated, err := mutilateHandle.Wait(mutilateClusterMaxExecution)
				if err != nil {
					logrus.Errorf("Mutilate cluster failed: %q", err)
					logrus.Panic("mutilate cluster failed " + err.Error())
				}
				if !mutilateTerminated {
					msg := fmt.Sprintf("Mutilate cluster failed to stop on its own in %s. Attempting to stop...", mutilateClusterMaxExecution)
					err := mutilateHandle.Stop()
					errutil.PanicWithContext(err, msg+" Stopping mutilate cluster errored")
					logrus.Panic(msg)
				}

				// Make sure that mutilate exited with 0 status.
				exitCode, _ := mutilateHandle.ExitCode()
				if exitCode != 0 {
					logrus.Panicf("Mutilate cluster has not stopped properly. Exit status: %d.", exitCode)
				}

				if useUSECollector {
					err = useSessionHandle.Stop()
					if err != nil {
						logrus.Errorf("Snap USE session failed: %s", err)
					}
				}

				// Launch and stop Snap task to collect mutilate metrics.
				mutilateSnapSessionHandle, err := mutilateSnapSession.LaunchSession(mutilateHandle, snapTags)
				errutil.PanicWithContext(err, "Snap mutilate session has not been started successfully")
				defer func() {
					err = mutilateSnapSessionHandle.Stop()
					if err != nil {
						logrus.Errorf("Cannot stop mutilate session: %v", err)
					}
				}()
				_, err = mutilateSnapSessionHandle.Wait(0)
				errutil.PanicWithContext(err, "Snap mutilate session has not collected metrics!")

				// It is ugly but there is no other way to make sure that data is written to Cassandra as of now.
				time.Sleep(10 * time.Second)
			}()
		}
	}
}
