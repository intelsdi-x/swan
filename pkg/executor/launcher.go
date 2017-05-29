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

import "fmt"

// Launcher responsibility is to launch previously configured job.
type Launcher interface {
	fmt.Stringer
	// Launch starts the workload (process or group of processes). It returns a workload
	// represented as a Task Handle instance.
	// Error is returned when Launcher is unable to start a job.
	Launch() (TaskHandle, error)
}
