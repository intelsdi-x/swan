package sensitivity

import (
	"errors"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"github.com/montanaflynn/stats"
	"strconv"
	"time"
)

// measurementPhase performs a measurement for given loadPointIndex.
// User of this Phase is responsible to orchestrate this phase for each loadPoint to
// fulfill the LoadPointsCount.
type measurement struct {
	namePrefix string
	// Latency Sensitivity (Production) workload.
	pr workloads.Launcher
	// Workload Generator for Latency Sensitivity task.
	lgForPr workloads.LoadGenerator
	// Aggressors (Best Effort) tasks to stress LC task.
	// It can be empty in case of baseline phase.
	bes []workloads.Launcher
	// Measurement duration in [s].
	loadDuration time.Duration
	// Number of load points to test.
	loadPointsCount int
	// Number of repetitions
	repetitions int
	// Current measurement's load point.
	currentLoadPointIndex int

	// Shared context variables:

	// Shared reference for measurement targetQPS resulted from Tuning Phase.
	TargetLoad *float64
	// Results for this measurement for current loadPointIndex, across repetitions.
	// Load which was achieved during experiment e.g QPS, RPS.
	loadResults []float64
	// Service Level Indicator which was achieved during experiment e.g latency in us
	sliResults []float64

	// Shared reference.
	MeasurementSliResult float64
}

// Returns measurement name.
func (m *measurement) Name() string {
	return m.namePrefix + " Measurement for LoadPointIndex " +
		strconv.Itoa(m.currentLoadPointIndex)
}

// Returns number of repetitions.
func (m *measurement) Repetitions() int {
	return m.repetitions
}

func (m *measurement) getLoadPoint() int {
	return int(float64(m.currentLoadPointIndex) * (*m.TargetLoad / float64(m.loadPointsCount)))
}

// Run runs a measurement for given loadPointIndex.
func (m *measurement) Run(log *logrus.Logger) error {
	if m.TargetLoad == nil {
		return errors.New("Target QPS for measurement was not given.")
	}

	prTask, err := m.pr.Launch()
	if err != nil {
		return err
	}
	defer prTask.Stop()
	defer prTask.Clean()

	// Launch specified aggressors if any.
	for _, be := range m.bes {
		beTask, err := be.Launch()
		if err != nil {
			return err
		}
		// This defer copies reference to the Task Handle and
		// defers it to the end of Run function.
		defer beTask.Stop()
		defer beTask.Clean()
	}

	loadPoint := m.getLoadPoint()

	log.Debug("Launching Load Generator with load ", loadPoint)
	achievedLoad, achievedSLI, err := m.lgForPr.Load(loadPoint, m.loadDuration)
	if err != nil {
		return err
	}

	// Save results.
	m.sliResults = append(m.sliResults, float64(achievedSLI))
	m.loadResults = append(m.loadResults, float64(achievedLoad))

	return nil
}

// Finalize is executed after all repetitions of given measurement.
func (m *measurement) Finalize() error {
	// TODO: Check if variance is not too high.

	var err error
	// Calculate average.
	m.MeasurementSliResult, err = stats.Mean(m.sliResults)
	if err != nil {
		m.MeasurementSliResult = -1
		return err
	}

	return nil
}
