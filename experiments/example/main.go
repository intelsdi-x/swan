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

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/mutilate"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	// This import is used to launch new PID namespace to make sure that all the processes will be terminated when experiment ends.
	_ "github.com/intelsdi-x/swan/pkg/utils/unshare"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
)

// We want to allow experimenter to pass some configuration to the experiment. "Flag" is appended to variable name by convention.
var (
	appName            = os.Args[0]
	minThreadCountFlag = conf.NewIntFlag("example_min_thread_count", "Minimum count of Memcached threads to be tested in the experiment", 1)
	maxThreadCountFlag = conf.NewIntFlag("example_max_thread_count", "Maximum count of Memcached threads to be tested in the experiment", 4)
)

func main() {
	// It's always useful to store experiment start time for future reference.
	experimentStart := time.Now()

	// Configuring application - parsing flags, configuring log level, dumping the configuration (if requested).
	experiment.Configure()

	// Generate an experiment ID that will be used to tag all the metrics collected as well as to identify experiment metadata.
	uid := uuid.New()

	// Initialize logger (and log some basic information: experiment name, UUID generated above etc).
	logger.Initialize(appName, uid)

	// Connect to metadata database (Cassandra is the only supported database).
	// Besides experiment results and platform metrics (we use Snap to gather them) we save certain deta about experiment configuration and environment (metadata).
	metadata, err := experiment.NewMetadata(uid, experiment.DefaultMetadataConfig())
	// errutil.CheckWithContext() is a helper function that will panic on error and provide some additional information about error origin.
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// We should validate OS configuration to make sure that it is ready to execute the experiment binary.
	validate.OS()

	// Now we need to get hold of experiment configuration.
	// Thread count is described at the rop of this file.
	minThreadCount := minThreadCountFlag.Value()
	maxThreadCount := maxThreadCountFlag.Value()
	// Each iteration of Memcached stressing will last amount of time that loadDuration defines.
	loadDuration := sensitivity.LoadDurationFlag.Value()
	// To determine Memcached capacity we will attempt to send various amount of QPS number of times.
	numberOfLoadPoints := sensitivity.LoadPointsCountFlag.Value()
	// Maximum amount of QPS to be able to calculate value for each load point.
	maxQPS := sensitivity.PeakLoadFlag.Value()

	// When configuration flags' values are available we should save it to the metadata database.
	// Each value needs to be a string so some conversion might be necessary
	experimentConfiguration := make(map[string]string)
	experimentConfiguration["min_thread_count"] = strconv.Itoa(minThreadCount)
	experimentConfiguration["max_thread_count"] = strconv.Itoa(maxThreadCount)
	experimentConfiguration["load_points"] = loadDuration.String()
	// Saving command line arguments and name of the binary might help with future analysis.
	experimentConfiguration["command_arguments"] = strings.Join(os.Args[1:], " ")
	experimentConfiguration["experiment_name"] = os.Args[0]
	// Finally - sava configuration to Cassandra.
	err = metadata.RecordMap(experimentConfiguration)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")

	// Create Mutilate Snap session launcher - it will be used to gather metrics about Memcached performance.
	mutilateSnapSession, err := mutilatesession.NewSessionLauncherDefault()
	errutil.CheckWithContext(err, "Cannot create Mutilate snap session")

	// Instansce of executor.Executor that allows to launch processes locally.
	executor := executor.NewLocal()

	// Instance of executor.LoadGenerator that will launch Mutilate locally,
	mutilateLauncher := mutilate.New(executor, mutilate.DefaultMutilateConfig())

	// Iterate over range of cores defined in the experiment configuration.
	for threadCount := minThreadCount; threadCount <= maxThreadCount; threadCount++ {
		// Create Memcached configuration and set number of threads.
		memcachedConfiguration := memcached.DefaultMemcachedConfig()
		memcachedConfiguration.NumThreads = threadCount

		// Iterate over number of load points.
		for loadPoint := 1; loadPoint <= numberOfLoadPoints; loadPoint++ {
			// Calculating number of QPS to be sent to Memcached.
			qps := int(maxQPS / numberOfLoadPoints * loadPoint)
			// The closure represents single phase of experiment (combination of number QPS and number of threads). It allows using  defer to stop all the processes launched.
			func() {
				// Instance of executor.Launcher that uses Executor and Memcached configuration to launch it.
				memcachedLauncher := memcached.New(executor, memcachedConfiguration)

				// Launching Memcached...
				memcached, err := memcachedLauncher.Launch()
				errutil.CheckWithContext(err, fmt.Sprintf("Cannot launch Memcached with %d threads, %d QPS and load duration of %s", threadCount, qps, loadDuration))
				// Making sure to stop Memcached at the end of phase.
				defer memcached.Stop()

				// Launching Mutilate to stress Memcached...
				mutilateTask, err := mutilateLauncher.Load(qps, loadDuration)
				errutil.CheckWithContext(err, fmt.Sprintf("Cannot launch Mutilate with %d threads, %d QPS and load duration of %s", threadCount, qps, loadDuration))
				defer mutilateTask.Stop()
				// Waiting until load generating finishes.
				mutilateTask.Wait(0)

				// Preparing tags for Snap metrics (they will be used to identify unique phases).
				tags := map[string]interface{}{
					"thread_count": threadCount,
					"load_point":   loadPoint,
					"qps":          qps,
					experiment.ExperimentKey: uid,
				}
				// Launching Mutilate Snap session in order to gather metrics on Memcached performance.
				mutilateSessionHandle, err := mutilateSnapSession.LaunchSession(mutilateTask, tags)
				errutil.CheckWithContext(err, fmt.Sprintf("Cannot gather Memcached performance metrics with %d threads, %d QPS and load duration of %s", threadCount, qps, loadDuration))
				defer mutilateSessionHandle.Stop()
				// Wait for Snap session to finish.
				_, err = mutilateSessionHandle.Wait(0)
				errutil.CheckWithContext(err, fmt.Sprintf("Cannot finish gathering Memcached performance metrics with %d threads, %d QPS and load duration of %s", threadCount, qps, loadDuration))
			}()
		}
	}
}
