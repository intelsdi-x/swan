package sensitivity

import (
	"fmt"
	"strconv"
	"strings"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/utils/err_collection"
	"github.com/intelsdi-x/swan/pkg/experiment/phase"
	"github.com/pkg/errors"
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
// TODO(bp): Change to UUID when completing SCE-376.
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

func (m *measurementPhase) clean() (err error) {
	var errCollection errcollection.ErrorCollection

	// Cleaning and stopping active Launchers' tasks.
	for _, task := range m.activeLaunchersTasks {
		err = task.Stop()
		if err != nil {
			errCollection.Add(errors.Wrap(err, "error while stopping task"))

			// Don't clean when stop failed.
			continue
		}

		err = task.Clean()
		if err != nil {
			errCollection.Add(errors.Wrap(err, "error while cleaning task"))
		}
	}
	m.activeLaunchersTasks = []executor.TaskHandle{}

	// Cleaning only active LoadGenerators' tasks.
	for _, task := range m.activeLoadGeneratorTasks {
		err = task.Clean()
		if err != nil {
			errCollection.Add(errors.Wrap(err, "error while cleaning task"))
		}
	}
	m.activeLoadGeneratorTasks = []executor.TaskHandle{}

	// Stopping only active Snap sessions.
	for _, snapSession := range m.activeSnapSessions {
		log.Debug("Waiting for snap session to complete it's work. ", snapSession)
		err = snapSession.Wait()
		if err != nil {
			errCollection.Add(errors.Wrap(err, "error while waiting for Snap session to complete it's work"))
		}

		err = snapSession.Stop()
		if err != nil {
			errCollection.Add(errors.Wrap(err, "error while  stopping Snap session"))
		}
	}
	m.activeSnapSessions = []snap.SessionHandle{}

	return errCollection.GetErrIfAny()
}

// Run runs a measurement for given loadPointIndex.
func (m *measurementPhase) Run(session phase.Session) error {
	if m.PeakLoad == nil {
		return errors.New("target QPS for measurement was not given")
	}

	// TODO(bp): Remove that when completing SCE-376
	session.LoadPointQPS = m.getLoadPoint()
	if len(m.bes) > 0 {
		session.AggressorName = ""
		for i, be := range m.bes {
			if i > 0 {
				session.AggressorName += "And"
			}

			session.AggressorName += be.Launcher.Name()
		}
	} else {
		session.AggressorName = "None"
	}

	errMsg := ""
	err := m.run(session)
	if err != nil {
		errMsg += " error while running measurement: " + err.Error()
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
		tags := createTagConfigItem(session)
		sessionHandle, err := launcher.LaunchSession(taskInfo, tags)
		if err != nil {
			return err
		}

		// Defer stopping launched Snap Session.
		m.activeSnapSessions = append(m.activeSnapSessions, sessionHandle)
	}

	return nil
}

// run in order:
// - start HP workload
// - populate HP (with data)
// - start aggressors workloads and their sessions (if any)
// - start HP session
// - start and wait to finish for "load generator"
// - start "load generator" session
func (m *measurementPhase) run(session phase.Session) error {
	// TODO(bp): Here trigger Snap session for gathering platform metrics.

	// Launch Latency Sensitive workload.
	prTask, err := m.pr.Launcher.Launch()
	if err != nil {
		return err
	}
	// Defer stopping and closing the prTask.
	m.activeLaunchersTasks = append(m.activeLaunchersTasks, prTask)

	log.Debug("Populating initial test data to LC task")
	err = m.lgForPr.LoadGenerator.Populate()
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

	// Launch Snap Session for Latency Sensitive workload if specified.
	err = m.launchSnapSession(prTask, session, m.pr.SnapSessionLauncher)
	if err != nil {
		return err
	}

	log.Debug("Launching Load Generator with load ", loadPoint)
	loadGeneratorTask, err := m.lgForPr.LoadGenerator.Load(loadPoint, m.loadDuration)
	if err != nil {
		return err
	}

	// Defer cleaning the loadGeneratorTask.
	m.activeLoadGeneratorTasks = append(m.activeLoadGeneratorTasks, loadGeneratorTask)

	// Wait for load generation to end.
	loadGeneratorTask.Wait(0)

	// Launch Snap Session for loadGenerator if specified.
	// NOTE: Common loadGenerators don't have HTTP and just save output to the file after
	// completing the load generation.
	// To have our snap task not disabled by Snap daemon because we could not read the file during
	// load, we need to run snap session (task) only after load generation work ended.
	err = m.launchSnapSession(loadGeneratorTask, session, m.lgForPr.SnapSessionLauncher)
	if err != nil {
		return err
	}

	// Check status of load generation.
	exitCode, err := loadGeneratorTask.ExitCode()
	if err != nil {
		return err
	}

	if exitCode != 0 {
		// Load generator failed.
		return errors.Errorf("executing Mutilate Load returned with exit code %d", exitCode)
	}

	return nil
}

// Finalize is executed after all repetitions of given measurement.
func (m *measurementPhase) Finalize() error {
	// All data should be aggregated in Snap. So nothing to do here.
	return nil
}

func createTagConfigItem(phaseSession phase.Session) string {
	// Constructing Tags config item as stated in
	// https://github.com/intelsdi-x/snap-plugin-processor-tag/README.md
	return fmt.Sprintf("%s:%s,%s:%s,%s:%d,%s:%d,%s:%s",
		phase.ExperimentKey, phaseSession.ExperimentID,
		phase.PhaseKey, phaseSession.PhaseID,
		phase.RepetitionKey, phaseSession.RepetitionID,
		// TODO: Remove that when completing SCE-376
		phase.LoadPointQPSKey, phaseSession.LoadPointQPS,
		phase.AggressorNameKey, phaseSession.AggressorName,
	)
}
