package sensitivity

import "github.com/intelsdi-x/swan/pkg/workloads"

type Configuration struct {
	SLO             int
	TuningTimeout   int
	LoadDuration    int
	LoadPointsCount int
}

type Experiment struct {
	// Latency Sensitivity workload
	pr workloads.Launcher
	// Workload Generator for Latency Sensitivity workload
	lgForPr workloads.LoadGenerator
	// Aggressors to be but aside to LC workload
	be         []workloads.Launcher
	ec         Configuration
	targetLoad int
	slis       [][]int
}

// Construct new Experiment object.
func NewExperiment(
	configuration Configuration,
	productionTaskLauncher workloads.Launcher,
	loadGeneratorForProductionTask workloads.LoadGenerator,
	aggressorTaskLaunchers []workloads.Launcher) *Experiment {

	// TODO(mpatelcz): Validate configuration.

	return &Experiment{
		pr:      productionTaskLauncher,
		lgForPr: loadGeneratorForProductionTask,
		be:      aggressorTaskLaunchers,
		ec:      configuration,
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
	sli, err = e.lgForPr.Load(qps, e.ec.LoadDuration)

	agrTask.Stop()
	return sli, err
}

func (e *Experiment) runPhase(
	aggressor workloads.Launcher) (slis []int, err error) {
	slis = make([]int, e.ec.LoadPointsCount)

	loadStep := int(e.targetLoad / e.ec.LoadPointsCount)

	for load := loadStep; load < e.targetLoad; load += loadStep {
		result, err := e.runMeasurement(aggressor, load)
		if err != nil {
			return nil, err
		}
		slis = append(slis, result)
	}
	return slis, err
}

func (e *Experiment) runTunning() error {
	prTask, err := e.pr.Launch()
	if err != nil {
		return err
	}
	defer prTask.Stop()

	e.targetLoad, err = e.lgForPr.Tune(e.ec.SLO, e.ec.TuningTimeout)
	if err != nil {
		e.targetLoad = -1
		return err
	}
	return err
}

// Prints nice output. TBD
func (e *Experiment) PrintSensitivityProfile() error {
	return nil
}

func (e *Experiment) Run() error {
	var err error

	err = e.runTunning()
	if err != nil {
		return err
	}

	for _, aggressor := range e.be {
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
