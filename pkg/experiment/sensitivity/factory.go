package sensitivity

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
)

//AggressorFactory is a helper for creating aggressors with/without Snap sessions.
type AggressorFactory struct {
	executor executor.Executor
}

//NewAggressorFactory creates instance of AggressorFactory using isolation passed.
func NewAggressorFactory(isolation isolation.Isolation) *AggressorFactory {
	return &AggressorFactory{executor: executor.NewLocalIsolated(isolation)}
}

//Create sets up an aggresor of chosen type with Snap session.
func (af *AggressorFactory) Create(name string) (LauncherSessionPair, error) {
	var aggressor LauncherSessionPair
	switch name {
	case l1data.ID:
		aggressor = NewLauncherWithoutSession(
			l1data.New(af.executor, l1data.DefaultL1dConfig()))
	case l1instruction.ID:
		aggressor = NewLauncherWithoutSession(
			l1instruction.New(af.executor, l1instruction.DefaultL1iConfig()))
	case memoryBandwidth.ID:
		aggressor = NewLauncherWithoutSession(
			memoryBandwidth.New(af.executor, memoryBandwidth.DefaultMemBwConfig()))
	case caffe.ID:
		aggressor = NewLauncherWithoutSession(
			caffe.New(af.executor, caffe.DefaultConfig()))
	case l3data.ID:
		aggressor = NewLauncherWithoutSession(
			l3data.New(af.executor, l3data.DefaultL3Config()))
	case stream.ID:
		aggressor = NewLauncherWithoutSession(
			stream.New(af.executor, stream.DefaultConfig()))
	default:
		return aggressor, fmt.Errorf("Aggressor '%s' not found", name)
	}

	return aggressor, nil
}
