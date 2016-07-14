package sensitivity

import (
	"errors"
	"testing"
	"time"

	"github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"

	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	snapMocks "github.com/intelsdi-x/swan/pkg/snap/mocks"
)

type SensitivityTestSuite struct {
	suite.Suite

	// LC launcher and task.
	mockedLcLauncher *executorMocks.Launcher
	mockedLcTask     *executorMocks.TaskHandle
	// LoadGenerator launcher and task.
	mockedLoadGenerator     *executorMocks.LoadGenerator
	mockedLoadGeneratorTask *executorMocks.TaskHandle
	// Aggressor launcher and task.
	mockedAggressor     *executorMocks.Launcher
	mockedAggressorTask *executorMocks.TaskHandle

	// LC collection Launcher and Handle.
	mockedLcSessionLauncher *snapMocks.SessionLauncher
	mockedLcSessionHandle   *snapMocks.SessionHandle
	// LoadGenerator Launcher and Handle.
	mockedLoadGeneratorSessionLauncher *snapMocks.SessionLauncher
	mockedLoadGeneratorSessionHandle   *snapMocks.SessionHandle
	// Aggressor Launcher and Handle.
	mockedAggressorSessionLauncher *snapMocks.SessionLauncher
	mockedAggressorSessionHandle   *snapMocks.SessionHandle

	configuration         Configuration
	sensitivityExperiment *Experiment

	mockedPeakLoad int
}

func (s *SensitivityTestSuite) SetupTest() {
	// LC launcher and task.
	s.mockedLcLauncher = new(executorMocks.Launcher)
	s.mockedLcTask = new(executorMocks.TaskHandle)
	// LoadGenerator launcher and task.
	s.mockedLoadGenerator = new(executorMocks.LoadGenerator)
	s.mockedLoadGeneratorTask = new(executorMocks.TaskHandle)
	// Aggressor launcher and task.
	s.mockedAggressor = new(executorMocks.Launcher)
	s.mockedAggressorTask = new(executorMocks.TaskHandle)

	// LC collection Launcher and Handle.
	s.mockedLcSessionLauncher = new(snapMocks.SessionLauncher)
	s.mockedLcSessionHandle = new(snapMocks.SessionHandle)
	// LoadGenerator Launcher and Handle.
	s.mockedLoadGeneratorSessionLauncher = new(snapMocks.SessionLauncher)
	s.mockedLoadGeneratorSessionHandle = new(snapMocks.SessionHandle)
	// Aggressor Launcher and Handle.
	s.mockedAggressorSessionLauncher = new(snapMocks.SessionLauncher)
	s.mockedAggressorSessionHandle = new(snapMocks.SessionHandle)

	s.configuration = Configuration{
		SLO:             1,
		LoadDuration:    1 * time.Second,
		LoadPointsCount: 2,
		Repetitions:     1,
		StopOnError:     true,
	}

	s.mockedPeakLoad = 2
}

func (s *SensitivityTestSuite) mockSingleLcWorkloadExecution() {
	s.mockedLcLauncher.On("Launch").Return(s.mockedLcTask, nil).Once()
	s.mockedLcTask.On("Stop").Return(nil).Once()
	s.mockedLcTask.On("Clean").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleLoadGeneratorTuning() {
	s.mockedLoadGenerator.On("Populate").Return(nil).Once()
	s.mockedLoadGenerator.On("Tune", s.configuration.SLO).Return(s.mockedPeakLoad, 4, nil).Once()
}

// Mocking single LoadGenerator flow when collection session for loadGenerator is successful
// as well.
func (s *SensitivityTestSuite) mockSingleLoadGeneratorLoad(loadPoint int) {
	s.mockedLoadGenerator.On("Populate").Return(nil).Once()
	s.mockedLoadGenerator.On(
		"Load", loadPoint, s.configuration.LoadDuration).Return(
		s.mockedLoadGeneratorTask, nil).Once()
	s.mockedLoadGeneratorTask.On(
		"Wait", 0*time.Nanosecond).Return(true).Once()
	s.mockedLoadGeneratorTask.On("ExitCode").Return(0, nil).Once()
	s.mockedLoadGeneratorTask.On("Clean").Return(nil).Once()
}

// Mocking single LoadGenerator flow when collection session for loadGenerator is NOT
// successful. That means ExitCode() methods will not be triggered.
func (s *SensitivityTestSuite) mockSingleLoadGeneratorLoadWithoutExitCode(loadPoint int) {
	s.mockedLoadGenerator.On("Populate").Return(nil).Once()
	s.mockedLoadGenerator.On(
		"Load", loadPoint, s.configuration.LoadDuration).Return(
		s.mockedLoadGeneratorTask, nil).Once()
	s.mockedLoadGeneratorTask.On(
		"Wait", 0*time.Nanosecond).Return(true).Once()
	s.mockedLoadGeneratorTask.On("Clean").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleAggressorWorkloadExecution() {
	s.mockedAggressor.On("Launch").Return(s.mockedAggressorTask, nil).Once()
	s.mockedAggressor.On("Name").Return("testName").Once()
	s.mockedAggressorTask.On("Stop").Return(nil).Once()
	s.mockedAggressorTask.On("Clean").Return(nil).Once()
}

func (s *SensitivityTestSuite) assertAllExpectations() {
	So(s.mockedLcLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLcTask.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLoadGenerator.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLoadGeneratorTask.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedAggressor.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedAggressorTask.AssertExpectations(s.T()), ShouldBeTrue)

	So(s.mockedLcSessionLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLcSessionHandle.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLoadGeneratorSessionLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLoadGeneratorSessionHandle.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedAggressorSessionLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedAggressorSessionHandle.AssertExpectations(s.T()), ShouldBeTrue)
}

func (s *SensitivityTestSuite) TestSensitivityTuningPhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment without any Snap session.

		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncherWithoutSession(s.mockedLcLauncher),
			NewLoadGeneratorWithoutSession(s.mockedLoadGenerator),
			[]LauncherSessionPair{
				NewLauncherWithoutSession(s.mockedAggressor),
			},
		)

		Convey("When production task can't be launched we expect failure", func() {
			s.mockedLcLauncher.On("Launch").Return(nil,
				errors.New("Production task can't be launched")).Once()

			err := s.sensitivityExperiment.Run()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEndWith, "Production task can't be launched")
		})

		Convey("When production task is launched successfully", func() {
			s.mockSingleLcWorkloadExecution()

			Convey("But load generator cannot populate and we expect failure", func() {
				s.mockedLoadGenerator.On("Populate").Return(errors.New("Populate")).Once()
				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Populate")
			})

			Convey("But load generator can't be tuned we expect failure", func() {
				s.mockedLoadGenerator.On("Populate").Return(nil).Once()
				s.mockedLoadGenerator.On("Tune", 1).Return(
					0, 0, errors.New("Load generator can't be tuned")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith, "Load generator can't be tuned")
			})

			Convey("When load generator can be tuned, but baseline fails "+
				"we are able to fetch TargetLoad result", func() {
				s.mockSingleLoadGeneratorTuning()

				// Make next measurement fail.
				s.mockedLcLauncher.On("Launch").Return(nil,
					errors.New("Production task can't be launched")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith, "Production task can't be launched")

				So(*s.sensitivityExperiment.tuningPhase.PeakLoad,
					ShouldEqual, s.mockedPeakLoad)
			})
			s.assertAllExpectations()
		})
	})
}

func (s *SensitivityTestSuite) TestSensitivityBaselinePhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment without any Snap session.
		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncherWithoutSession(s.mockedLcLauncher),
			NewLoadGeneratorWithoutSession(s.mockedLoadGenerator),
			[]LauncherSessionPair{
				NewLauncherWithoutSession(s.mockedAggressor),
			},
		)

		Convey("When tuning was successful", func() {
			// Mock successful tuning phase.
			s.mockSingleLcWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			Convey("But production task can't be launched during baseline we expect failure", func() {
				s.mockedLcLauncher.On("Launch").Return(nil,
					errors.New("Production task can't be launched during baseline")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith,
					"Production task can't be launched during baseline")
			})

			Convey("When production task is launched successfully during baseline", func() {
				s.mockSingleLcWorkloadExecution()

				Convey("But Populate fails for baseline and we expect failure", func() {
					s.mockedLoadGenerator.On("Populate").Return(errors.New("Populate")).Once()
					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Populate")
				})

				Convey("But load testing fails for baseline we expect failure", func() {
					s.mockedLoadGenerator.On("Populate").Return(nil).Once()
					s.mockedLoadGenerator.On("Load", 1, 1*time.Second).
						Return(nil, errors.New("Load testing failed")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Load testing failed")
				})

				Convey("When load is launched successfuly for first loadPoint", func() {
					s.mockSingleLoadGeneratorLoad(
						s.mockedPeakLoad / s.configuration.LoadPointsCount)

					Convey("After two baseline measurements for 2 loadpoints, we know "+
						"that Prod Launcher and Load Generator was launched 2 times", func() {
						// Mocking second iteration of baseline:
						s.mockSingleLcWorkloadExecution()
						s.mockSingleLoadGeneratorLoad(s.mockedPeakLoad)

						// Make next measurement fail.
						s.mockedAggressor.On("Name").Return("testName").Once()
						s.mockedLcLauncher.On("Launch").Return(nil,
							errors.New(
								"Production task can't be launched for aggressor phase")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEndWith,
							"Production task can't be launched for aggressor phase")
					})
				})
			})
		})

		s.assertAllExpectations()
	})
}

func (s *SensitivityTestSuite) TestSensitivityAggressorsPhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment without any Snap session.
		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncherWithoutSession(s.mockedLcLauncher),
			NewLoadGeneratorWithoutSession(s.mockedLoadGenerator),
			[]LauncherSessionPair{
				NewLauncherWithoutSession(s.mockedAggressor),
			},
		)

		Convey("When tuning and baselining was successful", func() {
			// Mock successful tuning phase.
			s.mockSingleLcWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			// Mock successful baseline phase (for 2 loadPoints)
			for i := 1; i <= s.configuration.LoadPointsCount; i++ {
				s.mockSingleLcWorkloadExecution()
				s.mockSingleLoadGeneratorLoad(
					i * (s.mockedPeakLoad / s.configuration.LoadPointsCount))
			}

			Convey("But production task can't be launched during aggressor phase, "+
				"we expect failure", func() {
				s.mockedAggressor.On("Name").Return("testName").Once()
				s.mockedLcLauncher.On("Launch").Return(nil,
					errors.New(
						"Production task can't be launched during aggressor phase")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith,
					"Production task can't be launched during aggressor phase")
			})

			Convey("When production task is launched successfully during aggressor phase", func() {
				s.mockSingleLcWorkloadExecution()

				Convey("But aggressor fails, we expect failure", func() {
					s.mockedAggressor.On("Launch").
						Return(nil, errors.New("Agressor failed")).Once()
					s.mockedAggressor.On("Name").Return("testName").Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Agressor failed")
				})

				Convey("When aggressor is launched successfuly", func() {
					s.mockSingleAggressorWorkloadExecution()

					Convey("But Populate fails for baseline and we expect failure", func() {
						s.mockedLoadGenerator.On("Populate").Return(errors.New("Populate")).Once()
						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEndWith, "Populate")
					})

					Convey("But when load testing fails, we expect failure", func() {
						s.mockedLoadGenerator.On("Populate").Return(nil).Once()
						s.mockedLoadGenerator.On("Load", 1, 1*time.Second).
							Return(nil, errors.New("Load testing failed")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEndWith, "Load testing failed")
					})

					Convey("When load is launched successfuly for last phase", func() {
						s.mockSingleLoadGeneratorLoad(
							s.mockedPeakLoad / s.configuration.LoadPointsCount)

						Convey("And next iteration is also sucessful, we have "+
							"made succesful experiment", func() {
							// Mocking second iteration of aggressor phase:
							s.mockSingleLcWorkloadExecution()
							s.mockSingleAggressorWorkloadExecution()
							s.mockSingleLoadGeneratorLoad(s.mockedPeakLoad)

							err := s.sensitivityExperiment.Run()
							So(err, ShouldBeNil)
						})
					})

				})
			})
		})

		s.assertAllExpectations()
	})
}

func (s *SensitivityTestSuite) mockSingleLcSessionExecution() {
	s.mockedLcSessionLauncher.On(
		"LaunchSession", s.mockedLcTask, mock.AnythingOfType("phase.Session")).Return(
		s.mockedLcSessionHandle, nil).Once()
	s.mockedLcSessionHandle.On("Stop").Return(nil).Once()
	s.mockedLcSessionHandle.On("Wait").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleLoadGeneratorSessionExecution() {
	s.mockedLoadGeneratorSessionLauncher.On(
		"LaunchSession", s.mockedLoadGeneratorTask, mock.AnythingOfType("phase.Session")).Return(
		s.mockedLoadGeneratorSessionHandle, nil).Once()
	s.mockedLoadGeneratorSessionHandle.On("Stop").Return(nil).Once()
	s.mockedLoadGeneratorSessionHandle.On("Wait").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleAggressorSessionExecution() {
	s.mockedAggressorSessionLauncher.On(
		"LaunchSession", s.mockedAggressorTask, mock.AnythingOfType("phase.Session")).Return(
		s.mockedAggressorSessionHandle, nil).Once()
	s.mockedAggressorSessionHandle.On("Stop").Return(nil).Once()
	s.mockedAggressorSessionHandle.On("Wait").Return(nil).Once()
}

func (s *SensitivityTestSuite) TestSensitivityWithSnapSessions() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment with Snap session per productionTask,
		// loadGenerator and aggressor.
		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewMonitoredLauncher(
				s.mockedLcLauncher,
				s.mockedLcSessionLauncher,
			),
			NewMonitoredLoadGenerator(
				s.mockedLoadGenerator,
				s.mockedLoadGeneratorSessionLauncher,
			),
			[]LauncherSessionPair{
				NewMonitoredLauncher(
					s.mockedAggressor, s.mockedAggressorSessionLauncher,
				),
			},
		)

		Convey("When tuning and baselining was successful, during aggressors phase", func() {
			// Mock successful tuning phase.
			s.mockSingleLcWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			// Mock successful baseline phase (for 2 loadPoints).
			for i := 1; i <= s.configuration.LoadPointsCount; i++ {
				s.mockSingleLcWorkloadExecution()
				s.mockSingleLcSessionExecution()
				s.mockSingleLoadGeneratorLoad(
					i * (s.mockedPeakLoad / s.configuration.LoadPointsCount))
				s.mockSingleLoadGeneratorSessionExecution()
			}

			// Mock first aggressors phase lcLauncher execution.
			s.mockSingleLcWorkloadExecution()
			Convey("When production task's Snap session can't be launched we expect failure",
				func() {
					s.mockedAggressor.On("Name").Return("testName").Once()
					s.mockedLcSessionLauncher.On(
						"LaunchSession",
						s.mockedLcTask,
						mock.AnythingOfType("phase.Session"),
					).Return(
						nil, errors.New("Production task's session can't be launched")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Production task's session can't be launched")
				})

			Convey("When production task's Snap session is launched successfully", func() {
				s.mockSingleLcSessionExecution()

				// Mock first aggressors phase aggressor execution.
				s.mockSingleAggressorWorkloadExecution()
				Convey("When aggressor's Snap session can't be launched we expect failure",
					func() {
						s.mockedAggressorSessionLauncher.On(
							"LaunchSession",
							s.mockedAggressorTask,
							mock.AnythingOfType("phase.Session")).Return(
							nil, errors.New("Aggressor's session can't be launched")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEndWith, "Aggressor's session can't be launched")
					})

				Convey("When aggressor's Snap session is launched successfully", func() {
					s.mockSingleAggressorSessionExecution()

					Convey("When loadGenerator Snap session can't be launched we expect failure",
						func() {
							s.mockSingleLoadGeneratorLoadWithoutExitCode(
								s.mockedPeakLoad / s.configuration.LoadPointsCount)
							// Mock error load generator session launcher.
							s.mockedLoadGeneratorSessionLauncher.On(
								"LaunchSession",
								s.mockedLoadGeneratorTask,
								mock.AnythingOfType("phase.Session")).Return(
								nil, errors.New("LoadGenerator session can't be launched")).Once()

							err := s.sensitivityExperiment.Run()
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldEndWith,
								"LoadGenerator session can't be launched")
						})

					Convey("When loadGenerator Snap session is launched successfully",
						func() {
							// Mock first aggressors phase loadGenerator execution.
							s.mockSingleLoadGeneratorLoad(
								s.mockedPeakLoad / s.configuration.LoadPointsCount)
							s.mockSingleLoadGeneratorSessionExecution()

							Convey("And next iteration is succesful, we expect no error", func() {
								// Last iteration.

								// Latency Sensitive task starts.
								s.mockSingleLcWorkloadExecution()
								// Latency Sensitive task's session starts.
								s.mockSingleLcSessionExecution()
								// Aggressor starts.
								s.mockSingleAggressorWorkloadExecution()
								// Aggressor's session starts.
								s.mockSingleAggressorSessionExecution()
								// Load generator starts.
								s.mockSingleLoadGeneratorLoad(s.mockedPeakLoad)
								// Load generator's session starts.
								s.mockSingleLoadGeneratorSessionExecution()

								err := s.sensitivityExperiment.Run()
								So(err, ShouldBeNil)
							})
						})
				})
			})
		})
		s.assertAllExpectations()
	})
}

func TestSensitivityExperiment(t *testing.T) {
	suite.Run(t, new(SensitivityTestSuite))
}
