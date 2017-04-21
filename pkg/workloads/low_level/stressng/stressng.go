package stressng

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
)

const (
	// IDCustom is used for specifying which aggressors should be used via parameters.
	IDCustom = "stress-ng-custom"
	// IDStream is used for specifying which aggressors should be used via parameters.
	IDStream = "stress-ng-stream"
	// IDCacheL1 is used for specifying which aggressors should be used via parameters.
	IDCacheL1 = "stress-ng-cache-l1"
	// IDCacheL3 is used for specifying which aggressors should be used via parameters.
	IDCacheL3 = "stress-ng-cache-l3"
	// IDMemCpy is used for specifying which aggressors should be used via parameters.
	IDMemCpy = "stress-ng-memcpy"
)

// StressngCustomArguments custom argument to run stress-ng with.
var StressngCustomArguments = conf.NewStringFlag("stressng_custom_arguments", "Custom arguments to stress-ng", "")

// StressngStreamProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngStreamProcessNumber = conf.NewIntFlag("stressng_stream_process_number", "Number of aggressors to be run", 1)

// StressngCacheL1ProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngCacheL1ProcessNumber = conf.NewIntFlag("stressng_cache_l1_process_number", "Number of aggressors to be run", 1)

// StressngCacheL3ProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngCacheL3ProcessNumber = conf.NewIntFlag("stressng_cache_l3_process_number", "Number of aggressors to be run", 1)

// StressngMemCpyProcessNumber represents number of stress-ng aggressor processes to be run.
var StressngMemCpyProcessNumber = conf.NewIntFlag("stressng_memcpy_process_number", "Number of aggressors to be run", 1)

// l1d is a launcher for stress-ng aggressor.
type stressng struct {
	executor  executor.Executor
	arguments string
	name      string
}

// New is a constructor for stress-ng aggressor.
func New(executor executor.Executor, name, arguments string) executor.Launcher {
	return stressng{
		executor:  executor,
		arguments: arguments,
		name:      name,
	}
}

// NewCustom constructor for specifc run of stress-ng.
func NewCustom(executor executor.Executor) executor.Launcher {
	return New(executor, fmt.Sprintf("stress-ng-custom %s", StressngCustomArguments.Value()), StressngCustomArguments.Value())
}

// NewStream constructor for stream based run of stress-ng.
func NewStream(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-stream", fmt.Sprintf("--stream=%d", StressngStreamProcessNumber.Value()))
}

// NewCacheL1 constructor for cache L1 run of stress-ng.
func NewCacheL1(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-cache-l1", fmt.Sprintf("--cache=%d --cache-level=1", StressngCacheL1ProcessNumber.Value()))
}

// NewCacheL3 constructor for cache L3 run of stress-ng.
func NewCacheL3(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-cache-l3", fmt.Sprintf("--cache=%d --cache-level=3", StressngCacheL3ProcessNumber.Value()))
}

// NewMemCpy constructor for memcpy stressor run of stress-ng.
func NewMemCpy(executor executor.Executor) executor.Launcher {
	return New(executor, "stress-ng-memcpy", fmt.Sprintf("--memcpy=%d", StressngMemCpyProcessNumber.Value()))
}

// Launch starts a workload.
func (s stressng) Launch() (executor.TaskHandle, error) {
	return s.executor.Execute(fmt.Sprintf("stress-ng %s", s.arguments))
}

// Name returns readable name.
func (s stressng) Name() string {
	return s.name
}
