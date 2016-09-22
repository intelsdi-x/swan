package loadgenerator

import (
	"fmt"
	"path"
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/parser"
	"github.com/pkg/errors"
)

const (
	defaultControllerIP   = "127.0.0.1"
	defaultTxICount       = 1 // Default number of Transaction Injector components.
	defaultCustomerNumber = 100
	defaultProductsNumber = 100
)

var (
	// PathToBinaryFlag specifies path to a SPECjbb2015 jar file.
	PathToBinaryFlag = conf.NewStringFlag("specjbb_path", "Path to SPECjbb jar", path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "specjbb2015.jar"))

	// PathToPropsFileFlag specifies path to a SPECjbb2015 properties file.
	PathToPropsFileFlag = conf.NewStringFlag("specjbb_props_path", "Path to SPECjbb properties file", path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "config", "specjbb2015.props"))

	// PathToOutputTemplateFlag specifies path to a SPECjbb2015 output template file.
	PathToOutputTemplateFlag = conf.NewStringFlag("specjbb_output_template_path", "Path to SPECjbb output template file", path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb", "config", "template-D.raw"))

	// IPFlag specifies IP address of a controller component of SPECjbb2015 benchmark.
	IPFlag = conf.NewIPFlag("specjbb_controller_ip", "IP address of a SPECjbb controller component", defaultControllerIP)

	// TxICountFlag specifies number of Transaction Injector components in one group.
	TxICountFlag = conf.NewIntFlag("specjbb_txl_count", "Number of Transaction injectors run in a cluster", defaultTxICount)

	// CustomerNumberFlag specifies number of customers.
	CustomerNumberFlag = conf.NewIntFlag("specjbb_customer_number", "Number of customers", defaultCustomerNumber)

	// ProductNumberFlag specifies number of products.
	ProductNumberFlag = conf.NewIntFlag("specjbb_product_number", "Number of products", defaultProductsNumber)

	// ControllerHostProperty - string name for property that specifies controller host.
	ControllerHostProperty = " -Dspecjbb.controller.host="

	// ControllerTypeProperty - string name for property that specifies controller type.
	ControllerTypeProperty = " -Dspecjbb.controller.type="

	// InjectionRateProperty - string name for property that specifies ir.
	InjectionRateProperty = " -Dspecjbb.controller.preset.ir="

	// PresetDurationProperty - string name for property that specifies preset duration.
	PresetDurationProperty = " -Dspecjbb.controller.preset.duration="

	// CustomerNumberProperty represents total number of customers.
	CustomerNumberProperty = " -Dspecjbb.input.number_customers="

	// ProductNumberProperty represents total number of products.
	ProductNumberProperty = " -Dspecjbb.input.number_products="

	// BinaryDataOutputDir represents directory for storing binary log file of the run.
	BinaryDataOutputDir = " -Dspecjbb.run.datafile.dir="
)

// Config is a config for a SPECjbb2015 Load Generator.,
// Supported options:
// IP - property "-Dspecjbb.controller.host=" - IP address of a SPECjbb controller component (default:127.0.0.1)
// PathToBinary - path to specjbb2015.jar
// PathToProps - path to property file that stores basic configuration.
// TxICount - number of Transaction Injectors in a group.
// CustomerNumber - number of customers used to generate load.
// ProductNuber - number of products used to generate load.
// BinaryDataOutputDir - dir where binary raw data file is stored during run of SPECjbb.
// PathToOutputTemplate - path to template used to generate report from.
type Config struct {
	ControllerIP         string
	PathToBinary         string
	PathToProps          string
	TxICount             int
	CustomerNumber       int
	ProductNumber        int
	BinaryDataOutputDir  string
	PathToOutputTemplate string
}

// NewDefaultConfig is a constructor for Config with default parameters.
func NewDefaultConfig() Config {
	return Config{
		ControllerIP:         IPFlag.Value(),
		PathToBinary:         PathToBinaryFlag.Value(),
		PathToProps:          PathToPropsFileFlag.Value(),
		TxICount:             TxICountFlag.Value(),
		CustomerNumber:       CustomerNumberFlag.Value(),
		ProductNumber:        ProductNumberFlag.Value(),
		BinaryDataOutputDir:  path.Join(fs.GetSwanWorkloadsPath(), "web_serving", "specjbb"),
		PathToOutputTemplate: PathToOutputTemplateFlag.Value(),
	}
}

type reporter struct {
	executor executor.Executor
	config   Config
}

func newReporter(executor executor.Executor, config Config) reporter {
	return reporter{
		executor: executor,
		config:   config,
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

// stop stops all tasks that run as transaction injectors.
func stop(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.Stop()
		if err != nil {
			logrus.Errorf("Can't stop transaction injectors because of an error:%s", err.Error())
		}
	}
}

// clean closes the transaction injectors tasks stdout & stderr files.
func clean(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.Clean()
		if err != nil {
			logrus.Errorf("Can't clean transaction injectors stdout/stderr files because of an error:%s", err.Error())
		}
	}
}

// erase removes transaction injectors tasks stdout & stderr files.
func erase(transactionInjectorsHandles []executor.TaskHandle) {
	for _, handle := range transactionInjectorsHandles {
		err := handle.EraseOutput()
		if err != nil {
			logrus.Errorf("Can't erase transaction injectors stdout/stderr files because of an error:%s", err.Error())
		}
	}
}

func clear(transactionInjectorsHandles []executor.TaskHandle) {
	stop(transactionInjectorsHandles)
	erase(transactionInjectorsHandles)
	clean(transactionInjectorsHandles)
}

func (loadGenerator loadGenerator) runTransactionInjectors() ([]executor.TaskHandle, error) {
	handles := []executor.TaskHandle{}
	for id, txI := range loadGenerator.transactionInjectors {
		command := getTxICommand(loadGenerator.config, id+1)
		handle, err := txI.Execute(command)
		if err != nil {
			logrus.Errorf("Could not start TransactionInjector with command %s", command)
			stop(handles)
			clean(handles)
			erase(handles)
			return nil, errors.Wrapf(err, "Could not start TransactionInjector with command %s", command)
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
// By using RT curve it generates report in which reporter
// calculates Geo-mean of (critical-jOPS@ 10ms, 25ms, 50ms, 75ms and 100ms response time SLAs).
// We use critical jops value because maximum capacity (HBIR) is high above our desired SLA.
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
		return 0, 0, fmt.Errorf("execution of SPECjbb HBIR RT transaction injectors failed with error: %s", err.Error())
	}

	controllerHandle.Wait(0)
	clear(txIHandles)

	outController, err := controllerHandle.StdoutFile()
	if err != nil {
		return 0, 0, errors.Wrapf(err, "could not read controller output file %s", outController.Name())
	}
	rawFileName, err := parser.FileWithRawFileName(outController.Name())
	if err != nil {
		return 0, 0, errors.Wrapf(err, "could not get binary file name from controller output file %s", outController.Name())
	}

	if rawFileName == "" {
		return 0, 0, fmt.Errorf("Could not get raw results file name from an output file %s", outController.Name())
	}

	// Run reporter to calculate critical jops value from raw results.
	reporterCommand := getReporterCommand(loadGenerator.config, rawFileName)
	reporter := newReporter(executor.NewLocal(), loadGenerator.config)
	reporterHandle, err := reporter.executor.Execute(reporterCommand)
	reporterHandle.Wait(0)

	outReporter, err := reporterHandle.StdoutFile()
	if err != nil {
		return 0, 0, errors.Wrapf(err, "could not read reporter output file %s", outReporter.Name())
	}
	hbirRt, err := parser.FileWithHBIRRT(outReporter.Name())
	if err != nil {
		return 0, 0, errors.Wrapf(err, "could not get critical jops from reporter output file %s", outReporter.Name())
	}

	controllerHandle.EraseOutput()
	controllerHandle.Clean()
	reporterHandle.EraseOutput()
	reporterHandle.Clean()

	return hbirRt, 0, err
}

// Load starts a load on the specific workload with the defined loadPoint (injection rate value).
// The task will do the load for specified amount of time (in milliseconds).
func (loadGenerator loadGenerator) Load(injectionRate int, duration time.Duration) (executor.TaskHandle, error) {
	loadCommand := getControllerLoadCommand(loadGenerator.config, injectionRate, duration)
	controllerHandle, err := loadGenerator.controller.Execute(loadCommand)
	if err != nil {
		return nil, errors.Wrapf(err, "execution of SPECjbb Load Generator failed. command: %q", loadCommand)
	}
	txIHandles, err := loadGenerator.runTransactionInjectors()
	if err != nil {
		return nil, fmt.Errorf("execution of SPECjbb HBIR RT transaction injectors failed with error: %s", err.Error())
	}

	return executor.NewClusterTaskHandle(controllerHandle, txIHandles), nil
}
