package workloads

import (
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/executor"
	"time"
)

var loadGeneratorAddrFlag = conf.NewFlag(
	"load_generator_addr", "IP of the target machine for the load generator", "127.0.0.1")

// LoadGeneratorAddrFlag represents load generator address flag.
func LoadGeneratorAddrFlag() *string {
	return conf.RegisterStringFlag(loadGeneratorAddrFlag)
}

// LoadGenerator launches stresser which generates load on specified workload.
type LoadGenerator interface {
	// Populate inserts initial data.
	Populate() error

	// Tune does the tuning phase which is a process of searching for a targetQPS
	// for given SLO.
	Tune(slo int) (achievedLoad int, achievedSLI int, err error)

	// Load starts a load on the specific workload with the defined loadPoint (number of QPS).
	// The task will do the load for specified amount of time.
	// Note: Results from Load needs to be fetched out of band e.g using Snap.
	Load(load int, duration time.Duration) (task executor.TaskHandle, err error)
}
