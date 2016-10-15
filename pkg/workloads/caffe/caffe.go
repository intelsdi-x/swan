package caffe

import (
	"fmt"
	"os"
	"path"

	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/utils/env"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID           = "caffe"
	binaryEnvKey = "SWAN_CAFFE_BINARY_PATH"

	defaultBinaryRelativePath = "deep_learning/caffe/train_quick_cifar10.sh"
)

func getPathFromEnvOrDefault(envkey string, relativePath string) string {
	return env.GetOrDefault(
		envkey, path.Join(fs.GetSwanWorkloadsPath(), relativePath))
}

// Config is a config for the Caffe.
type Config struct {
	BinaryPath string
}

// DefaultConfig is a constructor for caffe.Config with default parameters.
func DefaultConfig() Config {
	return Config{
		// TODO(bp): Make that consistent with other workloads.
		BinaryPath: getPathFromEnvOrDefault(binaryEnvKey, defaultBinaryRelativePath),
	}
}

// Caffe is a deep learning framework.
// Implements workload.Launcher.
type Caffe struct {
	exec executor.Executor
	conf Config
}

// New is a constructor for Caffe.
func New(exec executor.Executor, config Config) executor.Launcher {
	return Caffe{
		exec: exec,
		conf: config,
	}

}

func (c Caffe) buildCommand() string {
	return fmt.Sprintf("%s", c.conf.BinaryPath)
}

// Launch launches Caffe workload. It's implementation of workload.Launcher interface.
// Caffe needs to run from it's own working directory, because
// solver look for relative paths when dealing with training and testing
// sets.
func (c Caffe) Launch() (task executor.TaskHandle, err error) {
	currentWorkingDir, err := os.Getwd()
	if err != nil {
		return nil, errors.Wrap(err, "could not obtain working directory")
	}
	defer popWorkingDir(currentWorkingDir)

	task, err = c.exec.Execute(c.buildCommand())
	if err != nil {
		return nil, err
	}

	return task, err
}

func popWorkingDir(workdir string) {
	os.Chdir(workdir)
}

// Name returns human readable name for job.
func (c Caffe) Name() string {
	return "Caffe"
}
