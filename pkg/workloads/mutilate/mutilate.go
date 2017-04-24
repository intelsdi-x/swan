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

package mutilate

import (
	"strconv"
	"time"

	"os"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/intelsdi-x/swan/plugins/snap-plugin-collector-mutilate/mutilate/parse"
	"github.com/pkg/errors"
)

const (
	defaultPercentile             = "99"
	defaultTuningTime             = 10 * time.Second // [s]
	defaultRecords                = 10000
	defaultWarmupTime             = 10 * time.Second // [s]
	defaultAgentThreads           = 8
	defaultAgentPort              = 5556
	defaultAgentConnections       = 1
	defaultAgentConnectionsDepth  = 1
	defaultAgentAffinity          = false
	defaultAgentBlocking          = true
	defaultMasterThreads          = 8
	defaultMasterConnections      = 4
	defaultMasterConnectionsDepth = 4
	defaultMasterAffinity         = false
	defaultMasterBlocking         = true
	defaultMasterKeySize          = "30"          // [bytes]
	defaultMasterValueSize        = "200"         // [bytes]
	defaultMasterInterArrivalDist = "exponential" // disabled
	defaultMasterQPS              = 1000
)

var (
	tuningTimeFlag             = conf.NewDurationFlag("mutilate_tuning_time", "Mutilate tuning time [s].", defaultTuningTime)
	warmupTimeFlag             = conf.NewDurationFlag("mutilate_warmup_time", "Mutilate warmup time [s] (--warmup).", defaultWarmupTime)
	recordsFlag                = conf.NewIntFlag("mutilate_records", "Number of memcached records to use (-r).", defaultRecords)
	agentThreadsFlag           = conf.NewIntFlag("mutilate_agent_threads", "Mutilate agent threads (-T).", defaultAgentThreads)
	agentAgentPortFlag         = conf.NewIntFlag("mutilate_agent_port", "Mutilate agent port (-P).", defaultAgentPort)
	agentConnectionsFlag       = conf.NewIntFlag("mutilate_agent_connections", "Mutilate agent connections (-c).", defaultAgentConnections)
	agentConnectionsDepthFlag  = conf.NewIntFlag("mutilate_agent_connections_depth", "Mutilate agent connections (-d).", defaultAgentConnectionsDepth)
	agentAffinityFlag          = conf.NewBoolFlag("mutilate_agent_affinity", "Mutilate agent affinity (--affinity).", defaultAgentAffinity)
	agentBlockingFlag          = conf.NewBoolFlag("mutilate_agent_blocking", "Mutilate agent blocking (--blocking -B).", defaultAgentBlocking)
	masterThreadsFlag          = conf.NewIntFlag("mutilate_master_threads", "Mutilate master threads (-T).", defaultMasterThreads)
	masterConnectionsFlag      = conf.NewIntFlag("mutilate_master_connections", "Mutilate master connections (-C).", defaultMasterConnections)
	masterConnectionsDepthFlag = conf.NewIntFlag("mutilate_master_connections_depth", "Mutilate master connections depth (-C).", defaultMasterConnectionsDepth)
	masterAffinityFlag         = conf.NewBoolFlag("mutilate_master_affinity", "Mutilate master affinity (--affinity).", defaultMasterAffinity)
	masterBlockingFlag         = conf.NewBoolFlag("mutilate_master_blocking", "Mutilate master blocking (--blocking -B).", defaultMasterBlocking)
	masterQPSFlag              = conf.NewIntFlag("mutilate_master_qps", "Mutilate master QPS value (-Q).", defaultMasterQPS)
	masterKeySizeFlag          = conf.NewStringFlag("mutilate_master_keysize", "Length of memcached keys (-K).", defaultMasterKeySize)
	masterValueSizeFlag        = conf.NewStringFlag("mutilate_master_valuesize", "Length of memcached values (-V).", defaultMasterValueSize)
	masterInterArrivalDistFlag = conf.NewStringFlag("mutilate_master_interarrival_dist", "Inter-arrival distribution (-i).", defaultMasterInterArrivalDist)
)

// Config contains all data for running mutilate.
type Config struct {
	PathToBinary  string
	MemcachedHost string
	MemcachedPort int
	// WarmupTime represents warm up time for both Tune and Load.
	WarmupTime time.Duration

	// TODO(bplotka): Pack below parameters as flags.
	// Mutilate load Parameters
	TuningTime        time.Duration
	LatencyPercentile string
	Records           int

	AgentConnections      int    // -c
	AgentConnectionsDepth int    // Max length of request pipeline. -d
	MasterThreads         int    // -T
	MasterAffinity        bool   // Set CPU affinity for threads, round-robin (for Master)
	MasterBlocking        bool   // -B --blocking:  Use blocking epoll().  May increase latency (for Master).
	KeySize               string // Length of memcached keys. -K
	ValueSize             string // Length of memcached values. -V
	InterArrivalDist      string // Inter-arrival distribution. -i

	// Agent-mode options.
	AgentThreads           int  // Number of threads for all agents. -T
	AgentAffinity          bool // Set CPU affinity for threads, round-robin (for Agent).
	AgentBlocking          bool // -B --blocking:  Use blocking epoll().  May increase latency (for Agent).
	AgentPort              int  // Agent port. -p
	MasterConnections      int  // -C
	MasterConnectionsDepth int  // Max length of request pipeline. -D

	// Number of QPS which will be done by master itself, and only these requests
	// will measure the latency (!).
	// If it equals 0, than -Q will be not specified.
	MasterQPS int // -Q

	// Output flags.
	// TODO(bplotka): Move this flags to global experiment namespace.
	EraseTuneOutput     bool // false by default, we want to keep them, but remove during integration tests
	ErasePopulateOutput bool // false by default.

}

// DefaultMutilateConfig is a constructor for MutilateConfig with default parameters.
func DefaultMutilateConfig() Config {
	return Config{
		PathToBinary:  "mutilate",
		MemcachedHost: memcached.IPFlag.Value(),
		MemcachedPort: memcached.PortFlag.Value(),

		WarmupTime:        warmupTimeFlag.Value(),
		TuningTime:        tuningTimeFlag.Value(),
		LatencyPercentile: defaultPercentile,
		Records:           recordsFlag.Value(),

		AgentThreads:           agentThreadsFlag.Value(),
		AgentConnections:       agentConnectionsFlag.Value(),
		AgentConnectionsDepth:  agentConnectionsDepthFlag.Value(),
		AgentAffinity:          agentAffinityFlag.Value(),
		AgentBlocking:          agentBlockingFlag.Value(),
		MasterThreads:          masterThreadsFlag.Value(),
		MasterConnections:      masterConnectionsFlag.Value(),
		MasterConnectionsDepth: masterConnectionsDepthFlag.Value(),
		MasterAffinity:         masterAffinityFlag.Value(),
		MasterBlocking:         masterBlockingFlag.Value(),
		KeySize:                masterKeySizeFlag.Value(),
		ValueSize:              masterValueSizeFlag.Value(),
		InterArrivalDist:       masterInterArrivalDistFlag.Value(),
		MasterQPS:              masterQPSFlag.Value(),
		AgentPort:              agentAgentPortFlag.Value(),
	}
}

type mutilate struct {
	master executor.Executor
	agents []executor.Executor
	config Config
}

// New returns a new Mutilate Load Generator instance.
// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
func New(exec executor.Executor, config Config) executor.LoadGenerator {
	return mutilate{
		master: exec,
		agents: []executor.Executor{},
		config: config,
	}
}

// NewCluster returns a new Mutilate Load Generator instance composed of master
// and specified executors per each agent.
// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
func NewCluster(
	master executor.Executor, agents []executor.Executor, config Config) executor.LoadGenerator {
	return mutilate{
		master: master,
		agents: agents,
		config: config,
	}
}

func stopAgents(agentHandles []executor.TaskHandle) {
	for _, handle := range agentHandles {
		err := handle.Stop()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

func eraseAgentOutputs(agentHandles []executor.TaskHandle) {
	for _, handle := range agentHandles {
		err := handle.EraseOutput()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

func (m mutilate) runRemoteAgents() ([]executor.TaskHandle, error) {
	handles := []executor.TaskHandle{}

	command := getAgentCommand(m.config)
	for _, exec := range m.agents {
		handle, err := exec.Execute(command)
		if err != nil {
			// If one agent fails we need to stop these which are running.
			logrus.Errorf(
				"Mutilate: one of agents has failed (cmd: %q). Stopping already started %d agents",
				command, len(handles))
			stopAgents(handles)
			if m.config.EraseTuneOutput {
				eraseAgentOutputs(handles)
			}
			return nil, err
		}
		serviceHandle := executor.ServiceHandle{TaskHandle: handle}
		handles = append(handles, serviceHandle)
	}

	return handles, nil
}

// Populate load the initial test data into Memcached.
// Even for multi-node Mutilate, populate the Memcached is done only by master.
func (m mutilate) Populate() (err error) {
	populateCmd := getPopulateCommand(m.config)

	taskHandle, err := m.master.Execute(populateCmd)
	if err != nil {
		return err
	}

	taskHandle.Wait(0)

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.Errorf("memcached population exited with code: %d", exitCode)
	}

	if m.config.ErasePopulateOutput {
		return taskHandle.EraseOutput()
	}

	return nil
}

func (m mutilate) getQPSAndLatencyFrom(stdoutFile *os.File) (qps int, achievedSLI int, err error) {
	results, err := parse.Parse(stdoutFile)
	if err != nil {
		return qps, achievedSLI, errors.Wrap(err, "could not retrieve QPS from Mutilate Tune output")
	}

	rawQPS, ok := results.Raw[parse.MutilateQPS]
	if !ok {
		return qps, achievedSLI, errors.New("could not retrieve MutilateQPS from mutilate parser")
	}

	rawSLI, ok := results.Raw[parse.MutilatePercentile99th]
	if !ok {
		return qps, achievedSLI, errors.New("could not retrieve 99th percentile from mutilate parser")
	}

	return int(rawQPS), int(rawSLI), nil
}

// Tune returns the maximum achieved QPS where SLI is below target SLO.
func (m mutilate) Tune(slo int) (qps int, achievedSLI int, err error) {
	// Run agents when specified.
	agentHandles, err := m.runRemoteAgents()
	if err != nil {
		return qps, achievedSLI, errors.Wrap(err, "executing Mutilate Agents failed")
	}

	// Run master with tuning option.
	tuneCmd := getTuneCommand(m.config, slo, agentHandles)
	masterHandle, err := m.master.Execute(tuneCmd)
	if err != nil {
		stopAgents(agentHandles)
		if m.config.EraseTuneOutput {
			eraseAgentOutputs(agentHandles)
		}
		return qps, achievedSLI, errors.Wrapf(
			err, "mutilate Master Tune failed; Command: %q", tuneCmd)
	}

	taskHandle := executor.NewClusterTaskHandle(masterHandle, agentHandles)

	// Blocking wait for master (agents will be killed then).
	if !taskHandle.Wait(0) {
		return qps, achievedSLI, errors.Errorf("cannot terminate the Mutilate master. Leaving agents running.")
	}

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return qps, achievedSLI, err
	}

	if exitCode != 0 {
		return qps, achievedSLI, errors.New(
			"executing Mutilate Tune command returned with exit code: " +
				strconv.Itoa(exitCode))
	}

	stdoutFile, err := taskHandle.StdoutFile()
	if err != nil {
		return qps, achievedSLI, err
	}

	qps, achievedSLI, err = m.getQPSAndLatencyFrom(stdoutFile)
	if err != nil {
		return qps, achievedSLI, err
	}

	if m.config.EraseTuneOutput {
		if err = taskHandle.EraseOutput(); err != nil {
			logrus.Error("mutilate.Tune(): EraseOutput on master failed: ", err)
			return 0, 0, err
		}
	}

	return qps, achievedSLI, err
}

// Load starts a load on the specific workload with the defined loadPoint (number of QPS).
// The task will do the load for specified amount of time.
func (m mutilate) Load(qps int, duration time.Duration) (executor.TaskHandle, error) {
	agentHandles, err := m.runRemoteAgents()
	if err != nil {
		return nil, err
	}

	loadCommand := getLoadCommand(m.config, qps, duration, agentHandles)
	masterHandle, err := m.master.Execute(loadCommand)
	if err != nil {
		stopAgents(agentHandles)
		return nil, errors.Wrapf(err,
			"execution of Mutilate Master Load failed. command: %q",
			loadCommand)
	}

	return executor.NewClusterTaskHandle(masterHandle, agentHandles), nil
}
