package specjbb

import (
	"fmt"
	"time"
)

var (
	// controllerHostProperty - string name for property that specifies controller host.
	controllerHostProperty = " -Dspecjbb.controller.host="

	// controllerTypeProperty - string name for property that specifies controller type.
	controllerTypeProperty = " -Dspecjbb.controller.type="

	// injectionRateProperty - string name for property that specifies ir.
	injectionRateProperty = " -Dspecjbb.controller.preset.ir="

	// presetDurationProperty - string name for property that specifies preset duration.
	presetDurationProperty = " -Dspecjbb.controller.preset.duration="

	// customerNumberProperty represents total number of customers.
	customerNumberProperty = " -Dspecjbb.input.number_customers="

	// productNumberProperty represents total number of products.
	productNumberProperty = " -Dspecjbb.input.number_products="

	// binaryDataOutputDir represents directory for storing binary log file of the run.
	binaryDataOutputDir = " -Dspecjbb.run.datafile.dir="

	//specjbb.forkjoin.workers
	forkjoinWorkers = " -Dspecjbb.forkjoin.workers="

	// Timeout (in milliseconds) for initial Controller <-> Agent handshaking.
	handshakeTimeoutProperty      = " -Dspecjbb.controller.handshake.timeout="
	handshakeTimeoutPropertyValue = 600000
)

func getBackendCommand(conf BackendConfig) string {
	// See: https://intelsdi.atlassian.net/wiki/display/SCE/SpecJBB+experiment+tuning
	return fmt.Sprint("java",
		conf.GetJVMOptions(),
		controllerHostProperty, conf.ControllerAddress,
		forkjoinWorkers, conf.WorkerCount,
		" -jar ", conf.PathToBinary,
		" -m backend",
		" -G GRP1",
		" -J ", conf.JvmID,
		" -p ", PathToPropsFileForHpFlag.Value(),
	)
}

// Load command performs load of given injection rate for given duration.
func getControllerLoadCommand(config LoadGeneratorConfig, injectionRate int, duration time.Duration) string {
	return fmt.Sprint("java ",
		config.GetJVMOptions(),
		controllerTypeProperty, "PRESET", // PRESET: Takes IR set by specjbb.controller.preset.ir and runs on the IR for specjbb.controller.preset.duration milliseconds
		injectionRateProperty, injectionRate,
		presetDurationProperty, int(duration.Seconds())*1000, // [milliseconds] SPECjbb expects duration in milliseconds.
		controllerSubCommand(config),
	)
}

// HBIR RT command looks for maximum capacity of a machine (high bound injection rate).
// Then it performs load from 1% to 100% of calculated HBIR (step 1%).
func getControllerTuneCommand(config LoadGeneratorConfig) string {
	return fmt.Sprint("java",
		config.GetJVMOptions(),
		controllerTypeProperty, "HBIR_RT",
		controllerSubCommand(config),
	)
}

// Common part of Load & Tune command.
func controllerSubCommand(config LoadGeneratorConfig) string {
	return fmt.Sprint(
		customerNumberProperty, config.CustomerNumber,
		productNumberProperty, config.ProductNumber,
		binaryDataOutputDir, config.BinaryDataOutputDir,
		controllerHostProperty, config.ControllerAddress,
		handshakeTimeoutProperty, handshakeTimeoutPropertyValue,
		" -jar ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps,
	)
}

// Reporter command allows to generate report from raw file (binary data).
func getReporterCommand(config LoadGeneratorConfig, rawFileName string, slo int) string {
	return fmt.Sprint("java",
		config.GetJVMOptions(),
		" -jar ", config.PathToBinary,
		" -m reporter",
		" -cIRtarget ", slo,
		" -p ", config.PathToProps,
		" -raw ", config.PathToOutputTemplate,
		" -s ", rawFileName,
		" -p ", config.PathToProps)
}

// TxI command starts transaction injector.
func getTxICommand(config LoadGeneratorConfig, TxIJVMID int) string {
	return fmt.Sprint("java -jar",
		config.GetJVMOptions(),
		controllerHostProperty, config.ControllerAddress,
		" -jar ", config.PathToBinary,
		" -m txinjector",
		" -G GRP1",
		" -J JVM", TxIJVMID,
		" -p ", config.PathToProps)
}
