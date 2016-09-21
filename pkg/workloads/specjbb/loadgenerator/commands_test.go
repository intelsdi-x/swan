package loadgenerator

import (
	"fmt"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

const (
	transactionInjectorIndex = 1
	injectionRate            = 6000
	duration                 = 100
	customerNumber           = 100
	productsNumber           = 100
	jarPath                  = "/swan/workloads/web_serving/specjbb/specjbb2015.jar"
	propertiesFilePath       = "/swan/workloads/web_serving/specjbb/config/specjbb2015.props"
	outputDir                = "/swan/workloads/web_serving/specjbb"
	rawFileName              = "abc"
)

func TestCommandsWithDefaultConfig(t *testing.T) {

	Convey("While having default config", t, func() {
		defaultConfig := NewDefaultConfig()

		Convey("and SPECjbb transaction injector command", func() {
			command := getTxICommand(defaultConfig, transactionInjectorIndex)
			Convey("Should contain txinjector mode", func() {
				So(command, ShouldContainSubstring, "-m txinjector")
			})
			Convey("Should contain controller IP with host property", func() {
				So(command, ShouldContainSubstring, "-Dspecjbb.controller.host=127.0.0.1")
			})
			Convey("Should contain proper group", func() {
				So(command, ShouldContainSubstring, "-G GRP1")
			})
			Convey("Should contain proper JVM id", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-J JVM%d", transactionInjectorIndex))
			})
			Convey("Should contain path to binary", func() {
				So(command, ShouldContainSubstring, jarPath)
			})
			Convey("Should contain path to properties file", func() {
				So(command, ShouldContainSubstring, propertiesFilePath)
			})
		})

		Convey("and SPECjbb load command", func() {
			command := getControllerLoadCommand(defaultConfig, injectionRate, duration)
			Convey("Should contain controller mode", func() {
				So(command, ShouldContainSubstring, "-m distcontroller")
			})
			Convey("Should injection controller type property", func() {
				So(command, ShouldContainSubstring, "-Dspecjbb.controller.type=PRESET")
			})
			Convey("Should contain injection rate property", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-Dspecjbb.controller.preset.duration=%d", duration))
			})
			Convey("Should contain preset duration property", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-Dspecjbb.controller.preset.ir=%d", injectionRate))
			})
			Convey("Should contain customer number property", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-Dspecjbb.input.number_customers=%d", customerNumber))
			})
			Convey("Should contain product number property", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-Dspecjbb.input.number_products=%d", productsNumber))
			})
			Convey("Should contain path to binary", func() {
				So(command, ShouldContainSubstring, jarPath)
			})
			Convey("Should contain path to properties file", func() {
				So(command, ShouldContainSubstring, propertiesFilePath)
			})
			Convey("Should contain path to output dir", func() {
				So(command, ShouldContainSubstring, outputDir)
			})
		})

		Convey("and SPECjbb HBIR RT command", func() {
			command := getControllerHBIRRTCommand(defaultConfig)
			Convey("Should contain controller mode", func() {
				So(command, ShouldContainSubstring, "-m distcontroller")
			})
			Convey("Should injection controller type property", func() {
				So(command, ShouldContainSubstring, "-Dspecjbb.controller.type=HBIR")
			})
			Convey("Should contain customer number property", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-Dspecjbb.input.number_customers=%d", customerNumber))
			})
			Convey("Should contain product number property", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("-Dspecjbb.input.number_products=%d", productsNumber))
			})
			Convey("Should contain path to binary", func() {
				So(command, ShouldContainSubstring, jarPath)
			})
			Convey("Should contain path to properties file", func() {
				So(command, ShouldContainSubstring, propertiesFilePath)
			})
			Convey("Should contain path to output dir", func() {
				So(command, ShouldContainSubstring, outputDir)
			})
		})

		Convey("and SPECjbb reporter command", func() {
			command := getReporterCommand(defaultConfig, rawFileName)
			Convey("Should contain reporter mode", func() {
				So(command, ShouldContainSubstring, "-m reporter")
			})
			Convey("Should contain path to binary", func() {
				So(command, ShouldContainSubstring, jarPath)
			})
			Convey("Should contain path to output dir", func() {
				So(command, ShouldContainSubstring, fmt.Sprintf("%s/%s", outputDir, rawFileName))
			})
		})
	})
}
