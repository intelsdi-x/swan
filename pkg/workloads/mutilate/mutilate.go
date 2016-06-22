package mutilate

import (
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"regexp"
	"strconv"
	"time"

	"path"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/intelsdi-x/swan/pkg/workloads/memcached"
	"github.com/shopspring/decimal"
)

const (
	defaultMemcachedHost          = "127.0.0.1"
	defaultMemcachedPercentile    = "99" // TODO: it is not clear if custom values are handled correctly by tune - SCE-443
	defaultMemcachedTuningTime    = 10 * time.Second
	defaultMemcachedWarmupTime    = 10 * time.Second
	defaultAgentThreads           = 24
	defaultAgentConnections       = 1
	defaultAgentConnectionsDepth  = 1
	defaultMasterThreads          = 24
	defaultMasterConnections      = 1
	defaultMasterConnectionsDepth = 1
	defaultKeySize                = 30
	defaultValueSize              = 200
)

// PathFlag represents mutilate path flag.
var PathFlag = conf.NewStringFlag(
	"muitilate_path",
	"Path to mutilate binary",
	path.Join(fs.GetSwanWorkloadsPath(), "data_caching/memcached/mutilate/mutilate"),
)

// Config contains all data for running mutilate.
type Config struct {
	PathToBinary  string
	MemcachedHost string
	MemcachedPort int
	// WarmupTime represents warm up time for both Tune and Load.
	WarmupTime time.Duration

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

	// Output flags.
	EraseTuneOutput     bool // false by default, we want to keep them, but remove during integration tests
	ErasePopulateOutput bool // false by default.
}

// DefaultMutilateConfig is a constructor for MutilateConfig with default parameters.
func DefaultMutilateConfig() Config {
	percentile, _ := decimal.NewFromString(defaultMemcachedPercentile)

	return Config{
		PathToBinary:  PathFlag.Value(),
		MemcachedHost: defaultMemcachedHost,
		MemcachedPort: memcached.DefaultPort,

		WarmupTime:        defaultMemcachedWarmupTime,
		TuningTime:        defaultMemcachedTuningTime,
		LatencyPercentile: percentile,

		AgentThreads:           defaultAgentThreads,
		AgentConnections:       defaultAgentConnections,
		AgentConnectionsDepth:  defaultAgentConnectionsDepth,
		MasterThreads:          defaultMasterThreads,
		MasterConnections:      defaultAgentConnections,
		MasterConnectionsDepth: defaultAgentConnectionsDepth,
		KeySize:                defaultKeySize,
		ValueSize:              defaultValueSize,
		// TODO(bplotka): Do we need -B"
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

// New returns a new Mutilate Load Generator instance.
// Mutilate is a load generator for Memcached.
// https://github.com/leverich/mutilate
func NewMultinode(master executor.Executor, agents []executor.Executor, config Config) workloads.LoadGenerator {
	return mutilate{
		master: master,
		agents: agents,
		config: config,
	}
}

// Populate load the initial test data into Memcached.
func (m mutilate) Populate() (err error) {
	populateCmd := m.getPopulateCommand()

	taskHandle, err := m.master.Execute(populateCmd)
	if err != nil {
		return err
	}
	defer taskHandle.Clean()

	taskHandle.Wait(0)

	exitCode, err := taskHandle.ExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		return errors.New("Memcached population exited with code: " +
			strconv.Itoa(exitCode))
	}

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
		// TODO(skonefal): TBD
	}

	// Run master with tuning option.
	tuneCmd := m.getTuneCommand(slo, agentHandles)
	masterHandle, err := m.master.Execute(tuneCmd)
	if err != nil {
		errMsg := fmt.Sprintf("Executing Mutilate Tune command %s failed; ", tuneCmd)
		return qps, achievedSLI, errors.New(errMsg + err.Error())
	}

	taskHandle := MutilateTaskHandle{
		master: masterHandle,
		agents: agentHandles,
	}

	taskHandle.Wait(0)

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
	qps, achievedSLI, err = getQPSAndLatencyFrom(stdoutFile)
	if err != nil {
		errMsg := fmt.Sprintf("Could not retrieve QPS from Mutilate Tune output. ")
		return qps, achievedSLI, errors.New(errMsg + err.Error())
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
func (m mutilate) Load(qps int, duration time.Duration) (handle executor.TaskHandle, err error) {
	agents, err := m.runRemoteAgents()
	if err != nil {
		return handle, err
	}

	master, err := m.master.Execute(m.getLoadCommand(qps, duration, agents))
	if err != nil {
		// TODO(skonefal): Stop agents.
		return master, err
	}

	handle = MutilateTaskHandle{
		master: master,
		agents: agents,
	}

	return handle, err
}

func (m mutilate) getAgentStartCommand() string {
	return fmt.Sprintf("%s -T %d -A",
		m.config.PathToBinary,
		m.config.AgentThreads,
	)
}

func (m mutilate) runRemoteAgents() (handles []executor.TaskHandle, err error) {
	command := m.getAgentStartCommand()
	for _, exec := range m.agents {
		handle, err := exec.Execute(command)
		if err != nil {
			// TODO(skonefal): TBD
		}
		handles = append(handles, handle)
	}
	return handles, err
}

func getEnlistAgents(agentHandles []executor.TaskHandle) string {
	enlistAgentsString := ""
	for _, agent := range agentHandles {
		enlistAgentsString += fmt.Sprintf(" -a %s", agent.Address())
	}
	return enlistAgentsString
}

func (m mutilate) getPopulateCommand() string {
	return fmt.Sprintf("%s -s %s:%d --loadonly",
		m.config.PathToBinary,
		m.config.MemcachedHost,
		m.config.MemcachedPort,
	)
}

func (m mutilate) getBaseMutilateCommand(handles []executor.TaskHandle) string {
	enlistsAgents := getEnlistAgents(handles)
	return fmt.Sprint(
		fmt.Sprintf("%s", m.config.PathToBinary),
		fmt.Sprintf(" -s %s:%d", m.config.MemcachedHost, m.config.MemcachedPort),
		fmt.Sprintf(" --warmup=%d --noload ", int(m.config.WarmupTime.Seconds())),
		fmt.Sprintf(" -K %d -V %d", m.config.KeySize, m.config.ValueSize),
		fmt.Sprintf(" -T %d", m.config.MasterThreads),
		fmt.Sprintf(" -D %d -C %d", m.config.MasterConnectionsDepth, m.config.MasterConnections),
		fmt.Sprintf(" -d %d -c %d", m.config.AgentConnectionsDepth, m.config.AgentConnections),
		fmt.Sprintf(" %s", enlistsAgents),
	)
}

func (m mutilate) getLoadCommand(qps int, duration time.Duration, agentHandles []executor.TaskHandle) string {
	baseCommand := m.getBaseMutilateCommand(agentHandles)
	return fmt.Sprintf("%s -q %d -t %d --swanpercentile=%s",
		baseCommand, qps, int(duration.Seconds()), m.config.LatencyPercentile.String())
}

func (m mutilate) getTuneCommand(slo int, agentHandles []executor.TaskHandle) (command string) {
	baseCommand := m.getBaseMutilateCommand(agentHandles)
	command = fmt.Sprintf("%s --search=%s:%d -t %d",
		baseCommand, m.config.LatencyPercentile.String(), slo, int(m.config.TuningTime.Seconds()))
	return command
}

func matchNotFound(match []string) bool {
	return match == nil || len(match) < 2 || len(match[1]) == 0
}

func getQPSAndLatencyFrom(outputReader io.Reader) (qps int, latency int, err error) {
	buff, err := ioutil.ReadAll(outputReader)
	if err != nil {
		return qps, latency, err
	}
	output := string(buff)
	qps, qpsError := getQPSFrom(output)
	latency, latencyError := getLatencyFrom(output)

	var errorMsg string

	if qpsError != nil {
		errorMsg += "Could not get QPS from output: " + qpsError.Error() + ". "
	}
	if latencyError != nil {
		errorMsg += "Could not get Latency from output: " + latencyError.Error() + ". "
	}

	if errorMsg != "" {
		return 0, 0, errors.New(errorMsg)
	}

	return qps, latency, err
}

func getQPSFrom(output string) (qps int, err error) {
	getQPSRegex := regexp.MustCompile(`Total QPS =\s(\d+)`)
	match := getQPSRegex.FindStringSubmatch(output)
	if matchNotFound(match) {
		errMsg := fmt.Sprintf(
			"Cannot find regex match in output: %s", output)
		return 0, errors.New(errMsg)
	}

	qps, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse integer from string: %s; ", match[1])
		return 0, errors.New(errMsg + err.Error())
	}

	return qps, err
}

func getLatencyFrom(output string) (latency int, err error) {
	getLatencyRegex := regexp.MustCompile(`Swan latency for percentile \d+.\d+:\s(\d+)`)
	match := getLatencyRegex.FindStringSubmatch(output)
	if matchNotFound(match) {
		return 0, fmt.Errorf("Cannot find regex match in output: %s", output)
	}

	latency, err = strconv.Atoi(match[1])
	if err != nil {
		errMsg := fmt.Sprintf(
			"Cannot parse integer from string: %s; ", match[1])
		return 0, errors.New(errMsg + err.Error())
	}

	return latency, err
}
