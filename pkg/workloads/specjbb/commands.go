package specjbb

import (
	"fmt"
	"time"
)

func getBackendCommand(conf BackendConfig) string {
	// See: https://intelsdi.atlassian.net/wiki/display/SCE/SpecJBB+experiment+tuning
	return fmt.Sprint("java -jar",
		" -server", // Compilation takes more time but offers additional optimizations

		fmt.Sprintf(" -Djava.util.concurrent.ForkJoinPool.common.parallelism=%d", conf.Parallelism), // Amount of threads equal to amount of hyper threads

		fmt.Sprintf(" -Xms%dg -Xmx%dg", conf.JVMHeapMemoryGBs, conf.JVMHeapMemoryGBs), // Allocate whole heap available; docs: For best performance, set -Xms to the same size as the maximum heap size
		" -XX:NativeMemoryTracking=summary",                                           // Memory monitoring purposes
		" -XX:+UseParallelGC",                                                         // Parallel garbage collector
		fmt.Sprintf(" -XX:ParallelGCThreads=%d", conf.Parallelism),                    // Sets the value of n to the number of logical processors. The value of n is the same as the number of logical processors up to a value of 8.
		fmt.Sprintf(" -XX:ConcGCThreads=%d", conf.Parallelism/2),                      // Currently half of PGCThreads.
		" -XX:InitiatingHeapOccupancyPercent=80",                                      // Using more memory then default 45% before GC kicks in
		" -XX:MaxGCPauseMillis=100",                                                   //Sets a target value for desired maximum pause time. The default value is 200 milliseconds. The specified value does not adapt to your heap size.

		ControllerHostProperty, conf.ControllerAddress,
		" ", conf.PathToBinary,
		" -m backend",
		" -G GRP1",
		" -J ", conf.JvmID,
		" -p ", PathToPropsFileForHpFlag.Value(),
	)
}

// Load command performs load of given injection rate for given duration.
func getControllerLoadCommand(config LoadGeneratorConfig, injectionRate int, duration time.Duration) string {
	return fmt.Sprint("java -jar",
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
		" ", config.PathToBinary,
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
		ControllerHostProperty, config.ControllerAddress,
		" ", config.PathToBinary,
		" -m txinjector",
		" -G GRP1",
		" -J JVM", TxIJVMID,
		" -p ", config.PathToProps)
}
