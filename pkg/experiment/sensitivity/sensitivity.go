package sensitivity

import (
	"time"

	"github.com/intelsdi-x/swan/pkg/conf"
)

var (
	// SLOFlag indicates expected SLO
	SLOFlag = conf.NewIntFlag("slo", "Given SLO for the experiment. [us]", 500)
	// LoadPointsCountFlag represents number of load points per each aggressor
	LoadPointsCountFlag = conf.NewIntFlag("load_points", "Number of load points to test", 10)
	// LoadDurationFlag allows us to set repetition duration from command line argument or environmental variable
	LoadDurationFlag = conf.NewDurationFlag("load_duration", "Load duration [s].", 10*time.Second)
	// RepetitionsFlag indicates number of repetitions per each load point
	RepetitionsFlag = conf.NewIntFlag("reps", "Number of repetitions for each measurement", 3)
	// StopOnErrorFlag forces experiment to terminate on error
	StopOnErrorFlag = conf.NewBoolFlag("stop", "Stop experiment in a case of error", false)
	// PeakLoadFlag represents special case when peak load is provided instead of calculated from Tuning phase
	// It omits tuning phase.
	PeakLoadFlag = conf.NewIntFlag("peak_load", "Peakload max number of QPS without violating SLO (by default inducted from tuning phase).", 0) // "0" means include tuning phase
)

const (
	// RunTuningPhase represents PeakLoadFlag value indicating that tuning phase should be run.
	RunTuningPhase = 0
)
