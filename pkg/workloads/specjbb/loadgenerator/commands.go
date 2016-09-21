package loadgenerator

import (
	"fmt"
)

func getControllerLoadCommand(config Config, injectionRate int, duration int64) string {
	return fmt.Sprint("java -jar",
		ControllerTypeProperty, "PRESET",
		InjectionRateProperty, injectionRate,
		PresetDurationProperty, duration,
		CustomerNumberProperty, config.CustomerNumber,
		ProductNumberProperty, config.ProductNumber,
		BinaryDataOutputDir, config.BinaryDataOutputDir,
		" ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
}

func getControllerHBIRRTCommand(config Config) string {
	return fmt.Sprint("java -jar",
		ControllerTypeProperty, "HBIR_RT",
		CustomerNumberProperty, config.CustomerNumber,
		ProductNumberProperty, config.ProductNumber,
		BinaryDataOutputDir, config.BinaryDataOutputDir,
		" ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
}

func getReporterCommand(config Config, rawFileName string) string {
	return fmt.Sprint("java -jar",
		" ", config.PathToBinary,
		" -m reporter",
		" -s ", config.BinaryDataOutputDir, "/", rawFileName)
}

func getTxICommand(config Config, TxIJVMID int) string {
	return fmt.Sprint("java -jar",
		ControllerHostProperty, config.ControllerIP,
		" ", config.PathToBinary,
		" -m txinjector",
		" -G GRP1",
		" -J JVM", TxIJVMID,
		" -p ", config.PathToProps)
}
