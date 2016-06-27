package mutilate

import (
	"errors"
	"fmt"
	"strconv"
	"time"

	"path"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/misc/snap-plugin-collector-mutilate/mutilate/parse"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/shopspring/decimal"
)

const (
	defaultMemcachedHost          = "127.0.0.1"
	defaultPercentile             = "99" // TODO: it is not clear if custom values are handled correctly by tune - SCE-443
	defaultTuningTime             = 10 * time.Second
	defaultWarmupTime             = 10 * time.Second
	defaultAgentThreads           = 24
	defaultAgentConnections       = 1
	defaultAgentConnectionsDepth  = 1
	defaultMasterThreads          = 24
	defaultMasterConnections      = 1
	defaultMasterConnectionsDepth = 1
	defaultKeySize                = 30
	defaultValueSize              = 200
	defaultMasterQPS              = 0
)

// pathFlag represents mutilate path flag.
var pathFlag = conf.NewStringFlag("muitilate_path", "Path to mutilate binary",
	path.Join(fs.GetSwanWorkloadsPath(), "data_caching/memcached/mutilate/mutilate"),
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
	LatencyPercentile decimal.Decimal

	// Number of threads for all agents.
	AgentThreads           int // -T
	AgentConnections       int // -c
	AgentConnectionsDepth  int // Max length of request pipeline. -d
	MasterThreads          int // -T
	MasterConnections      int // -C
	MasterConnectionsDepth int // Max length of request pipeline. -D
	KeySize                int // Length of memcached keys. -K
	ValueSize              int // Length of memcached values. -V
	// TODO(bp): Decide if we want to use -B option as well.

	// Number of QPS which will be done by master itself, and only these requests
	// will measure the latency (!).
	// If it equals 0, than remote -Q option at all from master.
	// TODO(bp): Do we need to have that just per whole Load Generator?
	MasterQPS int // -Q

	// Output flags.
	EraseTuneOutput     bool // false by default, we want to keep them, but remove during integration tests
	ErasePopulateOutput bool // false by default.
}

// DefaultMutilateConfig is a constructor for MutilateConfig with default parameters.
func DefaultMutilateConfig() Config {
	percentile, _ := decimal.NewFromString(defaultPercentile)

	return Config{
		PathToBinary:  pathFlag.Value(),
		MemcachedHost: defaultMemcachedHost,
		MemcachedPort: memcached.DefaultPort,

		WarmupTime:        defaultWarmupTime,
		TuningTime:        defaultTuningTime,
		LatencyPercentile: percentile,

		AgentThreads:           defaultAgentThreads,
		AgentConnections:       defaultAgentConnections,
		AgentConnectionsDepth:  defaultAgentConnectionsDepth,
		MasterThreads:          defaultMasterThreads,
		MasterConnections:      defaultAgentConnections,
		MasterConnectionsDepth: defaultAgentConnectionsDepth,
		KeySize:                defaultKeySize,
		ValueSize:              defaultValueSize,
		MasterQPS:              defaultMasterQPS,
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
func New(exec executor.Executor, config Config) workloads.LoadGenerator {
	return mutilate{
		master: exec,
		agents: []executor.Executor{},
		config: config,
	}
}

// NewClustered returns a new Mutilate Load Generator instance composed of master
// and specified executors per each agent.
// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
func NewClustered(
	master executor.Executor, agents []executor.Executor, config Config) workloads.LoadGenerator {
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

func cleanAgents(agentHandles []executor.TaskHandle) {
	for _, handle := range agentHandles {
		err := handle.Clean()
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

	command := m.getAgentMutilateCommand()
	for _, exec := range m.agents {
		handle, err := exec.Execute(command)
		if err != nil {
			// If one agent fails we need to stop these which are running.
			logrus.Debug("One of agents failed (cmd: '",
				command, "'). Stopping already started ", len(handles), " agents.")
			stopAgents(handles)
			cleanAgents(handles)
			if m.config.EraseTuneOutput {
				eraseAgentOutputs(handles)
			}
			return nil, err
		}
		handles = append(handles, handle)
	}

	return handles, nil
}

// Populate load the initial test data into Memcached.
// Even for multi-node Mutilate, populate the Memcached is done only by master.
func (m mutilate) Populate() (err error) {
	populateCmd := m.getPopulateCommand()

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
		return errors.New("Memcached population exited with code: " +
			strconv.Itoa(exitCode))
	}

	taskHandle.Clean()
	if m.config.ErasePopulateOutput {
		return taskHandle.EraseOutput()
	}
	return nil
}

// Tune returns the maximum achieved QPS where SLI is below target SLO.
func (m mutilate) Tune(slo int) (qps int, achievedSLI int, err error) {
	// Run agents when specified.
	agentHandles, err := m.runRemoteAgents()
	if err != nil {
		return qps, achievedSLI,
			fmt.Errorf("Executing Mutilate Agents failed; %s", err.Error())
	}

	// Run master with tuning option.
	tuneCmd := m.getTuneCommand(slo, agentHandles)
	masterHandle, err := m.master.Execute(tuneCmd)
	if err != nil {
		logrus.Debug("Mutilate master execution failed (cmd: '",
			tuneCmd, "'). Stopping already started ", len(agentHandles), " agents.")
		stopAgents(agentHandles)
		cleanAgents(agentHandles)
		if m.config.EraseTuneOutput {
			eraseAgentOutputs(agentHandles)
		}
		return qps, achievedSLI,
			fmt.Errorf("Mutilate Master Tune failed; Command: %s; %s", tuneCmd, err.Error())
	}

	taskHandle := executor.NewClusterTaskHandle(masterHandle, agentHandles)

	// Blocking wait for master.
	if !taskHandle.Wait(0) {
		// If master was not terminate, then agents could be still running!
		stopAgents(agentHandles)
		cleanAgents(agentHandles)
		if m.config.EraseTuneOutput {
			eraseAgentOutputs(agentHandles)
		}
		return qps, achievedSLI, fmt.Errorf("Cannot terminate the Mutilate master. Stopping agents.")
	}

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return qps, achievedSLI, err
	}

	if exitCode != 0 {
		return qps, achievedSLI, errors.New(
			"Executing Mutilate Tune command returned with exit code: " +
				strconv.Itoa(exitCode))
	}

	stdoutFile, err := taskHandle.StdoutFile()
	if err != nil {
		return qps, achievedSLI, err
	}

	metricsMap, err := parse.OpenedFile(stdoutFile)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output. ")
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}

	rawQPS, ok := metricsMap[parse.MutilateQPS]
	if !ok {
		errMsg := fmt.Sprintf("Could not retrieve MutilateQPS from mutilate parser.")
		return qps, achievedSLI, errors.New(errMsg)
	}
	qps = int(rawQPS)

	// We don't need to have 'exact' flag retrieved.
	// TODO(bplotka): Do we need percentile in Decimal type? Float64 is not enough?
	// If float64 is not enough we need to fix parser and mutilate collector as well.
	floatPercentile, _ := m.config.LatencyPercentile.Float64()
	rawSLI, ok := metricsMap[parse.GenerateCustomPercentileKey(floatPercentile)]
	if !ok {
		errMsg := fmt.Sprintf("Could not retrieve Custom Percentile from mutilate parser.")
		return qps, achievedSLI, errors.New(errMsg)
	}
	achievedSLI = int(rawSLI)

	taskHandle.Clean()
	if m.config.EraseTuneOutput {
		if err := taskHandle.EraseOutput(); err != nil {
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

	masterHandle, err := m.master.Execute(m.getLoadCommand(qps, duration, agentHandles))
	if err != nil {
		stopAgents(agentHandles)
		return nil, fmt.Errorf(
			"Execution of Mutilate Master Load failed; Command: %s; %s",
			m.getLoadCommand(qps, duration, agentHandles), err.Error())
	}

	return executor.NewClusterTaskHandle(masterHandle, agentHandles), nil
}
