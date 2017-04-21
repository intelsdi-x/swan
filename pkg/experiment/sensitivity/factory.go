package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/caffe"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
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

var (
	// AggressorsFlag is a comma separated list of aggressors to be run during the experiment.
	AggressorsFlag = conf.NewStringSliceFlag(
		"experiment_aggressor_workloads", "Best Effort workloads that will be run sequentially in co-location with High Priority workload.",
		[]string{"l1d", "l1i", "l3", "stream", "caffe"},
	)

	threatAggressorsAsService = conf.NewBoolFlag(
		"debug_threat_aggressors_as_service", "Debug only: aggressors are wrapped in Service flags so that the experiment can track their lifectcle. Default `true` should not be changed without explicit reason.",
		true)

	// L1dProcessNumber represents number of L1 data cache aggressor processes to be run
	L1dProcessNumber = conf.NewIntFlag(
		"experiment_aggressor_l1d_process_number",
		"Number of L1 data cache aggressor processes to be run",
		l1dDefaultProcessNumber,
	)

	// L1iProcessNumber represents number of L1 instruction cache aggressor processes to be run
	L1iProcessNumber = conf.NewIntFlag(
		"experiment_aggressor_l1i_process_number",
		"Number of L1 instruction cache aggressors processes to be run",
		l1iDefaultProcessNumber,
	)

	// L3ProcessNumber represents number of L3 data cache aggressor processes to be run
	L3ProcessNumber = conf.NewIntFlag(
		"experiment_aggressor_l3_process_number",
		"Number of L3 data cache aggressors processes to be run",
		l3DefaultProcessNumber,
	)

	// MembwProcessNumber represents number of membw aggressor processes to be run
	MembwProcessNumber = conf.NewIntFlag(
		"experiment_aggressor_membw_process_number",
		"Number of membw aggressors processes to be run",
		membwDefaultProcessNumber,
	)

	// RunCaffeWithLLCIsolationFlag decides which isolations should be used for Caffe aggressor.
	RunCaffeWithLLCIsolationFlag = conf.NewBoolFlag(
		"experiment_run_caffe_with_llc_isolation",
		"If set, the Caffe workload will use the same isolation settings as for LLC aggressors, otherwise swan won't apply any performance isolation. User can use this flag to compare running task on separate cores and using OS scheduler.",
		true,
	)
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

	if threatAggressorsAsService.Value() {
		aggressor = executor.ServiceLauncher{Launcher: aggressor}
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

// PrepareAggressors prepare aggressors launchers.
// wrapped by session-less pair using given isolation and executor factory for aggressor workloads.
func PrepareAggressors(l1Isolation, llcIsolation isolation.Decorator, beExecutorFactory ExecutorFactoryFunc) (aggressorPairs []LauncherSessionPair, err error) {
	// Initialize aggressors with BE isolation wrapped as Snap session pairs.
	aggressorFactory := NewMultiIsolationAggressorFactory(l1Isolation, llcIsolation)

	for _, aggressorName := range AggressorsFlag.Value() {
		aggressorPair, err := aggressorFactory.Create(aggressorName, beExecutorFactory)
		if err != nil {
			return nil, err
		}

		var launcherSessionPair LauncherSessionPair

		switch aggressorName {
		case caffe.ID:
			caffeSession, err := caffeinferencesession.NewSessionLauncher(caffeinferencesession.DefaultConfig())
			if err != nil {
				return nil, err
			}
			launcherSessionPair = NewMonitoredLauncher(aggressorPair, caffeSession)
		case l1data.ID, l1instruction.ID, memoryBandwidth.ID, l3data.ID, stream.ID:
			launcherSessionPair = NewLauncherWithoutSession(aggressorPair)
		default:
			return nil, errors.Errorf("aggressor %q not found", aggressorName)
		}

		aggressorPairs = append(aggressorPairs, launcherSessionPair)
	}
	return
}
