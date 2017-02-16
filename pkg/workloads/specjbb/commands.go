package specjbb

import (
	"fmt"
	"time"
)

func getBackendCommand(conf BackendConfig) string {
	// See: https://intelsdi.atlassian.net/wiki/display/SCE/SpecJBB+experiment+tuning
	return fmt.Sprint("java",
		conf.GetJVMOptions(),
		ControllerHostProperty, conf.ControllerAddress,
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
		ControllerTypeProperty, "PRESET", // PRESET: Takes IR set by specjbb.controller.preset.ir and runs on the IR for specjbb.controller.preset.duration milliseconds
		InjectionRateProperty, injectionRate,
		PresetDurationProperty, int(duration.Seconds())*1000, // [milliseconds] SPECjbb expects duration in milliseconds.
		CustomerNumberProperty, config.CustomerNumber,
		ProductNumberProperty, config.ProductNumber,
		BinaryDataOutputDir, config.BinaryDataOutputDir,
		" -jar ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
}

// HBIR RT command looks for maximum capacity of a machine (high bound injection rate).
// Then it performs load from 1% to 100% of calculated HBIR (step 1%).
func getControllerHBIRRTCommand(config LoadGeneratorConfig) string {
	return fmt.Sprint("java",
		config.GetJVMOptions(),
		ControllerTypeProperty, "HBIR_RT",
		CustomerNumberProperty, config.CustomerNumber,
		ProductNumberProperty, config.ProductNumber,
		BinaryDataOutputDir, config.BinaryDataOutputDir,
		" -jar ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
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
		ControllerHostProperty, config.ControllerAddress,
		" -jar ", config.PathToBinary,
		" -m txinjector",
		" -G GRP1",
		" -J JVM", TxIJVMID,
		" -p ", config.PathToProps)
}
