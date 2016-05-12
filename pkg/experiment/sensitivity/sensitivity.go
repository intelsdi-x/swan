package sensitivity

import (
	log "github.com/Sirupsen/logrus"
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
	name                           string
	logLevel                       log.Level
	configuration                  Configuration
	productionTaskLauncher         workloads.Launcher
	loadGeneratorForProductionTask workloads.LoadGenerator
	aggressorTaskLaunchers         []workloads.Launcher

	// Generic Experiment.
	exp *experiment.Experiment
	// Phases.
	tuningPhase     *tuningPhase
	baselinePhase   []experiment.Phase
	aggressorPhases [][]experiment.Phase
}

// NewExperiment construct new Experiment object.
// Input parameters:
// configuration - Experiment configuration
// productionTaskLauncher - Latency Critical job launcher
// loadGeneratorForProductionTask - stresser for production task
// aggressorTasksLauncher - Best Effort jobs launcher
func NewExperiment(
	name string,
	logLevel log.Level,
	configuration Configuration,
	productionTaskLauncher workloads.Launcher,
	loadGeneratorForProductionTask workloads.LoadGenerator,
	aggressorTaskLaunchers []workloads.Launcher) *Experiment {

	// TODO(mpatelcz): Validate configuration.
	return &Experiment{
		name:                           name,
		logLevel:                       logLevel,
		configuration:                  configuration,
		productionTaskLauncher:         productionTaskLauncher,
		loadGeneratorForProductionTask: loadGeneratorForProductionTask,
		aggressorTaskLaunchers:         aggressorTaskLaunchers,
	}
}

func (e *Experiment) getTuningPhase() *tuningPhase {
	targetLoad := float64(-1)
	return &tuningPhase{
		pr:          e.productionTaskLauncher,
		lgForPr:     e.loadGeneratorForProductionTask,
		SLO:         e.configuration.SLO,
		repetitions: e.configuration.Repetitions,
		TargetLoad:  &targetLoad,
	}
}

func (e *Experiment) getBaselinePhases() []experiment.Phase {
	baseline := []experiment.Phase{}
	// It includes all baseline measurements for each LoadPoint.
	for i := 1; i <= e.configuration.LoadPointsCount; i++ {
		baseline = append(baseline, &measurementPhase{
			namePrefix:      "baseline",
			pr:              e.productionTaskLauncher,
			lgForPr:         e.loadGeneratorForProductionTask,
			bes:             []workloads.Launcher{},
			loadDuration:    e.configuration.LoadDuration,
			loadPointsCount: e.configuration.LoadPointsCount,
			repetitions:     e.configuration.Repetitions,
			TargetLoad:      e.tuningPhase.TargetLoad,
			// Measurements in baseline have different load point input.
			currentLoadPointIndex: i,
		})
	}

	return baseline
}

func (e *Experiment) getAggressorsPhases() [][]experiment.Phase {
	aggressorPhases := [][]experiment.Phase{}
	for beIndex, beLauncher := range e.aggressorTaskLaunchers {
		aggressorPhase := []experiment.Phase{}
		// Include measurements for each LoadPoint.
		for i := 1; i <= e.configuration.LoadPointsCount; i++ {
			aggressorPhase = append(aggressorPhase, &measurementPhase{
				namePrefix:            "aggressor_nr_" + strconv.Itoa(beIndex),
				pr:                    e.productionTaskLauncher,
				lgForPr:               e.loadGeneratorForProductionTask,
				bes:                   []workloads.Launcher{beLauncher},
				loadDuration:          e.configuration.LoadDuration,
				loadPointsCount:       e.configuration.LoadPointsCount,
				repetitions:           e.configuration.Repetitions,
				currentLoadPointIndex: i,
				TargetLoad:            e.tuningPhase.TargetLoad,
			})
		}
		aggressorPhases = append(aggressorPhases, aggressorPhase)
	}

	return aggressorPhases
}

func (e *Experiment) configureGenericExperiment() error {
	// Configure phases & measurements.
	// Each sensitivity phase (part of experiment) can include couple of measurements.
	var allMeasurements []experiment.Phase

	// Include Tuning Phase.
	e.tuningPhase = e.getTuningPhase()
	allMeasurements = append(allMeasurements, e.tuningPhase)

	// Include Baseline Phase.
	e.baselinePhase = e.getBaselinePhases()
	allMeasurements = append(allMeasurements, e.baselinePhase...)

	// Include Measurement Phases for each aggressor.
	e.aggressorPhases = e.getAggressorsPhases()
	for _, aggressorPhase := range e.aggressorPhases {
		allMeasurements = append(allMeasurements, aggressorPhase...)
	}

	var err error
	e.exp, err = experiment.NewExperiment(e.name, allMeasurements, os.TempDir(), e.logLevel)
	if err != nil {
		return err
	}

	return nil
}

// Run runs experiment.
// In the end it prints results to the standard output.
func (e *Experiment) Run() error {
	err := e.configureGenericExperiment()
	if err != nil {
		return err
	}

	err = e.exp.Run()
	defer e.exp.Finalize()
	if err != nil {
		return err
	}

	return nil
}
