// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package caffe

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

const (
	// ID is used for specifying which aggressors should be used via parameters.
	ID = "caffe"

	defaultCaffeWrapper = "caffe.sh"
	defaultModel        = "examples/cifar10/cifar10_quick_train_test.prototxt" // relative to caffe binnary
	defaultWeights      = "examples/cifar10/cifar10_quick_iter_5000.caffemodel.h5"
	defaultIterations   = 1000000000
	defaultSigintEffect = "stop"
)

var (
	caffeModel   = conf.NewStringFlag("caffe_model", "Path to trained model", defaultModel)
	caffeWeights = conf.NewStringFlag("caffe_weights", "Path to trained weights", defaultWeights)
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
		BinaryPath:       defaultCaffeWrapper,
		ModelPath:        caffeModel.Value(),
		WeightsPath:      caffeWeights.Value(),
		IterationsNumber: defaultIterations,
		SigintEffect:     defaultSigintEffect,
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
	return fmt.Sprintf("%s test -model %s -weights %s -iterations %d -sigint_effect %s",
		c.conf.BinaryPath,
		c.conf.ModelPath,
		c.conf.WeightsPath,
		c.conf.IterationsNumber,
		c.conf.SigintEffect)
}

// Launch launches Caffe workload. It's implementation of workload.Launcher interface.
func (c Caffe) Launch() (task executor.TaskHandle, err error) {
	command := c.buildCommand()
	task, err = c.exec.Execute(command)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot launch caffe with command %q", command)
	}
	return
}

// Name returns human readable name for job.
func (c Caffe) Name() string {
	return "Caffe"
}
