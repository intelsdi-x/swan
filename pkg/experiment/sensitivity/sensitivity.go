package sensitivity

import "github.com/intelsdi-x/swan/pkg/workloads"

// Configuration - set of parameters to controll the experiment.
type Configuration struct {
	// Given SLO for the experiment.
	SLO int
	// Tunning time in [s].
	TuningTimeout int
	// Each measurement duration in [s].
	LoadDuration int
	// Number of load points to test.
	LoadPointsCount int
}

// Experiment is handler structure for Experiment Driver. All fields shall be
// not visible to the experimenter.
type Experiment struct {
	// Latency Sensitivity (Production) workload.
	pr workloads.Launcher
	// Workload Generator for Latency Sensitivity task.
	lgForPr workloads.LoadGenerator
	// Aggressors (Best Effort) tasks to stress LC task
	bes []workloads.Launcher
	// Experiement configuration
	conf Configuration
	// Max load with SLO not violated.
	targetLoad int
	// Result SLIs in a [aggressor][loadpoint] layout.
	slis [][]int
}

// NewExperiment construct new Experiment object.
// Input parameters:
// configuration - Experiment configuration
// productionTaskLauncher - Latency Critical job launcher
// loadGeneratorForProductionTask - stresser for production task
// aggressorTasksLaucher - Best Effort jobs launcher
func NewExperiment(
	configuration Configuration,
	productionTaskLauncher workloads.Launcher,
	loadGeneratorForProductionTask workloads.LoadGenerator,
	aggressorTaskLaunchers []workloads.Launcher) *Experiment {

	// TODO(mpatelcz): Validate configuration.

	return &Experiment{
		pr:      productionTaskLauncher,
		lgForPr: loadGeneratorForProductionTask,
		bes:     aggressorTaskLaunchers,
		conf:    configuration,
	}
}

// Runs single measurement of PR workload with given aggressor.
// Takes aggressor workload and specific loadPoint (rps)
// Return (sli, nil) on success (0, error) otherwise.
func (e *Experiment) runMeasurement(
	aggressor workloads.Launcher,
	qps int) (sli int, err error) {

	prTask, err := e.pr.Launch()
	if err != nil {
		return 0, err
	}
	defer prTask.Stop()

	agrTask, err := aggressor.Launch()
	if err != nil {
		return 0, err
	}
	sli, err = e.lgForPr.Load(qps, e.conf.LoadDuration)

	agrTask.Stop()
	return sli, err
}

func (e *Experiment) runPhase(
	aggressor workloads.Launcher) (slis []int, err error) {
	slis = make([]int, e.conf.LoadPointsCount)

	loadStep := int(e.targetLoad / e.conf.LoadPointsCount)

	for load := loadStep; load < e.targetLoad; load += loadStep {
		result, err := e.runMeasurement(aggressor, load)
		if err != nil {
			return nil, err
		}
		slis = append(slis, result)
	}
	return slis, err
}

func (e *Experiment) runTuning() error {
	prTask, err := e.pr.Launch()
	if err != nil {
		return err
	}
	defer prTask.Stop()

	e.targetLoad, err = e.lgForPr.Tune(e.conf.SLO, e.conf.TuningTimeout)
	if err != nil {
		e.targetLoad = -1
		return err
	}
	return err
}

// PrintSensitivityProfile prints in a user friendly form Experiement's
// resutls.
func (e *Experiment) PrintSensitivityProfile() error {
	return nil
}

// Run runs experiment.
// In the end it prints results to the standard output.
func (e *Experiment) Run() error {
	var err error

	err = e.runTuning()
	if err != nil {
		return err
	}

	for _, aggressor := range e.bes {
		slis, err := e.runPhase(aggressor)
		if err != nil {
			return err
		}
		e.slis = append(e.slis, slis)
	}

	//That's TBD.
	err = e.PrintSensitivityProfile()
	return err
}
