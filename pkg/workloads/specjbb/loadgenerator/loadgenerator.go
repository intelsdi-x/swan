package loadgenerator

import (
	"path"
	"time"

	"fmt"
	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/parser"
	"github.com/pkg/errors"
)

const (
	defaultControllerIP   = "127.0.0.1"
	defaultTxICount       = 1 //default number of Transaction Injector components
	defaultCustomerNumber = 100
	defaultProductsNumber = 100
)

var (
	// PathToBinaryFlag specifies path to a SPECjbb2015 jar file.
	PathToBinaryFlag = conf.NewStringFlag("specjbb_path", "Path to SPECjbb jar", path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "specjbb2015.jar"))

	// PathToPropsFileFlag specifies path to a SPECjbb2015 properties file.
	PathToPropsFileFlag = conf.NewStringFlag("specjbb_path", "Path to SPECjbb jar", path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "config", "specjbb2015.props"))

	// IPFlag specifies IP address of a controller component of SPECjbb2015 benchmark.
	IPFlag = conf.NewIPFlag("specjbb_controller_ip", "IP address of a SPECjbb controller component", defaultControllerIP)

	// TxICountFlag specifies number of Transaction Injector components in one group.
	TxICountFlag = conf.NewIntFlag("specjbb_txl_count", "Number of Transaction injectors run in a cluster", defaultTxICount)

	// CustomerNumberFlag specifies number of customers.
	CustomerNumberFlag = conf.NewIntFlag("specjbb_customer_number", "Number of customers", defaultCustomerNumber)

	// ProductNumberFlag specifies number of products.
	ProductNumberFlag = conf.NewIntFlag("specjbb_product_number", "Number of products", defaultProductsNumber)

	// ControllerHostProperty string name for property that specifies controller host.
	ControllerHostProperty = " -Dspecjbb.controller.host="

	// ControllerTypeProperty string name for property that specifies controller type.
	ControllerTypeProperty = " -Dspecjbb.controller.type="

	// InjectionRateProperty string name for property that specifies ir.
	InjectionRateProperty = " -Dspecjbb.controller.preset.ir="

	// PresetDurationProperty string name for property that specifies preset duration.
	PresetDurationProperty = " -Dspecjbb.controller.preset.duration="

	// CustomerNumberProperty represents total number of customers.
	CustomerNumberProperty = " -Dspecjbb.input.number_customers="

	// ProductNumberProperty represents total number of products.
	ProductNumberProperty = " -Dspecjbb.input.number_products="

	// BinaryDataOutputDir represents directory for storing binary log file of the run.
	BinaryDataOutputDir = " -Dspecjbb.run.datafile.dir="
)

// Config is a config for a SPECjbb2015 Transaction Injector,
// Supported options:
// IP - property "-Dspecjbb.controller.host=" - IP address of a SPECjbb controller component (default:127.0.0.1)
type Config struct {
	ControllerIP        string
	PathToBinary        string
	PathToProps         string
	TxICount            int
	CustomerNumber      int
	ProductNumber       int
	BinaryDataOutputDir string
}

// NewDefaultConfig is a constructor for Config with default parameters.
func NewDefaultConfig() Config {
	return Config{
		ControllerIP:        IPFlag.Value(),
		PathToBinary:        PathToBinaryFlag.Value(),
		PathToProps:         PathToPropsFileFlag.Value(),
		TxICount:            TxICountFlag.Value(),
		CustomerNumber:      CustomerNumberFlag.Value(),
		ProductNumber:       ProductNumberFlag.Value(),
		BinaryDataOutputDir: path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb"),
	}
}

type loadGenerator struct {
	controller           executor.Executor
	transactionInjectors []executor.Executor
	config               Config
}

// NewLoadGenerator returns a new SPECjbb Load Generator instance composed of controller
// and transaction injectors.
// Transaction Injector and Controller are load generator for SPECjbb Backend.
// https://www.spec.org/jbb2015/docs/userguide.pdf
func NewLoadGenerator(controller executor.Executor, transactionInjectors []executor.Executor, config Config) executor.LoadGenerator {
	return loadGenerator{
		controller:           controller,
		transactionInjectors: transactionInjectors,
		config:               config,
	}
}

// stopTransactionInjectors stops all tasks that run as transaction injectors.
func stopTransactionInjectors(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.Stop()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

// cleanTransactionInjectors closes the transaction injectors tasks stdout & stderr files.
func cleanTransactionInjectors(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.Clean()
		if err != nil {
			logrus.Error(err.Error())
		}
	}
}

// eraseTransactionInjectors removes transaction injectors tasks stdout & stderr files.
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
	for id, txI := range loadGenerator.transactionInjectors {
		command := getTxICommand(loadGenerator.config, id+1)
		handle, err := txI.Execute(command)
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

// Populate is not implemented.
func (loadGenerator loadGenerator) Populate() (err error) {
	logrus.Warn("Populate function is not implemented.")
	return nil
}

// Tune calculates maximum capacity of a machine without any time constraints.
// Then it builds RT curve (increase load from 1% of HBIR to 100%, step 1%).
// By using RT curve it calculates Geo-mean of (critical-jOPS@ 10ms, 25ms, 50ms, 75ms and 100ms response time SLAs).
// We use critical jops value as maximum capacity (HBIR) is high above our desired SLA.
// Exemplary output for machine capacity, HBIR = 12000:
// critical-jOPS = Geomean ( jOPS @ 10000; 25000; 50000; 75000; 100000; SLAs )
// Response time percentile is 99-th
// SLA (us)	10000	25000	50000	75000	100000	Geomean
// jOPS		1789	2588	2848	3080	3428	2684
func (loadGenerator loadGenerator) Tune(slo int) (qps int, achievedSLI int, err error) {
	hbirRtCommand := getControllerHBIRRTCommand(loadGenerator.config)
	controllerHandle, err := loadGenerator.controller.Execute(hbirRtCommand)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "execution of SPECjbb HBIR RT failed. command: %q", hbirRtCommand)
	}
	txIHandles, err := loadGenerator.runTransactionInjectors()
	if err != nil {
		return 0, 0, err
	}

	controllerHandle.Wait(0)
	stopTransactionInjectors(txIHandles)
	eraseTransactionInjectors(txIHandles)
	cleanTransactionInjectors(txIHandles)

	out, err := controllerHandle.StdoutFile()
	if err != nil {
		return 0, 0, err
	}
	rawFileName, err := parser.FileWithRawFileName(out.Name())
	if err != nil {
		return 0, 0, err
	}

	if rawFileName == "" {
		return 0, 0, fmt.Errorf("Could not get raw results file name from an output file %s", out.Name())
	}

	// Run reporter to calculate critical jops value from raw results.
	reporterCommand := getReporterCommand(loadGenerator.config, rawFileName)
	reporterHandle, err := loadGenerator.controller.Execute(reporterCommand)
	reporterHandle.Wait(0)

	out, err = reporterHandle.StdoutFile()
	if err != nil {
		return 0, 0, err
	}
	hbirRt, err := parser.FileWithHBIRRT(out.Name())
	if err != nil {
		return 0, 0, err
	}
	controllerHandle.EraseOutput()
	controllerHandle.Clean()
	reporterHandle.EraseOutput()
	reporterHandle.Clean()

	return hbirRt, 0, err
}

// Load starts a load on the specific workload with the defined loadPoint (injection rate value).
// The task will do the load for specified amount of time.
func (loadGenerator loadGenerator) Load(injectionRate int, duration time.Duration) (executor.TaskHandle, error) {
	loadCommand := getControllerLoadCommand(loadGenerator.config, injectionRate, duration.Nanoseconds())
	controllerHandle, err := loadGenerator.controller.Execute(loadCommand)
	if err != nil {
		return nil, errors.Wrapf(err, "execution of SPECjbb Load Generator failed. command: %q", loadCommand)
	}
	txIHandles, err := loadGenerator.runTransactionInjectors()
	if err != nil {
		return nil, err
	}

	return executor.NewClusterTaskHandle(controllerHandle, txIHandles), nil
}
