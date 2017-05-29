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

package common

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/workloads/mutilate"
)

var (
	mutilatePercentileFlag = conf.NewStringFlag(
		"experiment_tail_latency_percentile",
		"Tail latency percentile for Memcached SLI",
		"99")

	mutilateMasterFlag = conf.NewStringFlag(
		"experiment_mutilate_master_address",
		"Address where Mutilate Master will be launched. Master coordinate agents and measures SLI.",
		"127.0.0.1")

	mutilateAgentsFlag = conf.NewStringSliceFlag(
		"experiment_mutilate_agent_addresses",
		"Addresses where Mutilate Agents will be launched, separated by commas (e.g: \"192.168.1.1,192.168.1.2\" Agents generate actual load on Memcached.",
		[]string{},
	)
)

// PrepareMutilateGenerator creates new LoadGenerator based on mutilate.
func PrepareMutilateGenerator(memcachedIP string, memcachedPort int) (executor.LoadGenerator, error) {
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcachedIP
	mutilateConfig.MemcachedPort = memcachedPort
	mutilateConfig.LatencyPercentile = mutilatePercentileFlag.Value()

	agentsLoadGeneratorExecutors := []executor.Executor{}

	masterLoadGeneratorExecutor, err := executor.NewShell(mutilateMasterFlag.Value())
	if err != nil {
		return nil, err
	}

	// Pack agents.
	for _, agent := range mutilateAgentsFlag.Value() {
		remoteExecutor, err := executor.NewShell(agent)
		if err != nil {
			return nil, err
		}
		agentsLoadGeneratorExecutors = append(agentsLoadGeneratorExecutors, remoteExecutor)
	}
	logrus.Debugf("Added %d mutilate agent(s) to mutilate cluster", len(agentsLoadGeneratorExecutors))

	// Validate mutilate cluster executors and their limit of
	// number of open file descriptors. Sane mutilate configuration requires
	// more than default (1024) for mutilate cluster.
	validate.ExecutorsNOFILELimit(
		append(agentsLoadGeneratorExecutors, masterLoadGeneratorExecutor),
	)

	// Initialize Mutilate Load Generator.
	mutilateLoadGenerator := mutilate.NewCluster(
		masterLoadGeneratorExecutor,
		agentsLoadGeneratorExecutors,
		mutilateConfig)

	return mutilateLoadGenerator, nil
}
