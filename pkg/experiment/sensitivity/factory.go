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
)

// CreateAggressor returns aggressor of chosen type with Snap session and isolation.
func CreateAggressor(name string, isolation isolation.Isolation) (LauncherSessionPair, error) {
	var aggressor LauncherSessionPair
	executor := executor.NewLocalIsolated(isolation)

	switch name {
	case l1data.ID:
		aggressor = NewLauncherWithoutSession(
			l1data.New(executor, l1data.DefaultL1dConfig()))
		break
	case l1instruction.ID:
		aggressor = NewLauncherWithoutSession(
			l1instruction.New(executor, l1instruction.DefaultL1iConfig()))
		break
	case memoryBandwidth.ID:
		aggressor = NewLauncherWithoutSession(
			memoryBandwidth.New(executor, memoryBandwidth.DefaultMemBwConfig()))
		break
	case caffe.ID:
		aggressor = NewLauncherWithoutSession(
			caffe.New(executor, caffe.DefaultConfig()))
		break
	case l3data.ID:
		aggressor = NewLauncherWithoutSession(
			l3data.New(executor, l3data.DefaultL3Config()))
		break
	default:
		return aggressor, fmt.Errorf("Aggressor '%s' not found", name)
	}

	return aggressor, nil
}
