package sensitivity

import (
	"errors"
	"github.com/Sirupsen/logrus"
	executorMocks "github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/workloads"
	workloadMocks "github.com/intelsdi-x/swan/pkg/workloads/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
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

	s.configuration = Configuration{
		SLO:             1,
		LoadDuration:    1 * time.Second,
		LoadPointsCount: 2,
		Repetitions:     1,
	}

	s.mockedTargetLoad = 2
}

func (s *SensitivityTestSuite) mockSingleProductionWorkloadExecution() {
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
	outputFile, err := ioutil.TempFile(os.TempDir(), "sensitivity")
	if err != nil {
		s.T().Fatal(err.Error())
	}
	s.mockedLoadGeneratorTask.On("StdoutFile").Return(outputFile, nil).Once()
	s.mockedLoadGeneratorTask.On("Clean").Return(nil).Once()
	s.mockedLoadGeneratorTask.On("ExitCode").Return(0, nil).Once()
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
}

func (s *SensitivityTestSuite) TestSensitivityTuningPhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment.
		s.sensitivityExperiment = NewExperiment("test", logrus.ErrorLevel, s.configuration,
			s.mockedLcLauncher, s.mockedLoadGenerator, []workloads.Launcher{s.mockedAggressor})

		Convey("When production task can't be launched we expect error", func() {
			s.mockedLcLauncher.On("Launch").Return(nil,
				errors.New("Production task can't be launched")).Once()

			err := s.sensitivityExperiment.Run()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldEqual, "Production task can't be launched")
		})

		Convey("When production task is launched successfully", func() {
			s.mockSingleProductionWorkloadExecution()

			Convey("But load generator can't be tuned we expect error", func() {

				s.mockedLoadGenerator.On("Tune", 1).Return(
					0, 0, errors.New("Load generator can't be tuned")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Load generator can't be tuned")
			})

			Convey("When load generator can be tuned, but baseline fails "+
				"we are able to fetch TargetLoad result", func() {
				s.mockSingleLoadGeneratorTuning()

				// Make next measurement fail.
				s.mockedLcLauncher.On("Launch").Return(nil,
					errors.New("Production task can't be launched")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual, "Production task can't be launched")

				So(*s.sensitivityExperiment.tuningPhase.TargetLoad, ShouldEqual, s.mockedTargetLoad)
			})
			s.assertAllExpectations()
		})
	})
}

func (s *SensitivityTestSuite) TestSensitivityBaselinePhase() {
	Convey("While using sensitivity profile experiment", s.T(), func() {
		// Create experiment.
		s.sensitivityExperiment = NewExperiment("test", logrus.ErrorLevel, s.configuration,
			s.mockedLcLauncher, s.mockedLoadGenerator, []workloads.Launcher{s.mockedAggressor})

		Convey("When tuning was successful", func() {
			// Mock successful tuning phase.
			s.mockSingleProductionWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			Convey("But production task can't be launched during baseline we expect error", func() {
				s.mockedLcLauncher.On("Launch").Return(nil,
					errors.New("Production task can't be launched during baseline")).Once()

				err := s.sensitivityExperiment.Run()
				So(err, ShouldNotBeNil)
				So(err.Error(), ShouldEqual,
					"Production task can't be launched during baseline")
			})

			Convey("When production task is launched successfully during baseline", func() {
				s.mockSingleProductionWorkloadExecution()

				Convey("But load testing fails for baseline we expect error", func() {
					s.mockedLoadGenerator.On("Load", 1, 1*time.Second).
						Return(nil, errors.New("Load testing failed")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Load testing failed")
				})

				Convey("When load is launched successfuly for first loadPoint", func() {
					s.mockSingleLoadGeneratorLoad(
						s.mockedTargetLoad / s.configuration.LoadPointsCount)

					Convey("After two baseline measurements for 2 loadpoints when we know "+
						"that Prod Launcher & Load Generator was launched 2 times", func() {
						// Mocking second iteration of baseline:
						s.mockSingleProductionWorkloadExecution()
						s.mockSingleLoadGeneratorLoad(s.mockedTargetLoad)

						// Make next measurement fail.
						s.mockedLcLauncher.On("Launch").Return(nil,
							errors.New(
								"Production task can't be launched for aggressor phase")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual,
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
		// Create experiment.
		s.sensitivityExperiment = NewExperiment("test", logrus.ErrorLevel, s.configuration,
			s.mockedLcLauncher, s.mockedLoadGenerator, []workloads.Launcher{s.mockedAggressor})

		Convey("When tuning & baselining was successful", func() {
			// Mock successful tuning phase.
			s.mockSingleProductionWorkloadExecution()
			s.mockSingleLoadGeneratorTuning()

			// Mock successful baseline phase (for 2 loadPoints)
			for i := 1; i <= s.configuration.LoadPointsCount; i++ {
				s.mockSingleProductionWorkloadExecution()
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
				So(err.Error(), ShouldEqual,
					"Production task can't be launched during aggressor phase")
			})

			Convey("When production task is launched successfully during aggressor phase", func() {
				s.mockSingleProductionWorkloadExecution()

				Convey("But aggressor fails, we expect error", func() {
					s.mockedAggressor.On("Launch").
						Return(nil, errors.New("Agressor failed")).Once()

					err := s.sensitivityExperiment.Run()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "Agressor failed")
				})

				Convey("When aggressor is launched successfuly", func() {
					s.mockSingleAggressorWorkloadExecution()

					Convey("But when load testing fails, we expect error", func() {
						s.mockedLoadGenerator.On("Load", 1, 1*time.Second).
							Return(nil, errors.New("Load testing failed")).Once()

						err := s.sensitivityExperiment.Run()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "Load testing failed")
					})

					Convey("When load is launched successfuly for last phase", func() {
						s.mockSingleLoadGeneratorLoad(
							s.mockedTargetLoad / s.configuration.LoadPointsCount)

						Convey("And next iteration is also sucessful, we have "+
							"made succesful experiment", func() {
							// Mocking second iteration of aggressor phase:
							s.mockSingleProductionWorkloadExecution()
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

func TestSensitivityExperiment(t *testing.T) {
	suite.Run(t, new(SensitivityTestSuite))
}
