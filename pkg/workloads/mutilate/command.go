package mutilate

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"time"
)

func (m mutilate) getAgentMutilateCommand() string {
	return fmt.Sprintf("%s -T %d -A",
		m.config.PathToBinary,
		m.config.AgentThreads,
	)
}

func getEnlistAgents(agentHandles []executor.TaskHandle) string {
	enlistAgentsString := ""
	for _, agent := range agentHandles {
		enlistAgentsString += fmt.Sprintf(" -a %s", agent.Address())
	}
	return enlistAgentsString
}

func (m mutilate) getMasterQPSOption() string {
	masterQPSOption := ""

	if m.config.MasterQPS != 0 {
		masterQPSOption = fmt.Sprintf(" -Q %d", m.config.MasterQPS)
	}
	return masterQPSOption
}

func (m mutilate) getPopulateCommand() string {
	return fmt.Sprintf("%s -s %s:%d --loadonly",
		m.config.PathToBinary,
		m.config.MemcachedHost,
		m.config.MemcachedPort,
	)
}

func (m mutilate) getBaseMasterMutilateCommand(handles []executor.TaskHandle) string {
	return fmt.Sprint(
		fmt.Sprintf("%s", m.config.PathToBinary),
		fmt.Sprintf(" -s %s:%d", m.config.MemcachedHost, m.config.MemcachedPort),
		fmt.Sprintf(" --warmup %d --noload ", int(m.config.WarmupTime.Seconds())),
		fmt.Sprintf(" -K %d -V %d", m.config.KeySize, m.config.ValueSize),
		fmt.Sprintf(" -T %d", m.config.MasterThreads),
		fmt.Sprintf(" -D %d -C %d", m.config.MasterConnectionsDepth, m.config.MasterConnections),
		fmt.Sprintf(" -d %d -c %d", m.config.AgentConnectionsDepth, m.config.AgentConnections),
		fmt.Sprintf("%s %s", m.getMasterQPSOption(), getEnlistAgents(handles)),
	)
}

func (m mutilate) getLoadCommand(qps int, duration time.Duration, agentHandles []executor.TaskHandle) string {
	baseCommand := m.getBaseMasterMutilateCommand(agentHandles)
	return fmt.Sprintf("%s -q %d -t %d --swanpercentile %s",
		baseCommand, qps, int(duration.Seconds()), m.config.LatencyPercentile.String())
}

func (m mutilate) getTuneCommand(slo int, agentHandles []executor.TaskHandle) (command string) {
	baseCommand := m.getBaseMasterMutilateCommand(agentHandles)
	command = fmt.Sprintf("%s --search %s:%d -t %d",
		baseCommand, m.config.LatencyPercentile.String(), slo, int(m.config.TuningTime.Seconds()))
	return command
}
