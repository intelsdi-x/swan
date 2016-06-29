package mutilate

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"time"
)

// getAgentCommand returns command for agent.
func getAgentCommand(config Config) string {
	return fmt.Sprintf("%s -T %d -A -p %d",
		config.PathToBinary,
		config.AgentThreads,
		config.AgentPort,
	)
}

func getMasterQPSOption(config Config) string {
	masterQPSOption := ""

	if config.MasterQPS != 0 {
		masterQPSOption = fmt.Sprintf(" -Q %d", config.MasterQPS)
	}
	return masterQPSOption
}

// getPopulateCommand returns command for master with populate action.
func getPopulateCommand(config Config) string {
	return fmt.Sprintf("%s -s %s:%d --loadonly",
		config.PathToBinary,
		config.MemcachedHost,
		config.MemcachedPort,
	)
}

// getBaseMasterCommand returns master base command for both agent and agentless mode tune & load.
func getBaseMasterCommand(config Config, isNotAgentless bool) string {
	baseCommand := fmt.Sprint(
		fmt.Sprintf("%s", config.PathToBinary),
		fmt.Sprintf(" -s %s:%d", config.MemcachedHost, config.MemcachedPort),
		fmt.Sprintf(" --warmup %d --noload ", int(config.WarmupTime.Seconds())),
		fmt.Sprintf(" -K %d -V %d", config.KeySize, config.ValueSize),
		fmt.Sprintf(" -T %d -B", config.MasterThreads), // -B option for all master commands.
		fmt.Sprintf(" -d %d -c %d", config.AgentConnectionsDepth, config.AgentConnections),
	)

	// Add master-only parameters.
	if isNotAgentless {
		baseCommand += fmt.Sprint(
			fmt.Sprintf(" -D %d -C %d", config.MasterConnectionsDepth, config.MasterConnections),
			fmt.Sprintf(" -p %d", config.AgentPort),
			fmt.Sprintf("%s", getMasterQPSOption(config)),
		)
	}
	return baseCommand
}

func getBaseMasterCommandBasedOnHandlers(config Config, agentHandles []executor.TaskHandle) string {
	isNotAgentless := len(agentHandles) > 0
	baseCommand := getBaseMasterCommand(config, isNotAgentless)
	// Check if it is NOT agentless mode.
	if isNotAgentless {
		// Enlist agents.
		for _, agent := range agentHandles {
			baseCommand += fmt.Sprintf(" -a %s", agent.Address())
		}
	}

	return baseCommand
}

func getBaseMasterCoomandBaseOnInt(config Config, agentNumber int) string {
	isNotAgentless := agentNumber > 0
	baseCommand := getBaseMasterCommand(config, isNotAgentless)
	for i:=0; i<agentNumber; i++ {
		baseCommand += fmt.Sprintf(" -a %s", i)
	}
	return baseCommand
}

func getBaseLoadCommand(baseCommand string, config Config, qps int, duration time.Duration) string {
	return fmt.Sprintf("%s -q %d -t %d --swanpercentile %s",
		baseCommand, qps, int(duration.Seconds()), config.LatencyPercentile)
}

func getBaseTuneCommand(baseCommand string, config Config, slo int) string {
	return fmt.Sprintf("%s --search %s:%d -t %d --swanpercentile %s",
		baseCommand,
		config.LatencyPercentile,
		slo,
		int(config.TuningTime.Seconds()),
		config.LatencyPercentile)
}

// getLoadCommand returns master load command for both agent and agentless mode.
func getLoadCommand(
	config Config, qps int, duration time.Duration, agentHandles []executor.TaskHandle) string {
	baseCommand := getBaseMasterCommandBasedOnHandlers(config, agentHandles)
	return getBaseLoadCommand(baseCommand, config, qps, duration)
}

// getTargetLoadCommand returns master load command for both agent and agentless mode.
func getTargetLoadCommand(
config Config, qps int, duration time.Duration, agentNumber int) string {
	baseCommand := getBaseMasterCoomandBaseOnInt(config, agentNumber)
	return getBaseLoadCommand(baseCommand, config, qps, duration)
}

// getTuneCommand returns master tune command for both agent and agentless mode.
func getTuneCommand(config Config, slo int, agentHandles []executor.TaskHandle) (command string) {
	baseCommand := getBaseMasterCommandBasedOnHandlers(config, agentHandles)
	command = getBaseTuneCommand(baseCommand, config, slo)
	return command
}

// getTargetTuneCommand returns master tune command for both agent and agentless mode.
func getTargetTuneCommand(config Config, slo int, agentNumber int) (command string) {
	baseCommand := getBaseMasterCoomandBaseOnInt(config, agentNumber)
	command = getBaseTuneCommand(baseCommand, config, slo)
	return command
}
