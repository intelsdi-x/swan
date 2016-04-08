package ExperimentDriver

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/provisioning"
)

//TBD: STUB before it's defined in provisioning
type Launcher interface {
	Launch() (provisioning.Task, error)
}

//TBD: STUB before it's defined in provisioning
type LoadGenerator interface {
	Tune(slo int, timeout int) (targetQPS int, e error)
	Load(rps int, duration int) (sli int, e error)
}

type ExperimentConfiguration struct {
	SLO             int
	TuningTimeout   int
	LoadDuration    int
	LoadPointsCount int
}

type SensitivitiProfileExperiment struct {
	// Latency Sensitivity workload
	pr Launcher
	// Workload Generator for Latency Sensitivity workload
	lgForPr LoadGenerator
	// Aggressors to be but aside to LC workload
	be         []Launcher
	ec         ExperimentConfiguration
	targetLoad int
	slis       [][]int
}

// Construct new SensitivityProfileExperiment object.
func NewSensitivitiProfileExperiment(c ExperimentConfiguration,
	pr Launcher,
	lgForPr LoadGenerator,
	be []Launcher) *SensitivitiProfileExperiment {

	return &SensitivitiProfileExperiment{pr: pr, lgForPr: lgForPr,
		be: be, ec: c}
}

// Runs single measurement of PR workload with given aggressor.
// Takes aggressor workload and specific loadPoint (rps)
// Return (sli, nil) on success (0, error) otherwise.
func (e *SensitivitiProfileExperiment) runMeasurement(
	aggressor Launcher,
	qps int) (sli int, err error) {

	prTask, err := e.pr.Launch()
	if err != nil {
		return 0, err
	}
	defer prTask.Stop()

	//Run aggressor
	agrTask, err := aggressor.Launch()
	if err != nil {
		return 0, err
	}
	//Run workload generator - blocking?
	//TBD: output? Raw data?
	sli, err = e.lgForPr.Load(qps, e.ec.LoadDuration)

	agrTask.Stop()
	return sli, err
}

// Executes single phase
func (e *SensitivitiProfileExperiment) runPhase(aggressor Launcher) (slis []int, err error) {
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

func (e *SensitivitiProfileExperiment) runTunning() error {
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
func (e *SensitivitiProfileExperiment) PrintSensitivityProfile() error {
	fmt.Print("SensitivityProfile is TBD")
	return nil
}

func (e *SensitivitiProfileExperiment) Run() error {
	var err error

	err = e.runTunning()
	if err != nil {
		//Stop here
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
