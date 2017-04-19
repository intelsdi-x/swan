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

	mutilateAgentsFlag = conf.NewSliceFlag(
		"experiment_mutilate_agent_addresses",
		"Addresses where Mutilate Agents will be launched. Agents generate actual load on Memcached.",
		[]string{"127.0.0.1", "127.0.0.1"},
	)
)

// PrepareMutilateGenerator creates new LoadGenerator based on mutilate.
func PrepareMutilateGenerator(memcachedIP string, memcachedPort int) (executor.LoadGenerator, error) {
	mutilateConfig := mutilate.DefaultMutilateConfig()
	mutilateConfig.MemcachedHost = memcachedIP
	mutilateConfig.MemcachedPort = memcachedPort
	mutilateConfig.LatencyPercentile = mutilatePercentileFlag.Value()

	agentsLoadGeneratorExecutors := []executor.Executor{}

	masterLoadGeneratorExecutor, err := executor.NewRemoteFromIP(mutilateMasterFlag.Value())
	if err != nil {
		return nil, err
	}

	// Pack agents.
	for _, agent := range mutilateAgentsFlag.Value() {
		remoteExecutor, err := executor.NewRemoteFromIP(agent)
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
