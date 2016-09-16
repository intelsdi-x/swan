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
		" ", config.PathToBinary,
		" -m distcontroller",
		" -p ", config.PathToProps)
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
