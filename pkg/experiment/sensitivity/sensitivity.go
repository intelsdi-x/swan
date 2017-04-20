package sensitivity

import (
	"time"

	"github.com/intelsdi-x/swan/pkg/conf"
)

const (
	// RunTuningPhase represents PeakLoadFlag value indicating that tuning phase should be run.
	RunTuningPhase = 0
)

var (
	// SLOFlag indicates expected SLO
	SLOFlag = conf.NewIntFlag("experiment_slo", "Given SLO for the HP workload in experiment. [us]", 500)
	// LoadPointsCountFlag represents number of load points per each aggressor
	LoadPointsCountFlag = conf.NewIntFlag("experiment_load_points", "Number of load points to test", 10)
	// LoadDurationFlag allows us to set repetition duration from command line argument or environmental variable
	LoadDurationFlag = conf.NewDurationFlag("experiment_load_duration", "Load duration on HP task.", 15*time.Second)
	// RepetitionsFlag indicates number of repetitions per each load point
	RepetitionsFlag = conf.NewIntFlag("experiment_repetitions", "Number of repetitions for each measurement", 3)
	// StopOnErrorFlag forces experiment to terminate on error
	StopOnErrorFlag = conf.NewBoolFlag("experiment_stop_on_error", "Stop experiment in a case of error", false)
	// PeakLoadFlag represents special case when peak load is provided instead of calculated from Tuning phase.
	PeakLoadFlag = conf.NewIntFlag("experiment_peak_load", "Maximum load that will be generated on HP workload. If value is `0`, then maximum possible load will be found by Swan.", RunTuningPhase)
	// LoadGeneratorWaitTimeoutFlag is a flag that indicates how log experiment should wait for load generator to stop
	LoadGeneratorWaitTimeoutFlag = conf.NewDurationFlag("experiment_load_generator_wait_timeout", "Amount of time to wait for load generator to stop before stopping it forcefully. In succesuful case, it should stop on it's own.", 0)
)
