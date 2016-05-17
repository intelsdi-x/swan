package sensitivity

import (
	"errors"
	"github.com/Sirupsen/logrus"
	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	snapMocks "github.com/intelsdi-x/swan/pkg/snap/mocks"
	workloadMocks "github.com/intelsdi-x/swan/pkg/workloads/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"testing"
	"time"
)

type SensitivityTestSuite struct {
	suite.Suite

	// LC launcher & task.
	mockedLcLauncher *workloadMocks.Launcher
	mockedLcTask     *executorMocks.TaskHandle
	// LoadGenerator launcher & task.
	mockedLoadGenerator     *workloadMocks.LoadGenerator
	mockedLoadGeneratorTask *executorMocks.TaskHandle
	// Aggressor launcher & task.
	mockedAggressor     *workloadMocks.Launcher
	mockedAggressorTask *executorMocks.TaskHandle

	// LC collection Launcher & Handle.
	mockedLcCollectionLauncher *snapMocks.SessionLauncher
	mockedLcCollectionHandle   *snapMocks.SessionHandle
	// LoadGenerator Launcher & Handle.
	mockedLoadGeneratorCollectionLauncher *snapMocks.SessionLauncher
	mockedLoadGeneratorCollectionHandle   *snapMocks.SessionHandle
	// Aggressor Launcher & Handle.
	mockedAggressorCollectionLauncher *snapMocks.SessionLauncher
	mockedAggressorCollectionHandle   *snapMocks.SessionHandle

	configuration         Configuration
	sensitivityExperiment *Experiment

	mockedTargetLoad int
}

func (s *SensitivityTestSuite) SetupTest() {
	// LC launcher & task.
	s.mockedLcLauncher = new(workloadMocks.Launcher)
	s.mockedLcTask = new(executorMocks.TaskHandle)
	// LoadGenerator launcher & task.
	s.mockedLoadGenerator = new(workloadMocks.LoadGenerator)
	s.mockedLoadGeneratorTask = new(executorMocks.TaskHandle)
	// Aggressor launcher & task.
	s.mockedAggressor = new(workloadMocks.Launcher)
	s.mockedAggressorTask = new(executorMocks.TaskHandle)

	// LC collection Launcher & Handle.
	s.mockedLcCollectionLauncher = new(snapMocks.SessionLauncher)
	s.mockedLcCollectionHandle = new(snapMocks.SessionHandle)
	// LoadGenerator Launcher & Handle.
	s.mockedLoadGeneratorCollectionLauncher = new(snapMocks.SessionLauncher)
	s.mockedLoadGeneratorCollectionHandle = new(snapMocks.SessionHandle)
	// Aggressor Launcher & Handle.
	s.mockedAggressorCollectionLauncher = new(snapMocks.SessionLauncher)
	s.mockedAggressorCollectionHandle = new(snapMocks.SessionHandle)

	s.configuration = Configuration{
		SLO:             1,
		LoadDuration:    1 * time.Second,
		LoadPointsCount: 2,
		Repetitions:     1,
	}

	s.mockedTargetLoad = 2
}

func (s *SensitivityTestSuite) mockSingleLcWorkloadExecution() {
	s.mockedLcLauncher.On("Launch").Return(s.mockedLcTask, nil).Once()
	s.mockedLcTask.On("Stop").Return(nil).Once()
	s.mockedLcTask.On("Clean").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleLoadGeneratorTuning() {
	s.mockedLoadGenerator.On("Tune", s.configuration.SLO).Return(s.mockedTargetLoad, 4, nil).Once()
}

func (s *SensitivityTestSuite) mockSingleLoadGeneratorLoad(loadPoint int) {
	s.mockedLoadGenerator.On(
		"Load", loadPoint, s.configuration.LoadDuration).Return(
		s.mockedLoadGeneratorTask, nil).Once()
	s.mockedLoadGeneratorTask.On(
		"Wait", 0*time.Nanosecond).Return(true).Once()
	s.mockedLoadGeneratorTask.On("Clean").Return(nil).Once()
	s.mockedLoadGeneratorTask.On("ExitCode").Return(0, nil).Once()
}

func (s *SensitivityTestSuite) mockSingleLoadGeneratorLoadWhenSessionErr(loadPoint int) {
	s.mockedLoadGenerator.On(
		"Load", loadPoint, s.configuration.LoadDuration).Return(
		s.mockedLoadGeneratorTask, nil).Once()
	s.mockedLoadGeneratorTask.On("Clean").Return(nil).Once()
	// Exceptional case for loadGenerator mock. Wait() and ExitCode()
	// won't be triggered since we expect error before that happens.
}

func (s *SensitivityTestSuite) mockSingleAggressorWorkloadExecution() {
	s.mockedAggressor.On("Launch").Return(s.mockedAggressorTask, nil).Once()
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

	So(s.mockedLcCollectionLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLcCollectionHandle.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLoadGeneratorCollectionLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedLoadGeneratorCollectionHandle.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedAggressorCollectionLauncher.AssertExpectations(s.T()), ShouldBeTrue)
	So(s.mockedAggressorCollectionHandle.AssertExpectations(s.T()), ShouldBeTrue)
}

func (s *SensitivityTestSuite) TestSensitivityTuningPhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment without any collection session.
		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncher(s.mockedLcLauncher),
			NewLoadGenerator(s.mockedLoadGenerator),
			[]LauncherWithCollection{
				NewLauncher(s.mockedAggressor),
			},
		)

		Convey("When production task can't be launched we expect error", func() {
			s.mockedLcLauncher.On("Launch").Return(nil,
				errors.New("Production task can't be launched")).Once()

			err := s.sensitivityExperiment.Run()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEndWith, "Production task can't be launched")
		})

		Convey("When production task is launched successfully", func() {
			s.mockSingleLcWorkloadExecution()

			Convey("But load generator can't be tuned we expect error", func() {

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

				So(*s.sensitivityExperiment.tuningPhase.TargetLoad, ShouldEqual, s.mockedTargetLoad)
			})
			s.assertAllExpectations()
		})
	})
}

func (s *SensitivityTestSuite) TestSensitivityBaselinePhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment without any collection session.
		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncher(s.mockedLcLauncher),
			NewLoadGenerator(s.mockedLoadGenerator),
			[]LauncherWithCollection{
				NewLauncher(s.mockedAggressor),
			},
		)

		Convey("When tuning was successful", func() {
			// Mock successful tuning phase.
			s.mockSingleLcWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			Convey("But production task can't be launched during baseline we expect error", func() {
				s.mockedLcLauncher.On("Launch").Return(nil,
					errors.New("Production task can't be launched during baseline")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEndWith,
					"Production task can't be launched during baseline")
			})

			Convey("When production task is launched successfully during baseline", func() {
				s.mockSingleLcWorkloadExecution()

				Convey("But load testing fails for baseline we expect error", func() {
					s.mockedLoadGenerator.On("Load", 1, 1*time.Second).
						Return(nil, errors.New("Load testing failed")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Load testing failed")
				})

				Convey("When load is launched successfuly for first loadPoint", func() {
					s.mockSingleLoadGeneratorLoad(
						s.mockedTargetLoad / s.configuration.LoadPointsCount)

					Convey("After two baseline measurements for 2 loadpoints, we know "+
						"that Prod Launcher & Load Generator was launched 2 times", func() {
						// Mocking second iteration of baseline:
						s.mockSingleLcWorkloadExecution()
						s.mockSingleLoadGeneratorLoad(s.mockedTargetLoad)

						// Make next measurement fail.
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
		// Create experiment without any collection session.
		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncher(s.mockedLcLauncher),
			NewLoadGenerator(s.mockedLoadGenerator),
			[]LauncherWithCollection{
				NewLauncher(s.mockedAggressor),
			},
		)

		Convey("When tuning & baselining was successful", func() {
			// Mock successful tuning phase.
			s.mockSingleLcWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			// Mock successful baseline phase (for 2 loadPoints)
			for i := 1; i <= s.configuration.LoadPointsCount; i++ {
				s.mockSingleLcWorkloadExecution()
				s.mockSingleLoadGeneratorLoad(
					i * (s.mockedTargetLoad / s.configuration.LoadPointsCount))
			}

			Convey("But production task can't be launched during aggressor phase, "+
				"we expect error", func() {
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

				Convey("But aggressor fails, we expect error", func() {
					s.mockedAggressor.On("Launch").
						Return(nil, errors.New("Agressor failed")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Agressor failed")
				})

				Convey("When aggressor is launched successfuly", func() {
					s.mockSingleAggressorWorkloadExecution()

					Convey("But when load testing fails, we expect error", func() {
						s.mockedLoadGenerator.On("Load", 1, 1*time.Second).
							Return(nil, errors.New("Load testing failed")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEndWith, "Load testing failed")
					})

					Convey("When load is launched successfuly for last phase", func() {
						s.mockSingleLoadGeneratorLoad(
							s.mockedTargetLoad / s.configuration.LoadPointsCount)

						Convey("And next iteration is also sucessful, we have "+
							"made succesful experiment", func() {
							// Mocking second iteration of aggressor phase:
							s.mockSingleLcWorkloadExecution()
							s.mockSingleAggressorWorkloadExecution()
							s.mockSingleLoadGeneratorLoad(s.mockedTargetLoad)

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

func (s *SensitivityTestSuite) mockSingleLcCollectionExecution() {
	s.mockedLcCollectionLauncher.On(
		"Launch", s.mockedLcTask, mock.AnythingOfType("phase.Session")).Return(
		s.mockedLcCollectionHandle, nil).Once()
	s.mockedLcCollectionHandle.On("Stop").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleLoadGeneratorCollectionExecution() {
	s.mockedLoadGeneratorCollectionLauncher.On(
		"Launch", s.mockedLoadGeneratorTask, mock.AnythingOfType("phase.Session")).Return(
		s.mockedLoadGeneratorCollectionHandle, nil).Once()
	s.mockedLoadGeneratorCollectionHandle.On("Stop").Return(nil).Once()
}

func (s *SensitivityTestSuite) mockSingleAggressorCollectionExecution() {
	s.mockedAggressorCollectionLauncher.On(
		"Launch", s.mockedAggressorTask, mock.AnythingOfType("phase.Session")).Return(
		s.mockedAggressorCollectionHandle, nil).Once()
	s.mockedAggressorCollectionHandle.On("Stop").Return(nil).Once()
}

func (s *SensitivityTestSuite) TestSensitivityWithCollectionSessions() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment with collection session per productionTask,
		// loadGenerator and aggressor.

		s.sensitivityExperiment = NewExperiment(
			"test",
			logrus.ErrorLevel,
			s.configuration,
			NewLauncherWithCollection(
				s.mockedLcLauncher,
				s.mockedLcCollectionLauncher,
			),
			NewLoadGeneratorWithCollection(
				s.mockedLoadGenerator,
				s.mockedLoadGeneratorCollectionLauncher,
			),
			[]LauncherWithCollection{
				NewLauncherWithCollection(
					s.mockedAggressor, s.mockedAggressorCollectionLauncher,
				),
			},
		)

		Convey("When tuning & baselining was successful, during aggressors phase", func() {
			// Mock successful tuning phase.
			s.mockSingleLcWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			// Mock successful baseline phase (for 2 loadPoints)
			for i := 1; i <= s.configuration.LoadPointsCount; i++ {
				s.mockSingleLcWorkloadExecution()
				s.mockSingleLcCollectionExecution()
				s.mockSingleLoadGeneratorLoad(
					i * (s.mockedTargetLoad / s.configuration.LoadPointsCount))
				s.mockSingleLoadGeneratorCollectionExecution()
			}

			// Mock first aggressors phase lcLauncher execution.
			s.mockSingleLcWorkloadExecution()
			Convey("When production task's collection session can't be launched we expect error",
				func() {
					s.mockedLcCollectionLauncher.On(
						"Launch", s.mockedLcTask, mock.AnythingOfType("phase.Session")).Return(
						nil, errors.New("Production task's collection can't be launched")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEndWith, "Production task's collection can't be launched")
				})

			Convey("When production task's collection session is launched successfully", func() {
				s.mockSingleLcCollectionExecution()

				// Mock first aggressors phase aggressor execution.
				s.mockSingleAggressorWorkloadExecution()
				Convey("When aggressor's collection session can't be launched we expect error",
					func() {
						s.mockedAggressorCollectionLauncher.On(
							"Launch",
							s.mockedAggressorTask,
							mock.AnythingOfType("phase.Session")).Return(
							nil, errors.New("Aggressor's collection can't be launched")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEndWith, "Aggressor's collection can't be launched")
					})

				Convey("When aggressor's collection session is launched successfully", func() {
					s.mockSingleAggressorCollectionExecution()

					Convey("When loadGenerator collection session can't be launched we expect error",
						func() {
							// Mock first aggressors phase loadGenerator execution.
							s.mockSingleLoadGeneratorLoadWhenSessionErr(
								s.mockedTargetLoad / s.configuration.LoadPointsCount)
							s.mockedLoadGeneratorCollectionLauncher.On(
								"Launch",
								s.mockedLoadGeneratorTask,
								mock.AnythingOfType("phase.Session")).Return(
								nil, errors.New("LoadGenerator collection can't be launched")).Once()

							err := s.sensitivityExperiment.Run()
							So(err, ShouldNotBeNil)
							So(err.Error(), ShouldEndWith,
								"LoadGenerator collection can't be launched")
						})

					Convey("When loadGenerator collection session is launched successfully",
						func() {
							// Mock first aggressors phase loadGenerator execution.
							s.mockSingleLoadGeneratorLoad(
								s.mockedTargetLoad / s.configuration.LoadPointsCount)
							s.mockSingleLoadGeneratorCollectionExecution()

							Convey("And next iteration is succesful, we expect no error", func() {
								// Last iteration.

								// Latency Sensitive task starts.
								s.mockSingleLcWorkloadExecution()
								// Latency Sensitive task's collection starts.
								s.mockSingleLcCollectionExecution()
								// Aggressor starts.
								s.mockSingleAggressorWorkloadExecution()
								// Aggressor's collection starts.
								s.mockSingleAggressorCollectionExecution()
								// Load generator starts.
								s.mockSingleLoadGeneratorLoad(s.mockedTargetLoad)
								// Load generator's collection starts.
								s.mockSingleLoadGeneratorCollectionExecution()

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
