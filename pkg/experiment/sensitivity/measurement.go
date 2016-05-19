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
	pr LauncherSessionPair
	// Workload Generator for Latency Sensitivity task.
	lgForPr LoadGeneratorSessionPair
	// Aggressors (Best Effort) tasks to stress LC task.
	// It can be empty in case of baseline phase.
	bes []LauncherSessionPair
	// Measurement duration in [s].
	loadDuration time.Duration
	// Number of load points to test.
	loadPointsCount int
	// Number of repetitions
	repetitions int
	// Current measurement's load point.
	currentLoadPointIndex int

	// Shared reference for measurement targetQPS resulted from Tuning Phase.
	PeakLoad *int

	activeSnapSessions       []snap.SessionHandle
	activeLaunchersTasks     []executor.TaskHandle
	activeLoadGeneratorTasks []executor.TaskHandle
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

// Gets current loadPoint from linear function y = a * x where `x` is loadPointIndex.
func (m *measurementPhase) getLoadPoint() int {
	// Since we know that the function is satisfied
	// when TargetLoadPoint = a * loadPointsCount, we can calculate `a` parameter.
	a := float64(*m.PeakLoad) / float64(m.loadPointsCount)
	x := float64(m.currentLoadPointIndex)

	return int(a * x)
}

func (m *measurementPhase) clean() error {
	var err error
	errMsg := ""
	// Cleaning and stopping active Launchers' tasks.
	for _, task := range m.activeLaunchersTasks {
		err = task.Stop()
		if err != nil {
			errMsg += " Error while stopping task: " + err.Error()
			// Don't clean when stop failed.
			continue
		}

		err = task.Clean()
		if err != nil {
			errMsg += " Error while cleaning task: " + err.Error()
		}
	}
	m.activeLaunchersTasks = []executor.TaskHandle{}

	// Cleaning only active LoadGenerators' tasks.
	for _, task := range m.activeLoadGeneratorTasks {
		err = task.Clean()
		if err != nil {
			errMsg += " Error while cleaning task: " + err.Error()
		}
	}
	m.activeLoadGeneratorTasks = []executor.TaskHandle{}

	// Stopping only active Snap sessions.
	for _, snapSession := range m.activeSnapSessions {
		// NOTE: snapSession needs to ensure inside if it completed its work.
		err = snapSession.Stop()
		if err != nil {
			errMsg += " Error while stopping Snap session: " + err.Error()
		}
	}
	m.activeSnapSessions = []snap.SessionHandle{}

	if strings.Compare(errMsg, "") != 0 {
		return errors.New(errMsg)
	}

	return nil
}

// Run runs a measurement for given loadPointIndex.
func (m *measurementPhase) Run(session phase.Session) error {
	if m.PeakLoad == nil {
		return errors.New("Target QPS for measurement was not given.")
	}

	errMsg := ""
	err := m.run(session)
	if err != nil {
		errMsg += " Error while running measurement: " + err.Error()
	}

	// Make sure that deferred stops and cleans are executed.
	err = m.clean()
	if err != nil {
		errMsg += " " + err.Error()
	}

	if strings.Compare(errMsg, "") != 0 {
		log.Errorf(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (m *measurementPhase) launchSnapSession(taskInfo executor.TaskInfo,
	session phase.Session, launcher snap.SessionLauncher) error {
	// Check if Snap Session is specified.
	if launcher != nil {
		// Launch specified Snap Session.
		sessionHandle, err := launcher.LaunchSession(taskInfo, session)
		if err != nil {
			return err
		}

		// Defer stopping launched Snap Session.
		m.activeSnapSessions = append(m.activeSnapSessions, sessionHandle)
	}

	return nil
}

func (m *measurementPhase) run(session phase.Session) error {
	// TODO:(bplotka): Here trigger Snap session for gathering platform metrics.

	// Launch Latency Sensitive workload.
	prTask, err := m.pr.Launcher.Launch()
	if err != nil {
		return err
	}
	// Defer stopping and closing the prTask.
	m.activeLaunchersTasks = append(m.activeLaunchersTasks, prTask)

	// Launch Snap Session for Latency Sensitive workload if specified.
	err = m.launchSnapSession(prTask, session, m.pr.SnapSessionLauncher)
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
		m.activeLaunchersTasks = append(m.activeLaunchersTasks, beTask)

		// Launch Snap Session for be workload if specified.
		err = m.launchSnapSession(beTask, session, be.SnapSessionLauncher)
		if err != nil {
			return err
		}
	}

	// TODO(bp): Push that to DB via Snap in tag or using SwanCollector.
	loadPoint := m.getLoadPoint()

	log.Debug("Launching Load Generator with load ", loadPoint)
	loadGeneratorTask, err := m.lgForPr.LoadGenerator.Load(loadPoint, m.loadDuration)
	if err != nil {
		return err
	}

	// Defer cleaning the loadGeneratorTask.
	m.activeLoadGeneratorTasks = append(m.activeLoadGeneratorTasks, loadGeneratorTask)

	// Launch Snap Session for loadGenerator if specified.
	err = m.launchSnapSession(loadGeneratorTask, session, m.lgForPr.SnapSessionLauncher)
	if err != nil {
		return err
	}

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
