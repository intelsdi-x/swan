package specjbb

import (
	"time"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/parser"
	"github.com/pkg/errors"
)

const (
	defaultControllerIP   = "127.0.0.1"
	defaultCustomerNumber = 100000 // Default as in SPECjbb.
	defaultProductsNumber = 100000 // Default as in SPECjbb.
)

var (
	// PathToBinaryForLoadGeneratorFlag specifies path to a SPECjbb2015 jar file for load generator.
	PathToBinaryForLoadGeneratorFlag = conf.NewStringFlag("specjbb_path_lg", "Path to SPECjbb jar for load generator", "/usr/share/specjbb/specjbb2015.jar")

	// PathToPropsFileForLoadGeneratorFlag specifies path to a SPECjbb2015 properties file for load generator.
	PathToPropsFileForLoadGeneratorFlag = conf.NewStringFlag("specjbb_props_path_lg", "Path to SPECjbb properties file for load generator", "/usr/share/specjbb/config/specjbb2015.props")

	// PathToOutputTemplateFlag specifies path to a SPECjbb2015 output template file.
	PathToOutputTemplateFlag = conf.NewStringFlag("specjbb_output_template_path", "Path to SPECjbb output template file", "/usr/share/specjbb/config/template-D.raw")

	// ControllerAddress specifies ControllerAddress address of a controller component of SPECjbb2015 benchmark.
	ControllerAddress = conf.NewStringFlag("specjbb_controller_ip", "ControllerAddress address of a SPECjbb controller component", defaultControllerIP)

	// CustomerNumberFlag specifies number of customers.
	CustomerNumberFlag = conf.NewIntFlag("specjbb_customer_number", "Number of customers", defaultCustomerNumber)

	// ProductNumberFlag specifies number of products.
	ProductNumberFlag = conf.NewIntFlag("specjbb_product_number", "Number of products", defaultProductsNumber)

	// BinaryDataOutputDirFlag specifies output dir for storing binary data.
	BinaryDataOutputDirFlag = conf.NewStringFlag("specjbb_output_dir", "Path to location of storing binary data", "/usr/share/specjbb/")
)

// LoadGeneratorConfig is a config for a SPECjbb2015 Load Generator.,
type LoadGeneratorConfig struct {
	JVMOptions
	ControllerAddress    string // ControllerAddress is an address of a SPECjbb controller component ("-Dspecjbb.controller.host=").
	PathToBinary         string // PathToBinary is a path to specjbb2015.jar.
	PathToProps          string // PathToProps is a path to property file that stores basic configuration.
	CustomerNumber       int    // CustomerNumber is a number of customers used to generate load.
	ProductNumber        int    // ProductNumber is a number of products used to generate load.
	BinaryDataOutputDir  string // BinaryDataOutputDir is a dir where binary raw data file is stored during run of SPECjbb.
	PathToOutputTemplate string // PathToOutputTemplate is a path to template used to generate report from.
	HandshakeTimeoutMs   int    // HandshakeTimeoutMs is timeout (in milliseconds) for initial Controller <-> Agent handshaking.
}

// DefaultLoadGeneratorConfig is a constructor for LoadGeneratorConfig with default parameters.
func DefaultLoadGeneratorConfig() LoadGeneratorConfig {
	return LoadGeneratorConfig{
		JVMOptions:           DefaultJVMOptions(),
		ControllerAddress:    ControllerAddress.Value(),
		PathToBinary:         PathToBinaryForLoadGeneratorFlag.Value(),
		PathToProps:          PathToPropsFileForLoadGeneratorFlag.Value(),
		CustomerNumber:       CustomerNumberFlag.Value(),
		ProductNumber:        ProductNumberFlag.Value(),
		BinaryDataOutputDir:  BinaryDataOutputDirFlag.Value(),
		PathToOutputTemplate: PathToOutputTemplateFlag.Value(),
		HandshakeTimeoutMs:   600000,
	}
}

type reporter struct {
	executor executor.Executor
	config   LoadGeneratorConfig
}

func newReporter(executor executor.Executor, config LoadGeneratorConfig) reporter {
	return reporter{
		executor: executor,
		config:   config,
	}
}

type loadGenerator struct {
	controller           executor.Executor
	transactionInjectors []executor.Executor
	config               LoadGeneratorConfig
}

// NewLoadGenerator returns a new SPECjbb Load Generator instance composed of controller
// and transaction injectors.
// Transaction Injector and Controller are load generator for SPECjbb Backend.
// https://www.spec.org/jbb2015/docs/userguide.pdf
func NewLoadGenerator(controller executor.Executor, transactionInjectors []executor.Executor, config LoadGeneratorConfig) executor.LoadGenerator {
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

// Populate. SpecJBB backend is ready to use immediately after run and it does not need any population.
func (loadGenerator loadGenerator) Populate() (err error) {
	return nil
}

// Tune calculates maximum number of "critical java operations" under SLO
// @param slo: SLO in us (sane values are above 5000us [5ms])
//
// It generates High Bound Injection Rate [HBIR] curve to determine the load under slo value.
// See SPECjbb readme (https://www.spec.org/jbb2015/docs/userguide.pdf) for details.
//
// Exemplary output for machine capacity, HBIR = 12000:
// RUN RESULT: hbIR (max attempted) = 12000, hbIR (settled) = 12000, max-jOPS = 11640, critical-jOPS = 2684
func (loadGenerator loadGenerator) Tune(slo int) (qps int, achievedSLI int, err error) {
	if slo < 5000 {
		logrus.Errorf("SLO for tuning SPECjbb should be above 5000us (5ms). Function received %d", slo)
		return 0, 0, errors.Errorf("SLO for tuning SPECjbb should be above 5000us (5ms). Function received %d", slo)
	}

	hbirRtCommand := getControllerTuneCommand(loadGenerator.config)
	controllerHandle, err := loadGenerator.controller.Execute(hbirRtCommand)
	if err != nil {
		return 0, 0, errors.Wrapf(err, "execution of SPECjbb HBIR RT failed. command: %q", hbirRtCommand)
	}
	txIHandles, err := loadGenerator.runTransactionInjectors()
	if err != nil {
		return 0, 0, errors.Errorf("execution of SPECjbb HBIR RT transaction injectors failed with error: %s", err.Error())
	}

	controllerHandle.Wait(0)
	clear(txIHandles)

	controllerStdOut, err := controllerHandle.StdoutFile()
	if err != nil {
		return 0, 0, errors.Wrapf(err, "could not read controller output file %s", controllerStdOut.Name())
	}
	rawFileName, err := parser.FileWithRawFileName(controllerStdOut.Name())
	if err != nil {
		return 0, 0, errors.Wrapf(err, "could not get binary file name from controller output file %s", controllerStdOut.Name())
	}

	if rawFileName == "" {
		return 0, 0, errors.Errorf("Could not get raw results file name from an output file %s", controllerStdOut.Name())
	}

	// Run reporter to calculate critical jops value from raw results.
	reporterCommand := getReporterCommand(loadGenerator.config, rawFileName, slo)
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

// Load starts SPECjbb load on backed with given injection rate value.
// The task will do the load for specified amount of time.
func (loadGenerator loadGenerator) Load(injectionRate int, duration time.Duration) (executor.TaskHandle, error) {
	loadCommand := getControllerLoadCommand(loadGenerator.config, injectionRate, duration)
	controllerHandle, err := loadGenerator.controller.Execute(loadCommand)
	if err != nil {
		return nil, errors.Wrapf(err, "execution of SPECjbb Load Generator failed. command: %q", loadCommand)
	}
	txIHandles, err := loadGenerator.runTransactionInjectors()
	if err != nil {
		return nil, errors.Errorf("execution of SPECjbb HBIR RT transaction injectors failed with error: %s", err.Error())
	}

	return executor.NewClusterTaskHandle(controllerHandle, txIHandles), nil
}
