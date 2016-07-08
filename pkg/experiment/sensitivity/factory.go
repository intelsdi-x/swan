package sensitivity

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
	"github.com/pkg/errors"
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
		return executor.NewLocalIsolated(f.l1AggressorIsolation)
	case l1instruction.ID:
		return executor.NewLocalIsolated(f.l1AggressorIsolation)
	default:
		return executor.NewLocalIsolated(f.otherAggressorIsolation)
	}
}
