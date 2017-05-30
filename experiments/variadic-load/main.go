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
	"bytes"
	"fmt"
	"io/ioutil"
	"math"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/experiments/memcached-sensitivity-profile/common"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/topology"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	_ "github.com/intelsdi-x/swan/pkg/utils/unshare"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/plugins/snap-plugin-collector-mutilate/mutilate/parse"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
)

var (
	appName             = os.Args[0]
	influxdbURLFlag     = conf.NewStringFlag("influxdb_url", "Tag phases in influxdb", "")
	innerRepetitionFlag = conf.NewIntFlag("inner_repetition", "inner repetition for running on BE workload", 1)
	minLoadFlag         = conf.NewIntFlag("min_load", "Minimum load for QPS", 10000)
)

func main() {
	// Preparing application - setting name, help, aprsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	uid := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, uid)

	// Validate preconditions.
	validate.OS()

	// Create isolations.
	hpIsolation, l1Isolation, llcIsolation := topology.NewIsolations()

	// Create executors with cleanup function.
	hpExecutor, beExecutorFactory, cleanup, err := sensitivity.PrepareExecutors(hpIsolation)
	errutil.Check(err)
	defer cleanup()

	// Create BE workloads.
	beLaunchers, err := sensitivity.PrepareAggressors(l1Isolation, llcIsolation, beExecutorFactory)
	errutil.Check(err)

	// Create HP workload.
	memcachedConfig := memcached.DefaultMemcachedConfig()
	hpLauncher := executor.ServiceLauncher{Launcher: memcached.New(hpExecutor, memcachedConfig)}

	// Load generator.
	loadGenerator, err := common.PrepareMutilateGenerator(memcachedConfig.IP, memcachedConfig.Port)
	errutil.Check(err)

	// Peak load
	peakload := sensitivity.PeakLoadFlag.Value()
	log.Infof("using peakload %d", peakload)

	loadPoints := sensitivity.LoadPointsCountFlag.Value()
	loadDuration := sensitivity.LoadDurationFlag.Value()
	repetitions := sensitivity.RepetitionsFlag.Value()
	minLoad := minLoadFlag.Value()

	// hp
	hpHandle, err := hpLauncher.Launch()
	errutil.Check(err)
	defer hpHandle.Stop()

	// populate
	errutil.Check(loadGenerator.Populate())

	postInflux := func(data string) {
		// influxdb tag
		if influxdbURLFlag.Value() != "" {
			for i := 0; i < 5; i++ {
				r, err := http.Post(influxdbURLFlag.Value(), "", bytes.NewBufferString(data))

				if err != nil {
					log.Warnf("error writing tag to influxdb: %v", err)
				} else {
					body, err := ioutil.ReadAll(r.Body)
					errutil.Check(err)
					log.Debug("influx response", string(body))
				}
				break
			}
		}
	}

	for repetition := 0; repetition < repetitions; repetition++ {
		for _, beLauncher := range beLaunchers {

			// aggressor name
			aggressorName := "None"
			if beLauncher.Launcher != nil {
				aggressorName = beLauncher.Launcher.String()
			}
			log.Infof("Aggressor: %s", aggressorName)
			postInflux(fmt.Sprintf("tags aggressor=\"%s\",repetition=%d", aggressorName, repetition))

			// BE
			var beHandle executor.TaskHandle
			if beLauncher.Launcher != nil {
				beHandle, err = beLauncher.Launcher.Launch()
				errutil.Check(err)
			}

			for innerRepetition := 0; innerRepetition < innerRepetitionFlag.Value(); innerRepetition++ {
				for loadPoint := 0; loadPoint < loadPoints; loadPoint++ {
					qps := int((math.Sin(float64(loadPoint)/float64(loadPoints)*math.Pi*2))*float64(peakload/2)) + peakload/2
					if qps < minLoad {
						qps = minLoad
					}
					log.Infof("QPS: %d (%d/%d) %s (%d/%d)", qps, innerRepetition, innerRepetitionFlag.Value(), aggressorName, repetition, repetitions)
					// load gen
					loadGeneratorHandle, err := loadGenerator.Load(qps, loadDuration)
					errutil.Check(err)
					loadGeneratorHandle.Wait(0)
					out, err := loadGeneratorHandle.StdoutFile()
					results, err := parse.File(out.Name())
					errutil.Check(err)
					log.Infof("mutilate output: %#v", results.Raw)

					data := []string{}
					for k, v := range results.Raw {
						data = append(data, fmt.Sprintf("%s=%f", strings.Replace(k, "/", "_", -1), v))
					}
					influxData := fmt.Sprintf("mutilate,aggressor=\"%s\",repetition=%d %s", aggressorName, repetition, strings.Join(data, ","))
					log.Debug("influx data:", influxData)
					postInflux(influxData)

				}
			}
			if beHandle != nil {
				beHandle.Stop()
			}
		}
	}

	log.Infof("Ended experiment %s with uid %s in %s", appName, uid, time.Since(experimentStart).String())
}
