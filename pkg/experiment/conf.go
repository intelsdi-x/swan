package experiment

import (
	"fmt"
	"os"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

var (
	// DumpConfigFlag name includes dash to excluded it from dumping.
	dumpConfigFlag = conf.NewBoolFlag("config-dump", "Dump configuration as environment script.", false)

	// DumpConfigExperimentIDFlag name includes dash to excluded it from dumping.
	dumpConfigExperimentIDFlag = conf.NewStringFlag("config-dump-experiment-id", "Dump configuration based on experiment ID.", "")
)

// ManageConfiguration handles configuration script generation and restoration based on config-* flags.
// Note: exits if configuration dump was requested.
func ManageConfiguration() {

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
