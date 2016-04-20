package workloads

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

const (
	mutilatePath  = "/usr/bin/local/mutilate"
	memcachedHost = "127.0.0.1"

	correctMutilateQPS    = 4450
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

	mExecutor *mocks.Executor
	mHandle   *mocks.Task
}

func (s *MutilateTestSuite) SetupTest() {
	s.config.mutilatePath = mutilatePath
	s.config.memcachedHost = memcachedHost
	s.config.latencyPercentile = 99.9
	s.config.tuningTime = 30 * time.Second

	s.defaultSlo = 1000

	s.mExecutor = new(mocks.Executor)
	s.mHandle = new(mocks.Task)
}

func (s *MutilateTestSuite) TestTuneMutilate() {
	mutilateTuneCommand := fmt.Sprintf("%s -s %s --search=%d:%d -t %d",
		s.config.mutilatePath,
		s.config.memcachedHost,
		999, // Latency Percentile translated to "Mutilate int"
		s.defaultSlo,
		int(s.config.tuningTime.Seconds()),
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

	Convey("When Tuning Memcached.", s.T(), func() {
		targetQPS, err := mutilate.Tune(s.defaultSlo)
		Convey("On success, error should be nil.", func() {
			So(err, ShouldBeNil)
		})
		Convey("TargetQPS should be correct.", func() {
			So(targetQPS, ShouldEqual, correctMutilateQPS)
		})
		Convey("Mock expectations are asserted.", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func (s *MutilateTestSuite) TestMutilateTuneExecutorError() {
	mutilate := NewMutilate(s.mExecutor, s.config)
	const errorMsg = "Error in execution"
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).
		Return(s.mHandle, errors.New(errorMsg))

	Convey("When executing Tune, and executor retuns error.", s.T(), func() {
		_, err := mutilate.Tune(s.defaultSlo)
		Convey("Error should not be nil.", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should contain \"Error in execution\".", func() {
			So(err.Error(), ShouldContainSubstring, errorMsg)
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

	loadCmd := fmt.Sprintf("%s -s %s -q %d -t %d --swanpercentile=%s",
		s.config.mutilatePath,
		s.config.memcachedHost,
		load,
		int(duration.Seconds()),
		percentile,
	)

	mutilate := NewMutilate(s.mExecutor, s.config)

	s.mExecutor.On("Execute", loadCmd).Return(s.mHandle, nil)
	s.mHandle.On("Wait", 0).Return(true)
	s.mHandle.On("Status").Return(executor.TERMINATED, &executorStatus)

	Convey("When generating Load.", s.T(), func() {
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

func (s *MutilateTestSuite) TestMutilateLoadExecutorError() {
	mutilate := NewMutilate(s.mExecutor, s.config)
	const errorMsg = "Error in execution"
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).
		Return(s.mHandle, errors.New(errorMsg))

	Convey("When executing Load and executor retuns error.", s.T(), func() {
		_, err := mutilate.Load(20, 1*time.Second)
		Convey("Error should not be nil.", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should contain \"Error in execution\".", func() {
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})
	})
}

func (s *MutilateTestSuite) TestPopulate() {
	mutilatePopulateCommand := fmt.Sprintf("%s -s %s --loadonly",
		s.config.mutilatePath,
		s.config.memcachedHost,
	)

	executorStatus := executor.Status{
		ExitCode: 0,
		Stdout:   "",
		Stderr:   "",
	}

	mutilate := NewMutilate(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mutilatePopulateCommand).Return(s.mHandle, nil)
	s.mHandle.On("Wait", 0).Return(true)
	s.mHandle.On("Status").Return(executor.TERMINATED, &executorStatus)

	Convey("When Populating Memcached.", s.T(), func() {
		err := mutilate.Populate()
		Convey("On success, error should be nil.", func() {
			So(err, ShouldBeNil)
		})

		Convey("Mock expectations are asserted.", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func TestMutilateTestSuite(t *testing.T) {
	suite.Run(t, new(MutilateTestSuite))
}
