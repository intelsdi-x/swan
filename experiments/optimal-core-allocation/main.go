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
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
)

var (
	appName            = os.Args[0]
	useCorePinningFlag = conf.NewBoolFlag("use_core_pinning", "Enables core pinning of memcached threads", false)
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
	metadata, err := experiment.NewMetadata(uid, experiment.MetadataConfigFromFlags())
	errutil.CheckWithContext(err, "Cannot connect to metadata database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment")

	// Read configuration.
	loadDuration := sensitivity.LoadDurationFlag.Value()
	loadPoints := sensitivity.LoadPointsCountFlag.Value()
	useCorePinning := useCorePinningFlag.Value()
	peakLoad := sensitivity.PeakLoadFlag.Value()

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
	errutil.CheckWithContext(err, "Cannot save metadata")

	// Validate preconditions.
	validate.OS()

	// Discover CPU topology.
	topology, err := topo.Discover()
	errutil.CheckWithContext(err, "Cannot discover CPU topology")
	physicalCores := topology.AvailableCores()

	// Launch Kubernetes cluster if necessary.
	var cleanup func() error
	if sensitivity.RunOnKubernetesFlag.Value() && !sensitivity.RunOnExistingKubernetesFlag.Value() {
		cleanup, err = sensitivity.LaunchKubernetesCluster()
		errutil.CheckWithContext(err, "Cannot launch Kubernetes cluster")
		defer cleanup()
	}

	// Create mutilate snap session launcher.
	mutilateSnapSession, err := mutilatesession.NewSessionLauncherDefault()
	if err != nil {
		errutil.CheckWithContext(err, "Cannot create snap session")
	}

	// Calculate value to increase QPS by on every iteration.
	qpsDelta := int(peakLoad / loadPoints)
	logrus.Debugf("Increasing QPS by %d every iteration up to peak load %d to achieve %d load points", qpsDelta, peakLoad, loadPoints)

	// Iterate over all physical cores available.
	for numberOfCores := 1; numberOfCores <= len(physicalCores); numberOfCores++ {
		// Iterate over load points that user requested.
		for qps := qpsDelta; qps <= peakLoad; qps += qpsDelta {
			func() {
				logrus.Infof("Running %d threads of memcached with load of %d QPS", numberOfCores, qps)

				// Check if core pinning should be enabled and set phase name.
				var isolators isolation.Decorators
				phaseName := fmt.Sprintf("memcached -t %d", numberOfCores)
				if useCorePinning {
					cores, err := physicalCores.Take(numberOfCores)
					errutil.PanicWithContext(err, "Cannot take %d cores for memcached")
					logrus.Infof("Core pinning enabled, using cores %q", cores.AsRangeString())
					isolators = append(isolators, isolation.Taskset{CPUList: cores})
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
				memcachedConfiguration.NumThreads = numberOfCores
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

				// Start sending traffic from mutilate cluster to memcached.
				mutilateHandle, err := loadGenerator.Load(qps, loadDuration)
				errutil.PanicWithContext(err, "Cannot start load generator")
				mutilateClusterMaxExecution := sensitivity.LoadGeneratorWaitTimeoutFlag.Value()
				if !mutilateHandle.Wait(mutilateClusterMaxExecution) {
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

				// Create tags to be used on Snap metrics.
				snapTags := fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s,%s:%d",
					experiment.ExperimentKey, uid,
					experiment.PhaseKey, strings.Replace(phaseName, ",", "'", -1),
					experiment.RepetitionKey, 0,
					experiment.LoadPointQPSKey, qps,
					experiment.AggressorNameKey, "No aggressor "+strings.Replace(phaseName, ",", "'", -1),
					"number_of_cores", numberOfCores,
				)

				// Launch and stop Snap task to collect mutilate metrics.
				mutilateSnapSessionHandle, err := mutilateSnapSession.LaunchSession(mutilateHandle, snapTags)
				errutil.PanicWithContext(err, "Snap mutilate session has not been started successfully")
				// It is ugly but there is no other way to make sure that data is written to Cassandra as of now.
				time.Sleep(5 * time.Second)
				err = mutilateSnapSessionHandle.Stop()
				errutil.PanicWithContext(err, "Cannot stop Mutilate Snap session")
			}()
		}
	}
}
