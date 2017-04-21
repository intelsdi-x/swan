// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
	// LoadGeneratorWaitTimeoutFlag is a flag that indicates how log experiment should wait for load generator to stop
	LoadGeneratorWaitTimeoutFlag = conf.NewDurationFlag("load_generator_wait_timeout", "amount of time to wait for load generator to stop before stopping it forcefully", 0)
)

const (
	// RunTuningPhase represents PeakLoadFlag value indicating that tuning phase should be run.
	RunTuningPhase = 0
)
