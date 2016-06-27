package mutilate

import (
	"fmt"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
)

func (s *MutilateTestSuite) TestGetPopulateCommand() {
	command := getPopulateCommand(s.mutilate.config)

	Convey("Mutilate populate command should contain mutilate binary path", s.T(), func() {
		expected := fmt.Sprintf("%s", s.mutilate.config.PathToBinary)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate populate command should contain target server host:port", s.T(), func() {
		expected := fmt.Sprintf(
			"-s %s:%d", s.mutilate.config.MemcachedHost, s.mutilate.config.MemcachedPort)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate populate command should contain --loadonly switch", s.T(), func() {
		expected := fmt.Sprintf("--loadonly")
		So(command, ShouldContainSubstring, expected)
	})
}

func (s *MutilateTestSuite) soExpectBaseCommandOptions(command string) {
	Convey("Mutilate base command should contain mutilate binary path", s.T(), func() {
		expected := fmt.Sprintf("%s", s.mutilate.config.PathToBinary)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain target server host:port", s.T(), func() {
		expected := fmt.Sprintf(
			"-s %s:%d", s.mutilate.config.MemcachedHost, s.mutilate.config.MemcachedPort)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain warmup", s.T(), func() {
		expected := fmt.Sprintf("--warmup %d", int(s.config.WarmupTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain noload and -B option", s.T(), func() {
		expected := fmt.Sprintf("--noload")
		So(command, ShouldContainSubstring, expected)
		expected = fmt.Sprintf("-B")
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain keySize and valuSize option", s.T(), func() {
		expected := fmt.Sprintf("-K %d", s.mutilate.config.KeySize)
		So(command, ShouldContainSubstring, expected)
		expected = fmt.Sprintf("-V %d", s.mutilate.config.ValueSize)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain master threads option", s.T(), func() {
		expected := fmt.Sprintf("-T %d", s.mutilate.config.MasterThreads)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain agents connection options", s.T(), func() {
		expected := fmt.Sprintf("-d %d", s.mutilate.config.AgentConnectionsDepth)
		So(command, ShouldContainSubstring, expected)
		expected = fmt.Sprintf("-c %d", s.mutilate.config.AgentConnections)
		So(command, ShouldContainSubstring, expected)
	})

	if s.mutilate.config.MasterQPS != 0 {
		Convey("Mutilate base command should contain masterQPS option", s.T(), func() {
			expected := fmt.Sprintf("-Q %d", s.mutilate.config.MasterQPS)
			So(command, ShouldContainSubstring, expected)
		})
	}
}

func (s *MutilateTestSuite) TestGetLoadCommand() {
	const load = 300
	const duration = 10 * time.Second

	s.mutilate.config.MasterQPS = 0
	command := getLoadCommand(s.mutilate.config, load, duration, []executor.TaskHandle{})

	s.soExpectBaseCommandOptions(command)

	Convey("Mutilate load command should contain load duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(duration.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain swan percentile option", s.T(), func() {
		expected := fmt.Sprintf("--swanpercentile %s", s.mutilate.config.LatencyPercentile.String())
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain qps option", s.T(), func() {
		expected := fmt.Sprintf("-q %d", load)
		So(command, ShouldContainSubstring, expected)
	})
}

func (s *MutilateTestSuite) TestGetMultinodeLoadCommand() {
	const load = 300
	const duration = 10 * time.Second
	const agentAddress1 = "255.255.255.001"
	const agentAddress2 = "255.255.255.002"

	s.mutilate.config.MasterQPS = 0

	s.mAgentHandle1.On("Address").Return(agentAddress1).Once()
	s.mAgentHandle2.On("Address").Return(agentAddress2).Once()
	command := getLoadCommand(s.mutilate.config, load, duration, []executor.TaskHandle{
		s.mAgentHandle1,
		s.mAgentHandle2,
	})

	s.soExpectBaseCommandOptions(command)

	Convey("Mutilate load command should contain load duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(duration.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain swan percentile option", s.T(), func() {
		expected := fmt.Sprintf("--swanpercentile %s", s.mutilate.config.LatencyPercentile.String())
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain qps option", s.T(), func() {
		expected := fmt.Sprintf("-q %d", load)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain agents", s.T(), func() {
		expected := fmt.Sprintf("-a %s -a %s", agentAddress1, agentAddress2)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate base command should contain master connection options", s.T(), func() {
		expected := fmt.Sprintf("-D %d", s.mutilate.config.MasterConnectionsDepth)
		So(command, ShouldContainSubstring, expected)
		expected = fmt.Sprintf("-C %d", s.mutilate.config.MasterConnections)
		So(command, ShouldContainSubstring, expected)
	})

	// Check with MasterQPS different to 0.
	s.mutilate.config.MasterQPS = 24234

	s.mAgentHandle1.On("Address").Return(agentAddress1).Once()
	s.mAgentHandle2.On("Address").Return(agentAddress2).Once()
	command = getLoadCommand(s.mutilate.config, load, duration, []executor.TaskHandle{
		s.mAgentHandle1,
		s.mAgentHandle2,
	})

	s.soExpectBaseCommandOptions(command)

	Convey("Assert expectation should be met", s.T(), func() {
		So(s.mAgentHandle1.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mAgentHandle2.AssertExpectations(s.T()), ShouldBeTrue)
	})
}

func (s *MutilateTestSuite) TestGetTuneCommand() {
	const slo = 300

	s.mutilate.config.MasterQPS = 0
	command := getTuneCommand(s.mutilate.config, slo, []executor.TaskHandle{})

	s.soExpectBaseCommandOptions(command)

	Convey("Mutilate tuning command should contain tuning phase duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(s.mutilate.config.TuningTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate tuning command should contain search option", s.T(), func() {
		expected := fmt.Sprintf("--search %s:%d",
			s.mutilate.config.LatencyPercentile.String(), slo)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Assert expectation should be met", s.T(), func() {
		So(s.mAgentHandle1.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mAgentHandle2.AssertExpectations(s.T()), ShouldBeTrue)
	})
}
func (s *MutilateTestSuite) TestGetMultinodeTuneCommand() {
	const slo = 300
	const agentAddress1 = "255.255.255.001"
	const agentAddress2 = "255.255.255.002"

	s.mutilate.config.MasterQPS = 0

	s.mAgentHandle1.On("Address").Return(agentAddress1).Once()
	s.mAgentHandle2.On("Address").Return(agentAddress2).Once()
	command := getTuneCommand(s.mutilate.config, slo, []executor.TaskHandle{
		s.mAgentHandle1,
		s.mAgentHandle2,
	})

	s.soExpectBaseCommandOptions(command)

	Convey("Mutilate tuning command should contain tuning phase duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(s.mutilate.config.TuningTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate tuning command should contain search option", s.T(), func() {
		expected := fmt.Sprintf("--search %s:%d",
			s.mutilate.config.LatencyPercentile.String(), slo)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate tuning command should contain agents", s.T(), func() {
		expected := fmt.Sprintf("-a %s -a %s", agentAddress1, agentAddress2)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate tuning command should contain port options", s.T(), func() {
		expected := fmt.Sprintf("-p %d", s.mutilate.config.AgentPort)
		So(command, ShouldContainSubstring, expected)
	})

	// Check with MasterQPS different to 0.
	s.mutilate.config.MasterQPS = 24234

	s.mAgentHandle1.On("Address").Return(agentAddress1).Once()
	s.mAgentHandle2.On("Address").Return(agentAddress2).Once()
	command = getTuneCommand(s.mutilate.config, slo, []executor.TaskHandle{
		s.mAgentHandle1,
		s.mAgentHandle2,
	})

	s.soExpectBaseCommandOptions(command)

	Convey("Mutilate tuning command should contain tuning phase duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(s.mutilate.config.TuningTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate tuning command should contain search option", s.T(), func() {
		expected := fmt.Sprintf("--search %s:%d",
			s.mutilate.config.LatencyPercentile.String(), slo)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate tuning command should contain agents", s.T(), func() {
		expected := fmt.Sprintf("-a %s -a %s", agentAddress1, agentAddress2)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Assert expectation should be met", s.T(), func() {
		So(s.mAgentHandle1.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mAgentHandle2.AssertExpectations(s.T()), ShouldBeTrue)
	})
}
