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
	"os"
	"time"

	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
)

// TaskState is an enum presenting current task state.
type TaskState int

const (
	// RUNNING task state means that task is still running.
	RUNNING TaskState = iota
	// TERMINATED task state means that task completed or stopped.
	TERMINATED
)

// TaskHandle represents an abstraction to control task lifecycle and status.
type TaskHandle interface {
	TaskInfo
	TaskControl
}

// TaskInfo represents task's address, status and output information.
type TaskInfo interface {
	// Name returns information about command and executor that lives under this TaskHandle.
	Name() string
	// Location returns address where task was located.
	Address() string
	// ExitCode returns a exitCode. If task is not terminated it returns error.
	ExitCode() (int, error)
	// Status returns a state of the task.
	Status() TaskState
	// StdoutFile returns a file handle for file to the task's stdout file.
	StdoutFile() (*os.File, error)
	// StderrFile returns a file handle for file to the task's stderr file.
	StderrFile() (*os.File, error)
}

// TaskControl controls task's lifecycle and garbage collection.
type TaskControl interface {
	// Stops a task.
	Stop() error
	// Wait does the blocking wait. It is a helper for waiting with a given timeout time.
	// It returns true if task is terminated.
	// In case of '0*time.Nanoseconds' timeout it waits infinitely for task completion.
	Wait(timeout time.Duration) bool
	// EraseOutput deletes the directory where output files resides.
	EraseOutput() error
}

// StopAndEraseOutput run stop and eraseOutput on taskHandle and add errors to errorCollection
func StopAndEraseOutput(handle TaskHandle) (errorCollection errcollection.ErrorCollection) {
	if handle != nil {
		errorCollection.Add(handle.Stop())
		errorCollection.Add(handle.EraseOutput())
	}

	return
}
