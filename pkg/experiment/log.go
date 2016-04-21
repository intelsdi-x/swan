package experiment

import (
	"errors"
	"log"
	"os"
)

func (e *Experiment) logInitialize() error {
	var err error

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
	e.logger.Println("Starting experiment with uuid: ", e.session.Name)
	return err
}

func (e *Experiment) logClose() {
	e.logger.SetOutput(os.Stdout)
	e.logFile.Close()
}
