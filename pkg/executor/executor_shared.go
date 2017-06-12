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

	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const killWaitTimeout = 1 * time.Second

// checkIfProcessFailedToExecute should be checked in the end of Execute(cmd) method.
// It checks if command execution failed and returns nil handle and error.
// If task is still running or exit code is equal to 0, it returns nil error.
//
// Commands usually fail because wrong parameters or binary that should be executed is not installed properly.
func checkIfProcessFailedToExecute(command string, executorName string, handle TaskHandle) error {
	if handle.Status() == TERMINATED {
		exitCode, err := handle.ExitCode()
		if err != nil {
			// Something really wrong happened, print error message + logs
			log.Errorf("Task %q launched on %q on address %q has failed, cannot get exit code: %s", command, executorName, handle.Address(), err.Error())
			logOutput(handle)
			return errors.Errorf("task %q launched using %q on address %q has failed, cannot get exit code: %s", command, executorName, handle.Address(), err.Error())
		}
		if exitCode != 0 {
			// Task failed, log.Error exit code & stdout/err
			log.Errorf("Task %q launched using %q on address %q has failed: exit code %d", command, executorName, handle.Address(), exitCode)
			logOutput(handle)
			return errors.Errorf("task %q launched using %q on address %q has failed with exit code %d", command, executorName, handle.Address(), exitCode)
		}

		// Exit code is zero, so task ended successfully.
		log.Debugf("task %q launched using %q on address %q has ended successfully (exit code: 0)", command, executorName, handle.Address())
		return nil
	}

	return nil
}

// getWaitTimeoutChan returns channel for timeout in Wait(timeout) function in TaskHandles.
// When timeout is 0, then timeout will never occur.
func getWaitTimeoutChan(timeout time.Duration) <-chan time.Time {
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		return time.After(timeout)
	}

	return make(<-chan time.Time)
}

// getWaitChannel returns channel that will return result (any encountered error) of
// Wait() method in provided handle.
func getWaitChannel(handle TaskControl) <-chan error {
	result := make(chan error)
	go func() {
		_, err := handle.Wait(0)
		result <- err
		return
	}()
	return result
}
