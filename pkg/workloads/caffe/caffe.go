package caffe

import (
	"fmt"
	"os"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID = "caffe"

	defaultBinaryRelativePath = "caffe"
	defualCaffeLibPath        = "/opt/swan/lib"
	defaultModel              = "/opt/swan/share/caffe/examples/cifar10/cifar10_quick_train_test.prototxt"
	defaultWeights            = "/tmp/caffe/cifar10_quick_iter_5000.caffemodel.h5"
	defaultIterations         = 1000000000
	defaultSigintEffect       = "stop"
)

var caffePath = conf.NewStringFlag(
	"caffe_path",
	"Path to script launching caffe as an aggressor", defaultBinaryRelativePath,
)

var caffeLibPath = conf.NewStringFlag(
	"caffe_lib_path",
	"Path to caffe libraries",
	defualCaffeLibPath,
)

var caffeModel = conf.NewStringFlag(
	"caffe_model",
	"Path to trained model",
	defaultModel,
)

var caffeWeights = conf.NewStringFlag(
	"caffe_weights",
	"Path to trained weight",
	defaultWeights,
)

var caffeIterations = conf.NewIntFlag(
	"caffe_iterations",
	"Number of iterations",
	defaultIterations,
)

var caffeSigintEffect = conf.NewStringFlag(
	"caffe_sigint",
	"Sigint effect for caffe",
	defaultSigintEffect,
)

// Config is a config for the Caffe.
type Config struct {
	BinaryPath       string
	LibPath          string
	ModelPath        string
	WeightsPath      string
	IterationsNumber int
	SigintEffect     string
}

// DefaultConfig is a constructor for caffe.Config with default parameters.
func DefaultConfig() Config {
	return Config{
		BinaryPath:       caffePath.Value(),
		LibPath:          caffeLibPath.Value(),
		ModelPath:        caffeModel.Value(),
		WeightsPath:      caffeWeights.Value(),
		IterationsNumber: caffeIterations.Value(),
		SigintEffect:     caffeSigintEffect.Value(),
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
	return fmt.Sprintf("LD_LIBRARY_PATH=%q %s test -model %s -weights %s -iterations %d -sigint_effect %s",
		c.conf.LibPath,
		c.conf.BinaryPath,
		c.conf.ModelPath,
		c.conf.WeightsPath,
		c.conf.IterationsNumber,
		c.conf.SigintEffect)
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
