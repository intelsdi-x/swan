package workloads

import (
	//"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	//"time"
	//"errors"
	"fmt"
	"time"
)

type mockedTaskHandle struct {
	mock.Mock
}

func (m *mockedTaskHandle) Stop() error {
	args := m.Called()
	return args.Error(0)
}

func (m *mockedTaskHandle) Status() (executor.TaskState, *executor.Status) {
	args := m.Called()
	return args.Get(0).(executor.TaskState),
		args.Get(1).(*executor.Status)
}

func (m *mockedTaskHandle) Wait(timeoutMs int) bool {
	args := m.Called(timeoutMs)
	return args.Bool(0)
}

type mockedExecutor struct {
	mock.Mock
}

func (m *mockedExecutor) Execute(command string) (executor.Task, error) {
	args := m.Called(command)
	return args.Get(0).(executor.Task), args.Error(1)
}

const (
	mutilate_path = "/usr/bin/local/mutilate"
	memcached_uri = "127.0.0.1"
	//latency_percentile = 99.999

	correctMutilateQps    = 4450
	correctMutilateOutput = `#type       avg     std     min     5th    10th    90th    95th    99th
read       20.9    11.9    11.9    12.5    13.1    32.4    39.0    56.8
update      0.0     0.0     0.0     0.0     0.0     0.0     0.0     0.0
op_q        1.0     0.0     1.0     1.0     1.0     1.1     1.1     1.1
Swan latency for percentile 99.980000: 3776.125312

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
	mutilate Mutilate
	config   MutilateConfig

	defaultMutilateCommand string
	defaultSlo             int

	mExecutor *mockedExecutor
	mHandle   *mockedTaskHandle
}

func (suite *MutilateTestSuite) SetupTest() {
	suite.config.mutilate_path = mutilate_path
	suite.config.memcached_uri = memcached_uri
	suite.config.latency_percentile = 99.9
	suite.config.tuning_time = 30 * time.Second

	suite.defaultSlo = 1000

	suite.mExecutor = new(mockedExecutor)
	suite.mHandle = new(mockedTaskHandle)
}

//func (s *MutilateTestSuite) TestMutilateTuneExecutorError() {
//mutilate := NewMutilate(s.mExecutor, s.config)
////err := errors.New("asa")
////_ = err
//
//s.mExecutor.On("Execute", s.defaultMutilateCommand).Return(s.mHandle, errors.New("error"))
//s.mExecutor.On("Execute", s.defaultMutilateCommand).Return(s.mHandle, err)
//
//s.mHandle.On("Wait", 0).Return(true)
//s.mHandle.On("Status").Return(executor.TERMINATED, &executorStatus)
//
//Convey("When Tuning Mutilate, and executor retuns error.", s.T(), func() {
//	_, err := mutilate.Tune(s.defaultSlo)
//	Convey("Error should not be nil.", func() {
//		So(err, ShouldNotBeNil)
//	})
//})
//_ = mutilate
//}

func (s *MutilateTestSuite) TestTuneMutilate() {
	mutilateTuneCommand := fmt.Sprintf("%s -s %s --search=%d:%d -t %d",
		s.config.mutilate_path,
		s.config.memcached_uri,
		999, // Latency Percentile translated to "Mutilate int"
		s.defaultSlo,
		int(s.config.tuning_time.Seconds()),
	)

	executorStatus := executor.Status{
		ExitCode: 0,
		Stdout:   correctMutilateOutput,
		Stderr:   "",
	}

	mutilate := NewMutilate(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mutilateTuneCommand).Return(s.mHandle, nil)
	s.mHandle.On("Wait", 0).Return(true)
	s.mHandle.On("Status").Return(executor.TERMINATED, &executorStatus)

	Convey("When Tuning Mutilate to Memcached.", s.T(), func() {
		targetQps, err := mutilate.Tune(s.defaultSlo)
		Convey("On success, error should be nil.", func() {
			So(err, ShouldBeNil)
		})
		Convey("TargetQPS should be correct.", func() {
			So(targetQps, ShouldEqual, correctMutilateQps)
		})
		Convey("Mock expectations are asserted.", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func (s *MutilateTestSuite) TestMutilateLoad() {
	const load = 1000
	const duration = 10 * time.Second
	const percentile = "99.9"
	const expectedLatency = 3776

	executorStatus := executor.Status{
		ExitCode: 0,
		Stdout:   correctMutilateOutput,
		Stderr:   "",
	}

	loadCmd := fmt.Sprintf("%s -s %s -q %d -t %d --swanpercentile=%f",
		s.config.mutilate_path,
		s.config.memcached_uri,
		load,
		duration.Seconds(),
		percentile,
	)

	mutilate := NewMutilate(s.mExecutor, s.config)

	s.mExecutor.On("Execute", loadCmd).Return(s.mHandle, nil)
	s.mHandle.On("Wait", 0).Return(true)
	s.mHandle.On("Status").Return(executor.TERMINATED, &executorStatus)

	Convey("When Load Generating through Mutilate..", s.T(), func() {
		sli, err := mutilate.Load(load, duration)
		Convey("On success, error should be nil.", func() {
			So(err, ShouldBeNil)
		})
		Convey("SLI should be correct.", func() {
			So(sli, ShouldEqual, expectedLatency)
		})
		Convey("Mock expectations are asserted.", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func TestMutilateTestSuite(t *testing.T) {
	suite.Run(t, new(MutilateTestSuite))
	assert.True(t, true)
}

//func TestGetTuningOutput(t *testing.T) {
//	const correctTargetQps = 4450
//
//	Convey("When given proper Mutilate tuning output: ", t, func() {
//		targetQps, err := getQpsFromOutput(correctMutilateOutput)
//
//		Convey("we receive proper targetQPS.", func() {
//			So(err, ShouldBeNil)
//		})
//
//		Convey("we receive nil error.", func() {
//			So(targetQps, ShouldEqual, correctTargetQps)
//		})
//	})
//
//	Convey("When given inproper Mutilate tuning output we receive error.", t, func() {
//		_, err := getQpsFromOutput("Inproper Output")
//		So(err, ShouldNotBeNil)
//	})
//}

//func TestGetLatenciesFromMutilateOutput(t *testing.T) {
//	const mutilateOutput = `1.041759 12.874603
//1.041772 13.113022
//1.041786 15.113022
//`
//	properLatencies := []int{12, 13, 15}
//	Convey("When given proper Mutilate tuning output: ", t, func() {
//		latencies := getLatenciesFromMutilateOutput(mutilateOutput)
//		Convey("So latencies are correct.", func() {
//			So(latencies, ShouldEqual, properLatencies)
//		})
//	})
//}

//func TestGetLatenciesFromMutilateOutput(t *testing.T) {
//	const mutilateOutput = `1.041759 12.874603
//1.041772 13.113022
//1.041786 13.113022
//1.041799 13.828278
//1.041813 13.113022
//1.041827 13.113022
//1.041840 12.874603
//1.041853 14.066696
//1.041867 12.874603
//1.041881 12.874603
//`
//
//}

//func TestTuneMutilate(t *testing.T) {
//	localExecutor := executor.NewLocal()
//
//	conf := DefaultMutilateConfig()
//
//	conf.latency_percentile = 50
//
//	const slo = 1000
//
//	mutilate := NewMutilate(localExecutor, conf)
//
//	targetQps, err := mutilate.Tune(slo)
//
//	SkipConvey("When Tuning Mutilate to Memcached.", t, func() {
//		Convey("Error should be nil.", func() {
//			So(err, ShouldBeNil)
//		})
//		Convey("TargetQPS should be more than 0.", func() {
//			So(targetQps, ShouldNotEqual, 0)
//		})
//	})
//}
