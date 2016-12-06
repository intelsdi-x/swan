package specjbb

import (
	"fmt"
	"time"
)

// Load command performs load of given injection rate for given duration.
func getControllerLoadCommand(config LoadGeneratorConfig, injectionRate int, duration time.Duration) string {
	return fmt.Sprint("java -jar",
		" -Dcom.sun.management.jmxremote.port=5556",
		" -Dcom.sun.management.jmxremote.ssl=false",
		" -Dcom.sun.management.jmxremote.authenticate=false",
		" -Djava.net.preferIPv4Stack=true",
		ControllerTypeProperty, "PRESET",
		InjectionRateProperty, injectionRate,
		PresetDurationProperty, int(duration.Seconds())*1000, // [milliseconds] SPECjbb expects duration in milliseconds.
		CustomerNumberProperty, config.CustomerNumber,
		ProductNumberProperty, config.ProductNumber,
		BinaryDataOutputDir, config.BinaryDataOutputDir,
		" ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
}

// HBIR RT command looks for maximum capacity of a machine (high bound injection rate).
// Then it performs load from 1% to 100% of calculated HBIR (step 1%).
func getControllerHBIRRTCommand(config LoadGeneratorConfig) string {
	return fmt.Sprint("java -jar",
		" -Dcom.sun.management.jmxremote.port=5557",
		" -Dcom.sun.management.jmxremote.ssl=false",
		" -Dcom.sun.management.jmxremote.authenticate=false",
		" -Djava.net.preferIPv4Stack=true",
		ControllerTypeProperty, "HBIR_RT",
		CustomerNumberProperty, config.CustomerNumber,
		ProductNumberProperty, config.ProductNumber,
		BinaryDataOutputDir, config.BinaryDataOutputDir,
		" ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
}

// Reporter command allows to generate report from raw file (binary data).
func getReporterCommand(config LoadGeneratorConfig, rawFileName string, slo int) string {
	return fmt.Sprint("java -jar",
		" -Dcom.sun.management.jmxremote.port=5557",
		" -Dcom.sun.management.jmxremote.ssl=false",
		" -Dcom.sun.management.jmxremote.authenticate=false",
		" -Djava.net.preferIPv4Stack=true",
		" ", config.PathToBinary,
		" -m reporter",
		" -cIRtarget ", slo,
		" -p ", config.PathToProps,
		" -raw ", config.PathToOutputTemplate,
		" -s ", rawFileName)
}

// TxI command starts transaction injector.
func getTxICommand(config LoadGeneratorConfig, TxIJVMID int) string {
	return fmt.Sprint("java -jar",
		" -Dcom.sun.management.jmxremote.port=5558",
		" -Dcom.sun.management.jmxremote.ssl=false",
		" -Dcom.sun.management.jmxremote.authenticate=false",
		" -Djava.net.preferIPv4Stack=true",
		ControllerHostProperty, config.ControllerIP,
		" ", config.PathToBinary,
		" -m txinjector",
		" -G GRP1",
		" -J JVM", TxIJVMID,
		" -p ", config.PathToProps)
}
