package experiment

import (
	"errors"
	"log"
	"os"
	"strconv"
)

func (e *Experiment) logInitialize() error {

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
	if err != nil {
		return err
	}

	// create master log file "ExperimentDriver.log"
	e.logFile, err = os.Create(e.experimentDirectory + "/Master.log")
	if err != nil {
		os.Chdir(e.startingDirectory)
		return err
	}
	e.logger = log.New(e.logFile, "", log.LstdFlags)
	if e.logger == nil {
		os.Chdir(e.startingDirectory)
		return errors.New("Failed to create master log file")
	}
	e.logger.Println("Starting experiment with uuid: ", e.Session.Name)
	return err
}

func (e *Experiment) logClose() {
	e.logger.SetOutput(os.Stdout)
	e.logFile.Close()
	os.Chdir(e.startingDirectory)
}

func (e *Experiment) logMkPhase(name string, iteration int) error {
	phaseDir := e.experimentDirectory + "/" +
		name + "_" + strconv.FormatInt(int64(iteration), 10)
	err := os.Mkdir(phaseDir, 0777)
	if err != nil {
		return err
	}
	err = os.Chdir(phaseDir)
	if err != nil {
		return err
	}
	return err
}
