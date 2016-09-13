package loadgenerator

import (
	"fmt"
	"time"
)

func getControllerTuneCommand(pathToBinary string) string {
	return fmt.Sprint("java -jar",
		ControllerTypeProperty, "HBIR",
		" ", pathToBinary,
		" -m distcontroller")
}

func getControllerLoadCommand(pathToBinary string, injectionRate int, duration time.Duration) string {
	return fmt.Sprint("java -jar",
		ControllerTypeProperty, "PRESET",
		InjectionRateProperty, injectionRate,
		PresetDurationProperty, duration,
		" ", pathToBinary,
		" -m distcontroller")
}

func getTxICommand(pathToBinary string, txIConfig Config) string {
	return fmt.Sprint("java -jar",
		ControllerHostProperty, txIConfig.ControllerIP,
		" ", pathToBinary,
		" -m txinjector",
		" -G GRP1",
		" -J JVM", txIConfig.JvmID)
}
