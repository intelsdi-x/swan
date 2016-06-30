package sensitivity

import (
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"os"
	"strconv"
	"time"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/metadata"
	"encoding/json"
	"strings"
)

// Configuration - set of parameters to control the experiment.
type Configuration struct {
	// Given SLO for the experiment.
	// TODO(bp): Push that to DB via Snap in tag or using SwanCollector.
	SLO int
	// Each measurement duration in [s].
	// TODO(bp): Push that to DB via Snap in tag or using SwanCollector.
	LoadDuration time.Duration
	// Number of load points to test.
	// TODO(bp): Push that to DB via Snap in tag or using SwanCollector.
	LoadPointsCount int
	// Repetitions.
	// TODO(bp): Push that to DB via Snap in tag or using SwanCollector.
	Repetitions int
}

// Experiment is handler structure for Experiment Driver. All fields shall be
// not visible to the experimenter.
type Experiment struct {
	name                           string
	logLevel                       log.Level
	configuration                  Configuration
	productionTaskLauncher         LauncherSessionPair
	loadGeneratorForProductionTask LoadGeneratorSessionPair
	aggressorTaskLaunchers         []LauncherSessionPair

	// Generic Experiment.
	exp *experiment.Experiment
	// Phases.
	tuningPhase     *tuningPhase
	baselinePhase   []phase.Phase
	aggressorPhases [][]phase.Phase

	experimentMetadata metadata.Experiment
	phasesMetadata []metadata.Phase
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
	productionTaskLauncher LauncherSessionPair,
	loadGeneratorForProductionTask LoadGeneratorSessionPair,
	aggressorTaskLaunchers []LauncherSessionPair) *Experiment {

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

func (e *Experiment) prepareTuningPhase() *tuningPhase {
	peakLoad := int(-1)
	e.phasesMetadata = append(e.phasesMetadata, metadata.Phase{
		BasePhase: metadata.BasePhase{
			ID: "tuning",
			LCParameters: e.productionTaskLauncher.Launcher.Name(),
			LCIsolation: e.productionTaskLauncher.Launcher.Isolators(),
		},

	})
	return &tuningPhase{
		pr:          e.productionTaskLauncher.Launcher,
		lgForPr:     e.loadGeneratorForProductionTask.LoadGenerator,
		SLO:         e.configuration.SLO,
		repetitions: e.configuration.Repetitions,
		PeakLoad:    &peakLoad,
	}
}

func (e *Experiment) prepareBaselinePhases() []phase.Phase {
	var baselineMeasurementsMetadata []metadata.Measurement
	baseline := []phase.Phase{}
	// It includes all baseline measurements for each LoadPoint.
	for i := 1; i <= e.configuration.LoadPointsCount; i++ {
		baselinePhase := &measurementPhase{
			namePrefix:      "baseline",
			pr:              e.productionTaskLauncher,
			lgForPr:         e.loadGeneratorForProductionTask,
			bes:             []LauncherSessionPair{},
			loadDuration:    e.configuration.LoadDuration,
			loadPointsCount: e.configuration.LoadPointsCount,
			repetitions:     e.configuration.Repetitions,
			PeakLoad:        e.tuningPhase.PeakLoad,
			// Measurements in baseline have different load point input.
			currentLoadPointIndex: i,
		}
		baselineMeasurementsMetadata = append(baselineMeasurementsMetadata, baselinePhase.GetMetadata())
		baseline = append(baseline, baselinePhase)
	}
	e.phasesMetadata = append(e.phasesMetadata, metadata.Phase{
		BasePhase: metadata.BasePhase{
			ID: "baseline",
			LCIsolation: e.productionTaskLauncher.Launcher.Isolators(),
			LCParameters: e.productionTaskLauncher.Launcher.Parameters(),
		},
		Measurements: baselineMeasurementsMetadata,
	})

	return baseline
}

func (e *Experiment) prepareAggressorsPhases() [][]phase.Phase {
	measurementPhases := []metadata.Measurement{}
	aggressors := []metadata.Aggressor{}
	aggressorPhases := [][]phase.Phase{}
	for beIndex, beLauncher := range e.aggressorTaskLaunchers {
		aggressorPhase := []phase.Phase{}
		// Include measurements for each LoadPoint.
		for i := 1; i <= e.configuration.LoadPointsCount; i++ {
			aggressorPhaseData := &measurementPhase{
				namePrefix:            "aggressor_nr_" + strconv.Itoa(beIndex),
				pr:                    e.productionTaskLauncher,
				lgForPr:               e.loadGeneratorForProductionTask,
				bes:                   []LauncherSessionPair{beLauncher},
				loadDuration:          e.configuration.LoadDuration,
				loadPointsCount:       e.configuration.LoadPointsCount,
				repetitions:           e.configuration.Repetitions,
				currentLoadPointIndex: i,
				PeakLoad:              e.tuningPhase.PeakLoad,
			}
			measurementPhases = append(measurementPhases, aggressorPhaseData.GetMetadata())
			aggressorPhase = append(aggressorPhase, aggressorPhaseData)
		}
		aggressors = append(aggressors, metadata.Aggressor{
			Name: beLauncher.Launcher.Name(),
			Parameters: beLauncher.Launcher.Parameters(),
		})

		aggressorPhases = append(aggressorPhases, aggressorPhase)
	}
	e.phasesMetadata = append(e.phasesMetadata, metadata.Phase{
		BasePhase: metadata.BasePhase{
			ID: "aggressor",
			LCParameters: e.productionTaskLauncher.Launcher.Parameters(),
			LCIsolation: e.productionTaskLauncher.Launcher.Isolators(),
		},
		Measurements: measurementPhases,
		Aggressors: aggressors,
	})
	return aggressorPhases
}

func (e *Experiment) generateMetadata() {
	e.experimentMetadata = metadata.Experiment{
		BaseExperiment: metadata.BaseExperiment{
			ID: e.exp.GetUUID(),
			LoadDuration: e.configuration.LoadDuration,
			LCName: e.productionTaskLauncher.Launcher.Name(),
			LGName: e.loadGeneratorForProductionTask.LoadGenerator.Name(),
			RepetitionsNumber: e.configuration.Repetitions,
			LoadPointsNumber: e.configuration.LoadPointsCount,
			SLO: e.configuration.SLO,
		},
		Phases: e.phasesMetadata,
	}
}


func (e *Experiment) configureGenericExperiment() error {
	// Configure phases & measurements.
	// Each sensitivity phase (part of experiment) can include couple of measurements.
	var allMeasurements []phase.Phase

	// Include Tuning Phase.
	e.tuningPhase = e.prepareTuningPhase()
	allMeasurements = append(allMeasurements, e.tuningPhase)

	// Include Baseline Phase.
	e.baselinePhase = e.prepareBaselinePhases()
	allMeasurements = append(allMeasurements, e.baselinePhase...)

	// Include Measurement Phases for each aggressor.
	e.aggressorPhases = e.prepareAggressorsPhases()
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
	e.generateMetadata()
	data, err := json.Marshal(e.experimentMetadata)
	log.Printf("%s\nDATA: %s\nERROR: %s\n%s", strings.Repeat("#", 10), data, err, strings.Repeat("#", 10))

	return nil
}
