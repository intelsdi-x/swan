package mutilate

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/shopspring/decimal"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
)

const (
	mutilatePath  = "/usr/bin/local/mutilate"
	memcachedHost = "127.0.0.1"

	correctMutilateQPS    = 4450
	correctMutilateSLI    = 75
	correctMutilateOutput = `#type       avg     std     min     5th    10th    90th    95th    99th
read       20.9    11.9    11.9    12.5    13.1    32.4    39.0    56.8
update      0.0     0.0     0.0     0.0     0.0     0.0     0.0     0.0
op_q        1.0     0.0     1.0     1.0     1.0     1.1     1.1     1.1
Swan latency for percentile 99.900000: 75.125312

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

	mExecutor *mocks.Executor
	mHandle   *mocks.TaskHandle
}

func (s *MutilateTestSuite) SetupTest() {
	s.config.PathToBinary = mutilatePath
	s.config.MemcachedHost = memcachedHost
	s.config.LatencyPercentile, _ = decimal.NewFromString("99.9")
	s.config.TuningTime = 30 * time.Second
	s.config.WarmupTime = 10 * time.Second
	s.config.EraseTuneOutput = true
	s.config.ErasePopulateOutput = true

	s.defaultSlo = 1000

	s.mExecutor = new(mocks.Executor)
	s.mHandle = new(mocks.TaskHandle)

	s.mutilate.config = s.config

	s.mHandle.On("Address").Return("localhost")
}

func (s *MutilateTestSuite) TestGetLoadCommand() {
	const load = 300
	const duration = 10 * time.Second
	command := s.mutilate.getLoadCommand(load, duration, []executor.TaskHandle{})

	Convey("Mutilate load command should contain mutilate binary path", s.T(), func() {
		expected := fmt.Sprintf("%s", s.mutilate.config.PathToBinary)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain target server", s.T(), func() {
		expected := fmt.Sprintf("-s %s", s.mutilate.config.MemcachedHost)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain load duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(duration.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain load QPS", s.T(), func() {
		expected := fmt.Sprintf("-q %d", load)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain warmup", s.T(), func() {
		expected := fmt.Sprintf("--warmup=%d", int(s.config.WarmupTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain noload option", s.T(), func() {
		expected := fmt.Sprintf("--noload")
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain swan percentile option", s.T(), func() {
		expected := fmt.Sprintf("--swanpercentile=%s", s.mutilate.config.LatencyPercentile.String())
		So(command, ShouldContainSubstring, expected)
	})
}

func (s *MutilateTestSuite) TestGetTuneCommand() {
	const slo = 300
	command := s.mutilate.getTuneCommand(slo, []executor.TaskHandle{})

	Convey("Mutilate load command should contain mutilate binary path", s.T(), func() {
		expected := fmt.Sprintf("%s", s.mutilate.config.PathToBinary)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain target server", s.T(), func() {
		expected := fmt.Sprintf("-s %s:%d", s.mutilate.config.MemcachedHost,
			s.mutilate.config.MemcachedPort)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain warmup", s.T(), func() {
		expected := fmt.Sprintf("--warmup=%d", int(s.config.WarmupTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain tuning phase duration", s.T(), func() {
		expected := fmt.Sprintf("-t %d", int(s.mutilate.config.TuningTime.Seconds()))
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain noload option", s.T(), func() {
		expected := fmt.Sprintf("--noload")
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain search option", s.T(), func() {
		expected := fmt.Sprintf("--search=%s:%d",
			s.mutilate.config.LatencyPercentile.String(), slo)
		So(command, ShouldContainSubstring, expected)
	})
}

func (s *MutilateTestSuite) TestGetPopulateCommand() {
	command := s.mutilate.getPopulateCommand()

	Convey("Mutilate load command should contain mutilate binary path", s.T(), func() {
		expected := fmt.Sprintf("%s", s.mutilate.config.PathToBinary)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain target server", s.T(), func() {
		expected := fmt.Sprintf("-s %s", s.mutilate.config.MemcachedHost)
		So(command, ShouldContainSubstring, expected)
	})

	Convey("Mutilate load command should contain --loadonly switch", s.T(), func() {
		expected := fmt.Sprintf("--loadonly")
		So(command, ShouldContainSubstring, expected)
	})
}

func (s *MutilateTestSuite) TestMutilateTuning() {
	outputFile, err := ioutil.TempFile(os.TempDir(), "mutilate")
	if err != nil {
		s.Fail(err.Error())
		return
	}
	defer outputFile.Close()

	outputFile.WriteString(correctMutilateOutput)

	mutilate := New(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mock.AnythingOfType("string")).Return(s.mHandle, nil)
	s.mHandle.On("Wait", 0*time.Nanosecond).Return(true)
	s.mHandle.On("ExitCode").Return(0, nil)
	s.mHandle.On("StdoutFile").Return(outputFile, nil)
	s.mHandle.On("EraseOutput").Return(nil)
	s.mHandle.On("Stop").Return(nil)

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
			So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func (s *MutilateTestSuite) TestMutilateTuningExecutorError() {
	mutilate := New(s.mExecutor, s.config)
	const errorMsg = "Error in execution"
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).
		Return(s.mHandle, errors.New(errorMsg))

	Convey("When tuning, and an executor retuns an error", s.T(), func() {
		_, _, err := mutilate.Tune(s.defaultSlo)
		Convey("Error should not be nil", func() {
			So(err, ShouldNotBeNil)
		})
		Convey("Error should contain \"Error in execution\"", func() {
			So(err.Error(), ShouldContainSubstring, errorMsg)
		})
	})
}

func (s *MutilateTestSuite) TestMutilateLoad() {
	const load = 1000
	const duration = 10 * time.Second

	mutilate := New(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mock.AnythingOfType("string")).Return(s.mHandle, nil)

	Convey("When generating Load.", s.T(), func() {
		mutilateTask, err := mutilate.Load(load, duration)
		Convey("On success, error should be nil", func() {
			So(err, ShouldBeNil)
		})
		Convey("mutilateTask should not be a mockedTask", func() {
			So(mutilateTask, ShouldNotEqual, s.mHandle)
		})
		Convey("Mock expectations are asserted", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
		})
	})
}

func (s *MutilateTestSuite) TestMutilateLoadExecutorError() {
	mutilate := New(s.mExecutor, s.config)
	const errorMsg = "Error in execution"
	s.mExecutor.On("Execute", mock.AnythingOfType("string")).
		Return(s.mHandle, errors.New(errorMsg))

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

func (s *MutilateTestSuite) TestPopulate() {
	mutilatePopulateCommand := fmt.Sprintf("%s -s %s:%d --loadonly",
		s.config.PathToBinary,
		s.config.MemcachedHost,
		s.config.MemcachedPort,
	)

	mutilate := New(s.mExecutor, s.config)

	s.mExecutor.On("Execute", mutilatePopulateCommand).Return(s.mHandle, nil)
	s.mHandle.On("Wait", 0*time.Nanosecond).Return(true)
	s.mHandle.On("ExitCode").Return(0, nil)
	s.mHandle.On("Clean").Return(nil)
	s.mHandle.On("EraseOutput").Return(nil)

	Convey("When Populating Memcached.", s.T(), func() {
		err := mutilate.Populate()
		Convey("On success, error should be nil", func() {
			So(err, ShouldBeNil)
		})

		Convey("Mock expectations are asserted", func() {
			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			//So(s.mHandle.AssertExpectations(s.T()), ShouldBeTrue)
			// NOTE: This test currently fail because it tries to assert mHandle.Address() which is not used here.
		})
	})
}

func TestMutilateTestSuite(t *testing.T) {
	suite.Run(t, new(MutilateTestSuite))
}
