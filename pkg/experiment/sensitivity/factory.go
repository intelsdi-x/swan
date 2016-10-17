package sensitivity

import (
	"github.com/intelsdi-x/athena/pkg/conf"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	"github.com/pkg/errors"
)

const (
	l1dDefaultProcessNumber   = 1
	l1iDefaultProcessNumber   = 1
	l3DefaultProcessNumber    = 1
	membwDefaultProcessNumber = 1
)

// RunCaffeWithLLCIsolationFlag decides which isolations should be used for Caffe aggressor.
var RunCaffeWithLLCIsolationFlag = conf.NewBoolFlag(
	"run_caffe_with_llcisolation",
	"If set, for Caffe workload use the same isolations like use for LLC aggressors, otherwise apply no isolation at all",
	true,
)

// L1dProcessNumber represents number of L1 data cache aggressor processes to be run
var L1dProcessNumber = conf.NewIntFlag(
	"l1d_process_number",
	"Number of L1 data cache aggressors to be run",
	l1dDefaultProcessNumber,
)

// L1iProcessNumber represents number of L1 instruction cache aggressor processes to be run
var L1iProcessNumber = conf.NewIntFlag(
	"l1i_process_number",
	"Number of L1 instruction cache aggressors to be run",
	l1iDefaultProcessNumber,
)

// L3ProcessNumber represents number of L3 data cache aggressor processes to be run
var L3ProcessNumber = conf.NewIntFlag(
	"l3_process_number",
	"Number of L3 data cache aggressors to be run",
	l3DefaultProcessNumber,
)

// MembwProcessNumber represents number of membw aggressor processes to be run
var MembwProcessNumber = conf.NewIntFlag(
	"membw_process_number",
	"Number of membw aggressors to be run",
	membwDefaultProcessNumber,
)

// AggressorFactory is factory for creating aggressor launchers with local executor
// and isolation.
// Isolation depends on type of aggressor that is given to factory.
type AggressorFactory struct {
	l1AggressorIsolation    isolation.Decorator
	otherAggressorIsolation isolation.Decorator
}

// NewSingleIsolationAggressorFactory returns aggressor launchers factory with local executor.
// Aggressors will have the same isolation.
func NewSingleIsolationAggressorFactory(isolation isolation.Decorator) AggressorFactory {
	return AggressorFactory{
		l1AggressorIsolation:    isolation,
		otherAggressorIsolation: isolation,
	}
}

// NewMultiIsolationAggressorFactory returns factory for aggressor launchers with local executor
// and prepared isolations.
// L1-Data and L1-Instruction cache aggressor will receive l1AggressorIsolation
// Other aggressors will receive otherAggressorIsolation
func NewMultiIsolationAggressorFactory(
	l1AggressorIsolation isolation.Decorator,
	otherAggressorIsolation isolation.Decorator) AggressorFactory {
	return AggressorFactory{
		l1AggressorIsolation:    l1AggressorIsolation,
		otherAggressorIsolation: otherAggressorIsolation,
	}
}

// ExecutorFactoryFunc is a type of function that every call returns new instance of executor with decorators.
type ExecutorFactoryFunc func(isolation.Decorators) (executor.Executor, error)

// Create returns aggressor of chosen type using provided executorFactor to create executor.
func (f AggressorFactory) Create(name string, executorFactory ExecutorFactoryFunc) (executor.Launcher, error) {
	var aggressor executor.Launcher

	decorators := f.getDecorators(name)

	exec, err := executorFactory(decorators)
	errutil.Check(err)

	switch name {
	case l1data.ID:
		aggressor = l1data.New(exec, l1data.DefaultL1dConfig())
	case l1instruction.ID:
		aggressor = l1instruction.New(exec, l1instruction.DefaultL1iConfig())
	case memoryBandwidth.ID:
		aggressor = memoryBandwidth.New(exec, memoryBandwidth.DefaultMemBwConfig())
	case caffe.ID:
		aggressor = caffe.New(exec, caffe.DefaultConfig())
	case l3data.ID:
		aggressor = l3data.New(exec, l3data.DefaultL3Config())
	case stream.ID:
		aggressor = stream.New(exec, stream.DefaultConfig())
	default:
		return nil, errors.Errorf("aggressor %q not found", name)
	}

	return aggressor, nil
}

// getDecorators returns decorators that should be applied on executor.
// L1-Data and L1-Instruction cache aggressor receives l1AggressorIsolation
// Other aggressors receive otherAggressorIsolation
func (f AggressorFactory) getDecorators(name string) isolation.Decorators {

	switch name {
	case l1data.ID:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		if L1dProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L1dProcessNumber.Value()))
		}
		return decorators
	case l1instruction.ID:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		if L1iProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L1iProcessNumber.Value()))
		}
		return decorators
	case l3data.ID:
		decorators := isolation.Decorators{f.otherAggressorIsolation}
		if L3ProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L3ProcessNumber.Value()))
		}
		return decorators
	case memoryBandwidth.ID:
		decorators := isolation.Decorators{f.otherAggressorIsolation}
		if MembwProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(MembwProcessNumber.Value()))
		}
		return decorators
	case caffe.ID:
		if RunCaffeWithLLCIsolationFlag.Value() {
			return isolation.Decorators{f.otherAggressorIsolation}
		}
		return isolation.Decorators{}
	default:
		return isolation.Decorators{f.otherAggressorIsolation}
	}
}
