package sensitivity

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/workloads"
	"strconv"
	"time"
)

// measurementPhase performs a measurement for given loadPointIndex.
// User of this Phase is responsible to orchestrate this phase for each loadPoint to
// fulfill the LoadPointsCount.
type measurementPhase struct {
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

	// Shared reference for measurement targetQPS resulted from Tuning Phase.
	TargetLoad *int
}

// Returns measurement name.
func (m *measurementPhase) Name() string {
	return m.namePrefix + "_measurement_for_loadpoint_id_" +
		strconv.Itoa(m.currentLoadPointIndex)
}

// Returns number of repetitions.
func (m *measurementPhase) Repetitions() int {
	return m.repetitions
}

func (m *measurementPhase) getLoadPoint() int {
	return int(float64(m.currentLoadPointIndex) *
		(float64(*m.TargetLoad) / float64(m.loadPointsCount)))
}

// Run runs a measurement for given loadPointIndex.
func (m *measurementPhase) Run() error {
	if m.TargetLoad == nil {
		return errors.New("Target QPS for measurement was not given.")
	}
	// TODO:(bplotka): Here trigger Snap session for gathering platform metrics.

	// Launch Latency Sensitive workload.
	prTask, err := m.pr.Launch()
	if err != nil {
		return err
	}
	defer prTask.Stop()
	defer prTask.Clean()

	// TODO:(bplotka): Here trigger Snap session for fetching SLI if it is supported for Prod Task.

	// Launch specified aggressors if any.
	for _, be := range m.bes {
		beTask, err := be.Launch()
		if err != nil {
			return err
		}
		// These defers copy reference to the Task Handle and
		// defers it to the end of Run function.
		defer beTask.Stop()
		defer beTask.Clean()

		// TODO:(bplotka): Here trigger Snap session for fetching SLI
		// if it is supported for aggressor Task.
	}

	loadPoint := m.getLoadPoint()

	log.Debug("Launching Load Generator with load ", loadPoint)
	loadGeneratorTask, err := m.lgForPr.Load(loadPoint, m.loadDuration)
	if err != nil {
		return err
	}

	defer loadGeneratorTask.Clean()

	stdoutFile, err := loadGeneratorTask.StdoutFile()
	if err != nil {
		return err
	}

	log.Debug("LoadGenerator filename: ", stdoutFile.Name())
	// TODO:(bplotka): Here trigger Snap session for fetching SLI.

	// Wait for load generation to end.
	loadGeneratorTask.Wait(0)

	// Check status of load generation.
	exitCode, err := loadGeneratorTask.ExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		// Load generator failed.
		return fmt.Errorf("Executing Mutilate Load returned with exit code %d", exitCode)
	}

	return nil
}

// Finalize is executed after all repetitions of given measurement.
func (m *measurementPhase) Finalize() error {
	// All data should be aggregated in Snap. So nothing to do here.
	return nil
}
