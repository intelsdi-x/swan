package loadgenerator

import (
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/pkg/errors"
)

const (
	defaultControllerIP = "127.0.0.1"
	defaultTxlCount     = 1 //default number of Transaction Injector components
)

var (
	// PathFlag specifies path to a SPECjbb2015 jar file.
	PathFlag = conf.NewStringFlag("specjbb_path", "Path to SPECjbb jar", path.Join(fs.GetSwanWorkloadsPath(), "workloads/web_serving/specjbb/specjbb2015.jar"))

	// IPFlag specifies IP address of a controller component of SPECjbb2015 benchmark.
	IPFlag = conf.NewIPFlag("specjbb_controller_ip", "IP address of a SPECjbb controller component", defaultControllerIP)

	// TxlCountFlag specifies number of Transaction Injector components in one group.
	TxlCountFlag = conf.NewIntFlag("specjbb_txl_count", "Number of Transaction injectors run in a cluster", defaultTxlCount)

	// ControllerHostProperty string name for property that specifies controller host.
	ControllerHostProperty = " -Dspecjbb.controller.host="

	// ControllerTypeProperty string name for property that specifies controller type.
	ControllerTypeProperty = " -Dspecjbb.controller.type="

	// InjectionRateProperty string name for property that specifies ir.
	InjectionRateProperty = " -Dspecjbb.controller.preset.ir="

	// PresetDurationProperty string name for property that specifies preset duration.
	PresetDurationProperty = " -Dspecjbb.controller.preset.duration="
)

type loadGenerator struct {
	controller           executor.Executor
	transactionInjectors []TxI
}

// NewLoadGenerator returns a new SPECjbb Load Generator instance composed of controller
// and transaction injectors.
// Transaction Injector and Controller are load generator for SPECjbb Backend.
// https://www.spec.org/jbb2015/docs/userguide.pdf
func NewLoadGenerator(controller executor.Executor, transactionInjectors []TxI) executor.LoadGenerator {
	return loadGenerator{
		controller:           controller,
		transactionInjectors: transactionInjectors,
	}
}

func stopTransactionInjectors(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.Stop()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

func cleanTransactionInjectors(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.Clean()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

func eraseTransactionInjectors(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.EraseOutput()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

func (loadGenerator loadGenerator) runTransactionInjectors() ([]executor.TaskHandle, error) {
	handles := []executor.TaskHandle{}
	for _, txI := range loadGenerator.transactionInjectors {
		command := getTxICommand(PathFlag.Value(), txI.Conf)
		handle, err := txI.Exec.Execute(command)
		if err != nil {
			logrus.Errorf("Could not start TransactionInjector with command %s", command)
			stopTransactionInjectors(handles)
			cleanTransactionInjectors(handles)
			eraseTransactionInjectors(handles)
			return nil, err
		}
		handles = append(handles, handle)
	}

	return handles, nil
}

// Populate does not do nothing here.
func (loadGenerator loadGenerator) Populate() (err error) {
	return nil
}

// Tune returns the High Bound Injection Rate achieved at maximum machine capacity.
func (loadGenerator loadGenerator) Tune(slo int) (qps int, achievedSLI int, err error) {
	return qps, achievedSLI, err
}

// Load starts a load on the specific workload with the defined loadPoint (injection rate value).
// The task will do the load for specified amount of time.
func (loadGenerator loadGenerator) Load(injectionRate int, duration time.Duration) (executor.TaskHandle, error) {
	txIHandles, err := loadGenerator.runTransactionInjectors()
	if err != nil {
		return nil, err
	}
	loadCommand := getControllerLoadCommand(PathFlag.Value(), injectionRate, duration)
	controllerHandle, err := loadGenerator.controller.Execute(loadCommand)
	if err != nil {
		stopTransactionInjectors(txIHandles)
		return nil, errors.Wrapf(err, "execution of SPECjbb Load Generator failed. command: %q", loadCommand)
	}

	return executor.NewClusterTaskHandle(controllerHandle, txIHandles), nil
}
