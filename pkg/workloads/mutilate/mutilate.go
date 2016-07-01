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
	"os"
)

const (
	defaultMemcachedHost          = "127.0.0.1"
	defaultPercentile             = "99" // TODO: it is not clear if custom values are handled correctly by tune - SCE-443
	defaultTuningTime             = 10 * time.Second
	defaultWarmupTime             = 10 * time.Second
	defaultAgentThreads           = 8
	defaultAgentPort              = 5556
	defaultAgentConnections       = 1
	defaultAgentConnectionsDepth  = 1
	defaultMasterThreads          = 8
	defaultMasterConnections      = 1
	defaultMasterConnectionsDepth = 1
	defaultKeySize                = 30
	defaultValueSize              = 200
	defaultMasterQPS              = 0
)

// pathFlag represents mutilate path flag.
var pathFlag = conf.NewFileFlag("mutilate_path", "Path to mutilate binary",
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
	LatencyPercentile string

	AgentConnections      int // -c
	AgentConnectionsDepth int // Max length of request pipeline. -d
	MasterThreads         int // -T
	KeySize               int // Length of memcached keys. -K
	ValueSize             int // Length of memcached values. -V

	// Agent-mode options.
	AgentThreads           int // Number of threads for all agents. -T
	AgentPort              int // Agent port. -p
	MasterConnections      int // -C
	MasterConnectionsDepth int // Max length of request pipeline. -D

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
		PathToBinary:  pathFlag.Value(),
		MemcachedHost: defaultMemcachedHost,
		MemcachedPort: memcached.DefaultPort,

		WarmupTime:        defaultWarmupTime,
		TuningTime:        defaultTuningTime,
		LatencyPercentile: defaultPercentile,

		AgentThreads:           defaultAgentThreads,
		AgentConnections:       defaultAgentConnections,
		AgentConnectionsDepth:  defaultAgentConnectionsDepth,
		MasterThreads:          defaultMasterThreads,
		MasterConnections:      defaultAgentConnections,
		MasterConnectionsDepth: defaultAgentConnectionsDepth,
		KeySize:                defaultKeySize,
		ValueSize:              defaultValueSize,
		MasterQPS:              defaultMasterQPS,
		AgentPort:              defaultAgentPort,
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

// NewCluster returns a new Mutilate Load Generator instance composed of master
// and specified executors per each agent.
// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
func NewCluster(
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

	command := getAgentCommand(m.config)
	for _, exec := range m.agents {
		handle, err := exec.Execute(command)
		if err != nil {
			// If one agent fails we need to stop these which are running.
			logrus.Debugf(
				"One of agents failed (cmd: '%s'). Stopping already started %d agents",
				command, len(handles))
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
		return errors.New("Memcached population exited with code: " +
			strconv.Itoa(exitCode))
	}

	err = taskHandle.Clean()
	if err != nil {
		return err
	}

	if m.config.ErasePopulateOutput {
		return taskHandle.EraseOutput()
	}

	return nil
}

func (m mutilate) getQPSAndLatencyFrom(stdoutFile *os.File) (qps int, achievedSLI int, err error) {
	results, err := parse.OpenedFile(stdoutFile)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output. ")
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}

	rawQPS, ok := results.Raw[parse.MutilateQPS]
	if !ok {
		errMsg := fmt.Sprintf("Could not retrieve MutilateQPS from mutilate parser.")
		return qps, achievedSLI, errors.New(errMsg)
	}

	rawSLI, ok := results.Raw[parse.MutilatePercentileCustom]
	if !ok {
		errMsg := fmt.Sprintf("Could not retrieve Custom Percentile from mutilate parser.")
		return qps, achievedSLI, errors.New(errMsg)
	}

	return int(rawQPS), int(rawSLI), nil
}

// Tune returns the maximum achieved QPS where SLI is below target SLO.
func (m mutilate) Tune(slo int) (qps int, achievedSLI int, err error) {
	// Run agents when specified.
	agentHandles, err := m.runRemoteAgents()
	if err != nil {
		return qps, achievedSLI,
			fmt.Errorf("Executing Mutilate Agents failed; %s", err.Error())
	}

	defer func() {
		logrus.Debug("Stopping %d agents", len(agentHandles))
		stopAgents(agentHandles)
		cleanAgents(agentHandles)
		if m.config.EraseTuneOutput {
			eraseAgentOutputs(agentHandles)
		}
	}()

	// Run master with tuning option.
	tuneCmd := getTuneCommand(m.config, slo, agentHandles)
	masterHandle, err := m.master.Execute(tuneCmd)
	if err != nil {
		logrus.Debug("Mutilate master execution failed (cmd: '%s')", tuneCmd)
		return qps, achievedSLI,
			fmt.Errorf("Mutilate Master Tune failed; Command: %s; %s", tuneCmd, err.Error())
	}

	taskHandle := executor.NewClusterTaskHandle(masterHandle, agentHandles)

	// Blocking wait for master.
	if !taskHandle.Wait(0) {
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

	qps, achievedSLI, err = m.getQPSAndLatencyFrom(stdoutFile)
	if err != nil {
		return qps, achievedSLI, err
	}

	err = taskHandle.Clean()
	if err != nil {
		return 0, 0, err
	}

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

	masterHandle, err := m.master.Execute(
		getLoadCommand(m.config, qps, duration, agentHandles))
	if err != nil {
		stopAgents(agentHandles)
		return nil, fmt.Errorf(
			"Execution of Mutilate Master Load failed; Command: %s; %s",
			getLoadCommand(m.config, qps, duration, agentHandles), err.Error())
	}

	return executor.NewClusterTaskHandle(masterHandle, agentHandles), nil
}

func (m mutilate) Name() string {
	return "Mutilate"
}

func (m mutilate) GetTuneParameters(slo int) string {
	tuneCommand := getTargetTuneCommand(m.config, slo, m.config.AgentConnections)
	return tuneCommand
}

func (m mutilate) GetLoadParameters(qps int, duration time.Duration) string {
	loadCommand := getTargetLoadCommand(m.config, qps, duration, m.config.AgentConnections)
	return loadCommand
}

