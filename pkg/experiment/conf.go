package experiment

import (
	"fmt"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

var (
	// DumpConfigFlag name includes dash to excluded it from dumping.
	dumpConfigFlag = conf.NewBoolFlag("config-dump", "Dump configuration as environment script.", false)

	// DumpConfigExperimentIDFlag name includes dash to excluded it from dumping.
	dumpConfigExperimentIDFlag = conf.NewStringFlag("config-dump-experiment-id", "Dump configuration based on experiment ID.", "")
)

// Configure handles configuration parsing, generation and restoration based on config-* flags.
// Note: exits if configuration generation was requested.
// This function must reside in experiment package because depends on metadata access.
func Configure() {

	err := conf.ParseFlags()
	if err != nil {
		logrus.Errorf("Cannot parse flags: %q", err.Error())
		os.Exit(ExUsage)
	}
	logrus.SetLevel(conf.LogLevel())

	if dumpConfigFlag.Value() {
		previousExperimentID := dumpConfigExperimentIDFlag.Value()
		if previousExperimentID != "" {
			metadata := NewMetadata(previousExperimentID, MetadataConfigFromFlags())
			err := metadata.Connect()
			errutil.Check(err)
			flags, err := metadata.GetGroup("flags")
			errutil.Check(err)
			fmt.Println(conf.DumpConfigMap(flags))
		} else {
			fmt.Println(conf.DumpConfig())
		}
		os.Exit(0)
	}
}
