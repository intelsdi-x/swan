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
func getBaseMasterCommand(config Config, agentHandles []executor.TaskHandle) string {
	baseCommand := fmt.Sprint(
		fmt.Sprintf("%s", config.PathToBinary),
		fmt.Sprintf(" -s %s:%d", config.MemcachedHost, config.MemcachedPort),
		fmt.Sprintf(" --warmup %d --noload ", int(config.WarmupTime.Seconds())),
		fmt.Sprintf(" -K %d -V %d", config.KeySize, config.ValueSize),
		fmt.Sprintf(" -T %d -B", config.MasterThreads), // -B option for all master commands.
		fmt.Sprintf(" -d %d -c %d", config.AgentConnectionsDepth, config.AgentConnections),
	)

	// Check if it is NOT agentless mode.
	if len(agentHandles) > 0 {
		// Add master-only parameters.
		baseCommand += fmt.Sprint(
			fmt.Sprintf(" -D %d -C %d", config.MasterConnectionsDepth, config.MasterConnections),
			fmt.Sprintf(" -p %d %s", config.AgentPort, getMasterQPSOption(config)),
		)

		// Enlist agents.
		for _, agent := range agentHandles {
			baseCommand += fmt.Sprintf(" -a %s", agent.Address())
		}
	}

	return baseCommand
}

// getLoadCommand returns master load command for both agent and agentless mode.
func getLoadCommand(
	config Config, qps int, duration time.Duration, agentHandles []executor.TaskHandle) string {
	baseCommand := getBaseMasterCommand(config, agentHandles)
	return fmt.Sprintf("%s -q %d -t %d --swanpercentile %s",
		baseCommand, qps, int(duration.Seconds()), config.LatencyPercentile.String())
}

// getTuneCommand returns master tune command for both agent and agentless mode.
func getTuneCommand(config Config, slo int, agentHandles []executor.TaskHandle) (command string) {
	baseCommand := getBaseMasterCommand(config, agentHandles)
	command = fmt.Sprintf("%s --search %s:%d -t %d --swanpercentile %s",
		baseCommand,
		config.LatencyPercentile.String(),
		slo,
		int(config.TuningTime.Seconds()),
		config.LatencyPercentile.String())
	return command
}
