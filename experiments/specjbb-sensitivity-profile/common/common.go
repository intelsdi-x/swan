package common

import (
	"fmt"
	"os"
	"path"
	"strconv"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	"github.com/pkg/errors"
)

// CreateRepetitionDir creates folders that store repetition logs inside experiment's directory.
func CreateRepetitionDir(experimentDirectory, phaseName string, repetition int) (string, error) {
	repetitionDir := path.Join(experimentDirectory, phaseName, strconv.Itoa(repetition))
	err := os.MkdirAll(repetitionDir, 0777)
	if err != nil {
		return "", errors.Wrapf(err, "could not create dir %q", repetitionDir)
	}

	err = os.Chdir(repetitionDir)
	if err != nil {
		return "", errors.Wrapf(err, "could not change to dir %q", repetitionDir)
	}

	return repetitionDir, nil
}

// CreateExperimentDir creates directory structure for the experiment.
func CreateExperimentDir(uuid string) (experimentDirectory string, logFile *os.File, err error) {
	experimentDirectory = path.Join(os.TempDir(), conf.AppName(), uuid)
	err = os.MkdirAll(experimentDirectory, 0777)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "cannot create experiment directory: ", experimentDirectory)
	}
	err = os.Chdir(experimentDirectory)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "cannot chdir to experiment directory", experimentDirectory)
	}

	masterLogFilename := path.Join(experimentDirectory, "master.log")
	logFile, err = os.OpenFile(masterLogFilename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "could not open log file %q", masterLogFilename)
	}

	return experimentDirectory, logFile, nil
}

// PrepareSnapSpecjbbSessionLauncher prepares a SessionLauncher that runs SPECjbb collector and records that into storage.
// TODO: this should be put into swan:/pkg/snap
func PrepareSnapSpecjbbSessionLauncher() (snap.SessionLauncher, error) {
	// NOTE: For debug it is convenient to disable snap for some experiment runs.
	if snap.SnapteldHTTPEndpoint.Value() != "none" {
		// Create connection with Snap.
		logrus.Info("Connecting to Snapteld on ", snap.SnapteldHTTPEndpoint.Value())
		specjbbConfig := specjbbsession.DefaultConfig()
		specjbbConfig.SnapteldAddress = snap.SnapteldHTTPEndpoint.Value()
		specjbbSnapSession, err := specjbbsession.NewSessionLauncher(specjbbConfig)
		if err != nil {
			return nil, err
		}
		return specjbbSnapSession, nil
	}
	return nil, fmt.Errorf("snap http endpoint is not present, cannot prepare SPECjbb session launcher")
}

// PrepareSpecjbbLoadGenerator creates new LoadGenerator based on specjbb.
func PrepareSpecjbbLoadGenerator(ip string) (executor.LoadGenerator, error) {
	var loadGeneratorExecutor executor.Executor
	var transactionInjectors []executor.Executor
	txICount := specjbb.TxICountFlag.Value()
	if ip != "127.0.0.1" {
		var err error
		loadGeneratorExecutor, err = sensitivity.NewRemote(ip)
		if err != nil {
			return nil, err
		}
		for i := 1; i <= txICount; i++ {
			transactionInjector, err := sensitivity.NewRemote(ip)
			if err != nil {
				return nil, err
			}
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	} else {
		loadGeneratorExecutor = executor.NewLocal()
		for i := 1; i <= txICount; i++ {
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	}

	specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()
	specjbbLoadGeneratorConfig.ControllerIP = ip
	specjbbLoadGeneratorConfig.TxICount = txICount

	loadGeneratorLauncher := specjbb.NewLoadGenerator(loadGeneratorExecutor,
		transactionInjectors, specjbbLoadGeneratorConfig)

	return loadGeneratorLauncher, nil
}

// GetPeakLoad runs tuning in order to determine the peak load.
func GetPeakLoad(hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator, slo int) (peakLoad int, err error) {
	prTask, err := hpLauncher.Launch()
	if err != nil {
		return 0, errors.Wrap(err, "cannot launch specjbb backend")
	}
	defer func() {
		// If function terminated with error then we do not want to overwrite it with any errors in defer.
		errStop := prTask.Stop()
		if err == nil {
			err = errStop
		}
		prTask.Clean()
	}()

	err = loadGenerator.Populate()
	if err != nil {
		return 0, errors.Wrap(err, "cannot populate memcached")
	}

	peakLoad, _, err = loadGenerator.Tune(slo)
	if err != nil {
		return 0, errors.Wrap(err, "tuning failed")
	}

	return
}
