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

package executor

import (
	"time"
)

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
	Load(load int, duration time.Duration) (task TaskHandle, err error)
}
