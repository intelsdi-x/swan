package sensitivity

import (
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"os"
	"strconv"
	"time"
)

// Configuration - set of parameters to control the experiment.
type Configuration struct {
	// Given SLO for the experiment.
	SLO int
	// Each measurement duration in [s].
	LoadDuration time.Duration
	// Number of load points to test.
	LoadPointsCount int
	// Repetitions.
	Repetitions int
}

// Experiment is handler structure for Experiment Driver. All fields shall be
// not visible to the experimenter.
type Experiment struct {
	exp             *experiment.Experiment
	tuning          *tuningPhase
	baselinePhase   []*measurement
	aggressorPhases [][]*measurement
	log             *logrus.Logger
}

// InitExperiment construct new Experiment object.
// Input parameters:
// configuration - Experiment configuration
// productionTaskLauncher - Latency Critical job launcher
// loadGeneratorForProductionTask - stresser for production task
// aggressorTasksLauncher - Best Effort jobs launcher
func InitExperiment(
	name string,
	logLvl logrus.Level,
	configuration Configuration,
	productionTaskLauncher workloads.Launcher,
	loadGeneratorForProductionTask workloads.LoadGenerator,
	aggressorTaskLaunchers []workloads.Launcher) (*Experiment, error) {

	// TODO(mpatelcz): Validate configuration.
	// Configure phases & measurements.
	// Each phases includes couple of measurements.
	var allMeasurements []experiment.Measurement
	// Include Tuning Phase.
	targetLoad := float64(-1)
	tuning := &tuningPhase{
		pr:          productionTaskLauncher,
		lgForPr:     loadGeneratorForProductionTask,
		SLO:         configuration.SLO,
		repetitions: configuration.Repetitions,
		TargetLoad:  &targetLoad,
	}
	allMeasurements = append(allMeasurements, tuning)

	// Include Baseline Phase.
	baselinePhase := []*measurement{}
	// It includes measurements for each LoadPoint.
	for i := 1; i <= configuration.LoadPointsCount; i++ {
		baselinePhase = append(baselinePhase, &measurement{
			namePrefix:            "Baseline",
			pr:                    productionTaskLauncher,
			lgForPr:               loadGeneratorForProductionTask,
			bes:                   []workloads.Launcher{},
			loadDuration:          configuration.LoadDuration,
			loadPointsCount:       configuration.LoadPointsCount,
			repetitions:           configuration.Repetitions,
			currentLoadPointIndex: i,
			TargetLoad:            tuning.TargetLoad,
		})
		allMeasurements = append(allMeasurements, baselinePhase[i-1])
	}

	// Include Measurement Phases for each aggressor.
	aggressorPhases := [][]*measurement{}
	for beIndex, beLauncher := range aggressorTaskLaunchers {
		aggressorPhase := []*measurement{}
		// Include measurements for each LoadPoint.
		for i := 1; i <= configuration.LoadPointsCount; i++ {
			aggressorPhase = append(aggressorPhase, &measurement{
				namePrefix:            "Aggressor nr " + strconv.Itoa(beIndex),
				pr:                    productionTaskLauncher,
				lgForPr:               loadGeneratorForProductionTask,
				bes:                   []workloads.Launcher{beLauncher},
				loadDuration:          configuration.LoadDuration,
				loadPointsCount:       configuration.LoadPointsCount,
				repetitions:           configuration.Repetitions,
				currentLoadPointIndex: i,
				TargetLoad:            tuning.TargetLoad,
			})

			allMeasurements = append(allMeasurements, aggressorPhase[i-1])
		}
		aggressorPhases = append(aggressorPhases, aggressorPhase)
	}

	exp, err := experiment.NewExperiment(name, allMeasurements, os.TempDir(), logLvl)
	if err != nil {
		return nil, err
	}

	return &Experiment{
		exp:             exp,
		baselinePhase:   baselinePhase,
		aggressorPhases: aggressorPhases,
		log:             exp.Log,
	}, nil
}

// Run runs experiment.
// In the end it prints results to the standard output.
func (e *Experiment) Run() error {
	err := e.exp.Run()
	defer e.exp.Finalize()
	if err != nil {
		return err
	}

	// TODO(bp): Save to file. In future this will be passed to Snap.
	for _, baselineMeasurement := range e.baselinePhase {
		e.log.Debug(baselineMeasurement.Name(), " = ",
			baselineMeasurement.MeasurementSliResult)
	}
	for _, aggressorPhase := range e.aggressorPhases {
		for _, aggrMeasurment := range aggressorPhase {
			e.log.Debug(aggrMeasurment.Name(), " = ",
				aggrMeasurment.MeasurementSliResult)
		}
	}

	return nil
}
