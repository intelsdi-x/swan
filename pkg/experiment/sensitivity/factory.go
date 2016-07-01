package sensitivity

import (
	"fmt"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/caffe"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l1instruction"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/l3data"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/memoryBandwidth"
	"github.com/intelsdi-x/swan/pkg/workloads/low_level/stream"
)

// CreateAggressor returns aggressor of chosen type with Snap session.
func CreateAggressor(name string, executor executor.Executor) (LauncherSessionPair, error) {
	var aggressor LauncherSessionPair

	switch name {
	case l1data.ID:
		aggressor = NewLauncherWithoutSession(
			l1data.New(executor, l1data.DefaultL1dConfig()))
	case l1instruction.ID:
		aggressor = NewLauncherWithoutSession(
			l1instruction.New(executor, l1instruction.DefaultL1iConfig()))
	case memoryBandwidth.ID:
		aggressor = NewLauncherWithoutSession(
			memoryBandwidth.New(executor, memoryBandwidth.DefaultMemBwConfig()))
	case caffe.ID:
		aggressor = NewLauncherWithoutSession(
			caffe.New(executor, caffe.DefaultConfig()))
	case l3data.ID:
		aggressor = NewLauncherWithoutSession(
			l3data.New(executor, l3data.DefaultL3Config()))
	case stream.ID:
		aggressor = NewLauncherWithoutSession(
			stream.New(executor, stream.DefaultConfig()))
	default:
		return aggressor, fmt.Errorf("Aggressor '%s' not found", name)
	}

	return aggressor, nil
}
