package sensitivity

import (
	"errors"
	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/intelsdi-x/swan/pkg/snap"
	"strconv"
	"strings"
	"time"
)

// measurementPhase performs a measurement for given loadPointIndex.
// User of this Phase is responsible to orchestrate this phase for each loadPoint to
// fulfill the LoadPointsCount.
type measurementPhase struct {
	namePrefix string
	// Latency Sensitivity (Production) workload.
	pr LauncherWithCollection
	// Workload Generator for Latency Sensitivity task.
	lgForPr LoadGeneratorWithCollection
	// Aggressors (Best Effort) tasks to stress LC task.
	// It can be empty in case of baseline phase.
	bes []LauncherWithCollection
	// Measurement duration in [s].
	loadDuration time.Duration
	// Number of load points to test.
	loadPointsCount int
	// Number of repetitions
	repetitions uint
	// Current measurement's load point.
	currentLoadPointIndex int

	// Shared reference for measurement targetQPS resulted from Tuning Phase.
	TargetLoad *int

	deferredCollectionHandlesToClose []snap.SessionHandle
	deferredTasksToStop              []executor.TaskHandle
	deferredTasksToClean             []executor.TaskHandle
}

// Returns measurement name.
func (m *measurementPhase) Name() string {
	return m.namePrefix + "_measurement_for_loadpoint_id_" +
		strconv.Itoa(m.currentLoadPointIndex)
}

// Returns number of repetitions.
func (m *measurementPhase) Repetitions() uint {
	return m.repetitions
}

// Gets current loadPoint from linear function y = a * x where `x` is loadPointIndex.
func (m *measurementPhase) getLoadPointUsingLinearFunction() int {
	// Since we know that the function is satisfied
	// when TargetLoadPoint = a * loadPointsCount, we can calculate `a` parameter.
	a := float64(*m.TargetLoad) / float64(m.loadPointsCount)
	x := float64(m.currentLoadPointIndex)

	return int(a * x)
}

func (m *measurementPhase) closeAllHandles() error {
	var err error
	errMsg := ""
	for _, task := range m.deferredTasksToStop {
		err = task.Stop()
		if err != nil {
			errMsg += " Error while stopping task: " + err.Error()
		}
	}
	m.deferredTasksToStop = []executor.TaskHandle{}

	for _, task := range m.deferredTasksToClean {
		err = task.Clean()
		if err != nil {
			errMsg += " Error while cleaning task: " + err.Error()
		}
	}
	m.deferredTasksToClean = []executor.TaskHandle{}

	for _, collection := range m.deferredCollectionHandlesToClose {
		// NOTE: Collection needs to ensure inside if it completed its work.
		err = collection.Stop()
		if err != nil {
			errMsg += " Error while stopping Snap session: " + err.Error()
		}
	}
	m.deferredCollectionHandlesToClose = []snap.SessionHandle{}

	if strings.Compare(errMsg, "") != 0 {
		return errors.New(errMsg)
	}

	return nil
}

// Run runs a measurement for given loadPointIndex.
func (m *measurementPhase) Run(session phase.Session) error {
	if m.TargetLoad == nil {
		return errors.New("Target QPS for measurement was not given.")
	}

	errMsg := ""
	err := m.runMeasurementScenario(session)
	if err != nil {
		errMsg += " Error while running measurement: " + err.Error()
	}

	// Make sure that deferred stops and cleans are executed.
	err = m.closeAllHandles()
	if err != nil {
		errMsg += " " + err.Error()
	}

	if strings.Compare(errMsg, "") != 0 {
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (m *measurementPhase) launchCollectionSession(taskInfo executor.TaskInfo,
	session phase.Session, launcher snap.SessionLauncher) error {
	// Check if Snap Session is specified.
	if launcher != nil {
		// Launch specified Snap Session.
		collectionHandle, err := launcher.Launch(taskInfo, session)
		if err != nil {
			return err
		}

		// Defer stopping launched Snap Session.
		m.deferredCollectionHandlesToClose =
			append(m.deferredCollectionHandlesToClose, collectionHandle)
	}

	return nil
}

func (m *measurementPhase) runMeasurementScenario(session phase.Session) error {
	// TODO:(bplotka): Here trigger Snap session for gathering platform metrics.

	// Launch Latency Sensitive workload.
	prTask, err := m.pr.Launcher.Launch()
	if err != nil {
		return err
	}
	// Defer stopping and closing the prTask.
	m.deferredTasksToStop = append(m.deferredTasksToStop, prTask)
	m.deferredTasksToClean = append(m.deferredTasksToClean, prTask)

	// Launch Snap Session for Latency Sensitive workload if specified.
	err = m.launchCollectionSession(prTask, session, m.pr.CollectionLauncher)
	if err != nil {
		return err
	}

	// Launch specified aggressors if any.
	for _, be := range m.bes {
		beTask, err := be.Launcher.Launch()
		if err != nil {
			return err
		}
		// Defer stopping and closing the beTask.
		m.deferredTasksToStop = append(m.deferredTasksToStop, beTask)
		m.deferredTasksToClean = append(m.deferredTasksToClean, beTask)

		// Launch Snap Session for be workload if specified.
		err = m.launchCollectionSession(beTask, session, be.CollectionLauncher)
		if err != nil {
			return err
		}
	}

	// NOTE: We can specify here different functions for specifying loadPoints distribution.
	loadPoint := m.getLoadPointUsingLinearFunction()

	log.Debug("Launching Load Generator with load ", loadPoint)
	loadGeneratorTask, err := m.lgForPr.LoadGenerator.Load(loadPoint, m.loadDuration)
	if err != nil {
		return err
	}

	// Defer closing the loadGeneratorTask.
	m.deferredTasksToClean = append(m.deferredTasksToClean, loadGeneratorTask)

	// Launch Snap Session for loadGenerator if specified.
	err = m.launchCollectionSession(loadGeneratorTask, session, m.lgForPr.CollectionLauncher)
	if err != nil {
		return err
	}

	//log.Debug("LoadGenerator filename: ", stdoutFile.Name())
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
