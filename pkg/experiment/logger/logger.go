package logger

import (
	"fmt"
	"io"
	"os"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
)

//Initialize creates experiment logs directory and configures logrus for an experiment.
func Initialize(appName, uuid string) {
	// Create experiment directory
	experimentDirectory, logFile, err := experiment.CreateExperimentDir(uuid, appName)
	errutil.CheckWithContext(err, "Cannot create experiment logs directory")

	// Setup logging set to both output and logFile.
	logrus.SetFormatter(&logrus.TextFormatter{FullTimestamp: true, TimestampFormat: "2006-01-02 15:04:05.100"})
	logrus.Infof("Working directory %q", experimentDirectory)
	logrus.SetOutput(io.MultiWriter(logFile, os.Stderr))

	// Logging and outputting experiment ID.
	logrus.Info("Starting Experiment ", appName, " with uid ", uuid)
	fmt.Println(uuid)
}
