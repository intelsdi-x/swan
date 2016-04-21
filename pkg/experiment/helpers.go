package experiment

import (
	"os"
	"strconv"
)

// mkExperimentDir creates unique directory for Experiment logs and results.
func (e *Experiment) mkExperimentDir() error {
	e.startingDirectory, _ = os.Getwd()
	if len(e.conf.WorkingDirectory) > 0 {
		e.experimentDirectory = e.conf.WorkingDirectory + "/" + e.Session.Name
	} else {
		e.experimentDirectory = e.startingDirectory + "/" + e.Session.Name
	}
	err := os.MkdirAll(e.experimentDirectory, 0777)
	if err != nil {
		return err
	}

	err = os.Chdir(e.experimentDirectory)
	return err
}
func (e *Experiment) exitExperimentDir() {
	os.Chdir(e.startingDirectory)
}

func (e *Experiment) finalize() {
	e.logClose()
	e.exitExperimentDir()
}

func (e *Experiment) mkPhaseDir(phase Phase, iteration int) error {
	phaseDir := e.experimentDirectory + "/" +
		phase.Name() + "/" + strconv.FormatInt(int64(iteration), 10)
	err := os.MkdirAll(phaseDir, 0777)
	if err != nil {
		return err
	}
	err = os.Chdir(phaseDir)
	if err != nil {
		return err
	}
	return err
}
