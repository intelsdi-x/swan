package common

import (
	"fmt"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	"github.com/pkg/errors"
)

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
		loadGeneratorExecutor, err = executor.NewRemoteFromIP(ip)
		if err != nil {
			return nil, err
		}
		for i := 1; i <= txICount; i++ {
			transactionInjector, err := executor.NewRemoteFromIP(ip)
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
