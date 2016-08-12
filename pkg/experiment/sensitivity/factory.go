package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/launchers/caffe"
	"github.com/intelsdi-x/swan/pkg/launchers/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/launchers/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/launchers/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/launchers/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/launchers/low_level/stream"
	"github.com/pkg/errors"
)

const (
	l1dDefaultProcessNumber = 1
	l1iDefaultProcessNumber = 1
	l3DefaultProcessNumber  = 1
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

// Create returns aggressor of chosen type with Snap session.
func (f AggressorFactory) Create(name string) (LauncherSessionPair, error) {
	var aggressor LauncherSessionPair

	exec := f.createIsolatedExecutor(name)

	switch name {
	case l1data.ID:
		aggressor = NewLauncherWithoutSession(
			l1data.New(exec, l1data.DefaultL1dConfig()))
	case l1instruction.ID:
		aggressor = NewLauncherWithoutSession(
			l1instruction.New(exec, l1instruction.DefaultL1iConfig()))
	case memoryBandwidth.ID:
		aggressor = NewLauncherWithoutSession(
			memoryBandwidth.New(exec, memoryBandwidth.DefaultMemBwConfig()))
	case caffe.ID:
		aggressor = NewLauncherWithoutSession(
			caffe.New(exec, caffe.DefaultConfig()))
	case l3data.ID:
		aggressor = NewLauncherWithoutSession(
			l3data.New(exec, l3data.DefaultL3Config()))
	case stream.ID:
		aggressor = NewLauncherWithoutSession(
			stream.New(exec, stream.DefaultConfig()))
	default:
		return aggressor, errors.Errorf("aggressor %q not found", name)
	}

	return aggressor, nil
}

// CreateIsolatedExecutor returns local executor with prepared isolation.
// L1-Data and L1-Instruction cache aggressor receives l1AggressorIsolation
// Other aggressors receive otherAggressorIsolation
func (f AggressorFactory) createIsolatedExecutor(name string) executor.Executor {
	switch name {
	case l1data.ID:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		if L1dProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L1dProcessNumber.Value()))
		}
		return executor.NewLocalIsolated(decorators)
	case l1instruction.ID:
		decorators := isolation.Decorators{f.l1AggressorIsolation}
		if L1iProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L1iProcessNumber.Value()))
		}
		return executor.NewLocalIsolated(decorators)
	case l3data.ID:
		decorators := isolation.Decorators{f.otherAggressorIsolation}
		if L3ProcessNumber.Value() != 1 {
			decorators = append(decorators, executor.NewParallel(L3ProcessNumber.Value()))
		}
		return executor.NewLocalIsolated(decorators)
	default:
		return executor.NewLocalIsolated(f.otherAggressorIsolation)
	}
}
