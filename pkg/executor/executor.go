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

// Executor is responsible for creating execution environment for given workload.
// It returns Task handle when workload started gracefully.
// Workload is executed asynchronously.
type Executor interface {
	fmt.Stringer
	// Execute executes command on underlying platform.
	// Invokes "bash -c <command>" and waits for short time to make sure that process has started.
	// Returns error if command exited immediately with non-zero exit status.
	Execute(command string) (TaskHandle, error)
}
