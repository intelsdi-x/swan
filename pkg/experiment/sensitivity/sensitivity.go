package sensitivity

import (
	"os"
	"strconv"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
)

var (
	sloFlag             = conf.NewIntFlag("slo", "Given SLO for the experiment. [us]", 500)
	loadPointsCountFlag = conf.NewIntFlag("load_points", "Number of load points to test", 10)
	loadDurationFlag    = conf.NewDurationFlag("load_duration", "Load duration [s].", 10*time.Second)
	repetitionsFlag     = conf.NewIntFlag("reps", "Number of repetitions for each measurement", 3)
	// peakLoadFlag represents special case when peak load is provided instead of calculated from Tuning phase.
	// It omits tuning phase.
	peakLoadFlag = conf.NewIntFlag("peak_load", "Peakload max number of QPS without violating SLO (by default inducted from tunning phase).", 0) // "0" means include tunning phase.
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
	// PeakLoad. If set >0 skip tuning phase.
	PeakLoad int
}

// DefaultConfiguration returns default configuration for experiment from Conf flags.
func DefaultConfiguration() Configuration {
	return Configuration{
		SLO:             sloFlag.Value(),
		LoadDuration:    loadDurationFlag.Value(),
		LoadPointsCount: loadPointsCountFlag.Value(),
		Repetitions:     repetitionsFlag.Value(),
		PeakLoad:        peakLoadFlag.Value(),
	}
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
	return &tuningPhase{
		pr:          e.productionTaskLauncher.Launcher,
		lgForPr:     e.loadGeneratorForProductionTask.LoadGenerator,
		SLO:         e.configuration.SLO,
		repetitions: e.configuration.Repetitions,
		PeakLoad:    &e.configuration.PeakLoad,
	}
}

func (e *Experiment) prepareBaselinePhases() []phase.Phase {
	baseline := []phase.Phase{}
	// It includes all baseline measurements for each LoadPoint.
	for i := 1; i <= e.configuration.LoadPointsCount; i++ {
		baseline = append(baseline, &measurementPhase{
			namePrefix:      "baseline",
			pr:              e.productionTaskLauncher,
			lgForPr:         e.loadGeneratorForProductionTask,
			bes:             []LauncherSessionPair{},
			loadDuration:    e.configuration.LoadDuration,
			loadPointsCount: e.configuration.LoadPointsCount,
			repetitions:     e.configuration.Repetitions,
			PeakLoad:        &e.configuration.PeakLoad,
			// Measurements in baseline have different load point input.
			currentLoadPointIndex: i,
		})
	}

	return baseline
}

func (e *Experiment) prepareAggressorsPhases() [][]phase.Phase {
	aggressorPhases := [][]phase.Phase{}
	for beIndex, beLauncher := range e.aggressorTaskLaunchers {
		aggressorPhase := []phase.Phase{}
		// Include measurements for each LoadPoint.
		for i := 1; i <= e.configuration.LoadPointsCount; i++ {
			aggressorPhase = append(aggressorPhase, &measurementPhase{
				namePrefix:            "aggressor_nr_" + strconv.Itoa(beIndex),
				pr:                    e.productionTaskLauncher,
				lgForPr:               e.loadGeneratorForProductionTask,
				bes:                   []LauncherSessionPair{beLauncher},
				loadDuration:          e.configuration.LoadDuration,
				loadPointsCount:       e.configuration.LoadPointsCount,
				repetitions:           e.configuration.Repetitions,
				currentLoadPointIndex: i,
				PeakLoad:              &e.configuration.PeakLoad,
			})
		}
		aggressorPhases = append(aggressorPhases, aggressorPhase)
	}

	return aggressorPhases
}

func (e *Experiment) configureGenericExperiment() error {
	// Configure phases & measurements.
	// Each sensitivity phase (part of experiment) can include couple of measurements.
	var allMeasurements []phase.Phase

	// Include Tuning Phase if PeakLoad wasn't given.
	if e.configuration.PeakLoad == 0 {
		e.tuningPhase = e.prepareTuningPhase()
		allMeasurements = append(allMeasurements, e.tuningPhase)
	} else {
		log.Debugf("skipping Tunning phase (peakload=%d)", e.configuration.PeakLoad)
	}

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

	return nil
}
