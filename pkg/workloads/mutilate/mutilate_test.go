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
	"errors"
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
)

const (
	mutilatePath           = "/usr/bin/local/mutilate"
	memcachedHost          = "123.123.22.34"
	memcachedPort          = 2145
	agentThreads           = 24
	agentConnections       = 9
	agentConnectionsDepth  = 2
	masterThreads          = 4
	masterConnections      = 23
	masterConnectionsDepth = 1
	keySize                = "3"
	valueSize              = "5"
	intearrivaldist        = "fb_ia"
	latencyPercentile      = "99.9234"

	correctMutilateQPS    = 4450
	correctMutilateSLI    = 56
	correctMutilateOutput = `#type       avg     std     min     5th    10th    90th    95th    99th
read       20.9    11.9    11.9    12.5    13.1    32.4    39.0    56.8
update      0.0     0.0     0.0     0.0     0.0     0.0     0.0     0.0
op_q        1.0     0.0     1.0     1.0     1.0     1.1     1.1     1.1

Total QPS = 4450.3 (89007 / 20.0s)
Peak QPS  = 71164.8

Misses = 0 (0.0%)
Skipped TXs = 0 (0.0%)

RX   22044729 bytes :    1.1 MB/s
TX    3204252 bytes :    0.2 MB/s
`
)

type MutilateTestSuite struct {
	suite.Suite
	mutilate mutilate
	config   Config

	defaultMutilateCommand string
	defaultSlo             int

	mExecutor          *executor.MockExecutor
	mExecutorForAgent1 *executor.MockExecutor
	mExecutorForAgent2 *executor.MockExecutor
	mMasterHandle      *executor.MockTaskHandle
	mAgentHandle1      *executor.MockTaskHandle
	mAgentHandle2      *executor.MockTaskHandle
}

func (s *MutilateTestSuite) SetupTest() {
	s.defaultSlo = 1000

	// Mutilate Configuration.
	s.config.PathToBinary = mutilatePath
	s.config.MemcachedHost = memcachedHost
	s.config.MemcachedPort = memcachedPort
	s.config.LatencyPercentile = latencyPercentile
	s.config.TuningTime = 934 * time.Second
	s.config.WarmupTime = 1231 * time.Second
	s.config.AgentThreads = agentThreads
	s.config.AgentConnections = agentConnections
	s.config.AgentConnectionsDepth = agentConnectionsDepth
	s.config.MasterThreads = masterThreads
	s.config.MasterConnections = masterConnections
	s.config.MasterConnectionsDepth = masterConnectionsDepth
	s.config.KeySize = keySize
	s.config.ValueSize = valueSize
	s.config.InterArrivalDist = intearrivaldist
	s.mutilate.config = s.config

	s.mAgentHandle1 = new(executor.MockTaskHandle)
	s.mAgentHandle2 = new(executor.MockTaskHandle)

	s.mMasterHandle = new(executor.MockTaskHandle)
	s.mExecutor = new(executor.MockExecutor)
	s.mExecutorForAgent1 = new(executor.MockExecutor)
	s.mExecutorForAgent2 = new(executor.MockExecutor)

	// Don't want to have not erased output after tests.
	s.config.EraseTuneOutput = true
	s.config.ErasePopulateOutput = true
}

// Testing successful master-only mutilate tuning case.
func (s *MutilateTestSuite) TestMutilateTuning() {
	outputFile, err := ioutil.TempFile(os.TempDir(), "mutilate")
	if err != nil {
		s.Fail(err.Error())
		return
	}
	defer outputFile.Close()

	outputFile.WriteString(correctMutilateOutput)

	mutilate := New(s.mExecutor, s.config)

	// This is needed to know how many times the mock should expect function execution.
	numberOfConveys := 4
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).Return(s.mMasterHandle, nil)
	s.mMasterHandle.On("Wait", 0*time.Nanosecond).Return(true, nil).Times(numberOfConveys)
	s.mMasterHandle.On("ExitCode").Return(0, nil).Times(numberOfConveys)
	s.mMasterHandle.On("StdoutFile").Return(outputFile, nil).Times(numberOfConveys)
	s.mMasterHandle.On("EraseOutput").Return(nil).Times(numberOfConveys)

	Convey("When Tuning Memcached.", s.T(), func() {
		outputFile.Seek(0, 0)
		targetQPS, sli, err := mutilate.Tune(s.defaultSlo)
		Convey("On success, error should be nil.", func() {
			So(err, ShouldBeNil)
		})
		Convey("TargetQPS should be correct.", func() {
			So(targetQPS, ShouldEqual, correctMutilateQPS)
		})

		Convey("SLI should be correct.", func() {
			So(sli, ShouldEqual, correctMutilateSLI)
		})

		Convey("Mock expectations are asserted.", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mMasterHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

// Testing successful clustered mutilate tuning case.
func (s *MutilateTestSuite) TestClusterMutilateTuning() {
	outputFile, err := ioutil.TempFile(os.TempDir(), "mutilate")
	if err != nil {
		s.Fail(err.Error())
		return
	}
	defer outputFile.Close()

	outputFile.WriteString(correctMutilateOutput)

	mutilate := NewCluster(s.mExecutor, []executor.Executor{
		s.mExecutorForAgent1,
		s.mExecutorForAgent2,
	}, s.config)

	// This is needed to know how many times the mock should expect function execution.
	numberOfConveys := 4
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).Return(s.mMasterHandle, nil)
	s.mMasterHandle.On("Wait", 0*time.Nanosecond).Return(true, nil).Times(numberOfConveys)
	s.mMasterHandle.On("ExitCode").Return(0, nil).Times(numberOfConveys)
	s.mMasterHandle.On("StdoutFile").Return(outputFile, nil).Times(numberOfConveys)
	s.mMasterHandle.On("EraseOutput").Return(nil).Times(numberOfConveys)

	s.mExecutorForAgent1.On("Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle1, nil)
	s.mAgentHandle1.On("Address").Return("255.255.255.001").Times(numberOfConveys)
	s.mAgentHandle1.On("Status").Return(executor.RUNNING)
	// Those function shouldn't be called in normal execution path
	s.mAgentHandle1.On("Stop").Return(nil).Times(0)
	s.mAgentHandle1.On("EraseOutput").Return(nil).Times(0)

	s.mExecutorForAgent2.On("Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle2, nil)
	s.mAgentHandle2.On("Address").Return("255.255.255.002").Times(numberOfConveys)
	s.mAgentHandle2.On("Status").Return(executor.RUNNING)
	// Those function shouldn't be called in normal execution path
	s.mAgentHandle2.On("Stop").Return(nil).Times(0)
	s.mAgentHandle2.On("EraseOutput").Return(nil).Times(0)

	Convey("When Tuning Memcached.", s.T(), func() {
		outputFile.Seek(0, 0)
		targetQPS, sli, err := mutilate.Tune(s.defaultSlo)
		Convey("On success, error should be nil.", func() {
			So(err, ShouldBeNil)
		})
		Convey("TargetQPS should be correct.", func() {
			So(targetQPS, ShouldEqual, correctMutilateQPS)
		})

		Convey("SLI should be correct.", func() {
			So(sli, ShouldEqual, correctMutilateSLI)
		})

		Convey("Mock expectations are asserted.", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mExecutorForAgent1.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mExecutorForAgent2.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mMasterHandle.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mAgentHandle1.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mAgentHandle2.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

// Testing master-only mutilate tuning with master executor failure.
func (s *MutilateTestSuite) TestMutilateTuningExecutorError() {
	mutilate := New(s.mExecutor, s.config)
	const errorMsg = "Error in execution"
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).
		Return(nil, errors.New(errorMsg))

	Convey("When tuning, and an executor returns an error", s.T(), func() {
		_, _, err := mutilate.Tune(s.defaultSlo)
		Convey("Error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should contain \"Error in execution\"", func() {
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})
	})
}

// Testing clustered mutilate tuning with different failures.
func (s *MutilateTestSuite) TestClusterMutilateTuningErrors() {
	Convey("While having clustered mutilate", s.T(), func() {
		mutilate := NewCluster(s.mExecutor, []executor.Executor{
			s.mExecutorForAgent1,
			s.mExecutorForAgent2,
		}, s.config)

		Convey("Having failure with master executor, tune should return error and agents should be stopped", func() {
			const errorMsg = "Error in execution"

			s.mExecutor.On("Execute", mock.AnythingOfType("string")).
				Return(nil, errors.New(errorMsg)).Once()
			s.mExecutorForAgent1.On(
				"Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle1, nil).Once()
			s.mAgentHandle1.On("Address").Return("255.255.255.001").Once()
			s.mAgentHandle1.On("Stop").Return(nil).Once()
			s.mAgentHandle1.On("EraseOutput").Return(nil).Once()
			s.mAgentHandle1.On("Status").Return(executor.RUNNING)

			s.mExecutorForAgent2.On(
				"Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle2, nil).Once()
			s.mAgentHandle2.On("Address").Return("255.255.255.002").Once()
			s.mAgentHandle2.On("Stop").Return(nil).Once()
			s.mAgentHandle2.On("EraseOutput").Return(nil).Once()
			s.mAgentHandle2.On("Status").Return(executor.RUNNING)

			_, _, err := mutilate.Tune(s.defaultSlo)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})

		Convey("Having failure with one agent executor, tune should return error"+
			", agents should be stopped and master not started", func() {
			const errorMsg = "Error in execution"

			s.mExecutorForAgent1.On(
				"Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle1, nil).Once()
			s.mAgentHandle1.On("Stop").Return(nil).Once()
			s.mAgentHandle1.On("EraseOutput").Return(nil).Once()

			s.mExecutorForAgent2.On(
				"Execute", mock.AnythingOfType("string")).Return(nil, errors.New(errorMsg)).Once()

			_, _, err := mutilate.Tune(s.defaultSlo)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})

		Convey("Having failure with master's Wait, tune should return error and agents "+
			"should be stopped and output erased", func() {
			const errorMsg = "Mutilate Cluster Failed"

			s.mExecutor.On("Execute", mock.AnythingOfType("string")).Return(s.mMasterHandle, nil).Once()
			// IsTerminated will be false.
			s.mMasterHandle.On("Wait", 0*time.Nanosecond).Return(false, errors.New(errorMsg)).Once()

			s.mExecutorForAgent1.On(
				"Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle1, nil).Once()
			s.mAgentHandle1.On("Address").Return("255.255.255.001").Once()
			s.mAgentHandle1.On("Stop").Return(nil)

			s.mExecutorForAgent2.On(
				"Execute", mock.AnythingOfType("string")).Return(s.mAgentHandle2, nil).Once()
			s.mAgentHandle2.On("Address").Return("255.255.255.002").Once()
			s.mAgentHandle2.On("Stop").Return(nil)

			_, _, err := mutilate.Tune(s.defaultSlo)
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})

		So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mExecutorForAgent1.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mExecutorForAgent2.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mMasterHandle.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mAgentHandle1.AssertExpectations(s.T()), ShouldBeTrue)
		So(s.mAgentHandle2.AssertExpectations(s.T()), ShouldBeTrue)
	})
}

// Testing successful master-only mutilate load case.
// NOTE: No need to test clustered case. It is tested in cluster_task_handle_test.go
func (s *MutilateTestSuite) TestMutilateLoad() {
	const load = 1000
	const duration = 10 * time.Second
	const masterAddress = "255.255.255.001"

	mutilate := New(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mock.AnythingOfType("string")).Return(s.mMasterHandle, nil)
	s.mMasterHandle.On("Address").Return(masterAddress).Twice()

	Convey("When generating Load.", s.T(), func() {
		mutilateTask, err := mutilate.Load(load, duration)
		Convey("On success, error should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("mutilateTask's address should be match masterTask's address", func() {
			So(mutilateTask.Address(), ShouldEqual, s.mMasterHandle.Address())
		})

		Convey("Mock expectations are asserted", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mMasterHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

// Testing master-only mutilate load with excutor failure.
func (s *MutilateTestSuite) TestMutilateLoadExecutorError() {
	mutilate := New(s.mExecutor, s.config)
	const errorMsg = "Error in execution"
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).
		Return(s.mMasterHandle, errors.New(errorMsg))

	Convey("When executing Load and executor returns an error.", s.T(), func() {
		_, err := mutilate.Load(20, 1*time.Second)
		Convey("Error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should contain \"Error in execution\"", func() {
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})
	})
}

// Testing successful mutilate populate case.
func (s *MutilateTestSuite) TestPopulate() {
	mutilatePopulateCommand := fmt.Sprintf("%s -s %s:%d -r 0 --loadonly",
		s.config.PathToBinary,
		s.config.MemcachedHost,
		s.config.MemcachedPort,
	)

	mutilate := New(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mutilatePopulateCommand).Return(s.mMasterHandle, nil)
	s.mMasterHandle.On("Wait", 0*time.Nanosecond).Return(true, nil)
	s.mMasterHandle.On("ExitCode").Return(0, nil)
	s.mMasterHandle.On("EraseOutput").Return(nil)

	Convey("When Populating Memcached.", s.T(), func() {
		err := mutilate.Populate()
		Convey("On success, error should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Mock expectations are asserted", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mMasterHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func TestMutilateTestSuite(t *testing.T) {
	suite.Run(t, new(MutilateTestSuite))
}
