package executor

import (
	log "github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// checkIfProcessFailedToExecute should be checked in the end of Execute(cmd) method.
// It checks if command execution failed and returns nil handle and error.
// If task is still running or exit code is equal to 0, it returns nil error.
//
// Commands usually fail because wrong parameters or binary that should be executed is not installed properly.
func checkIfProcessFailedToExecute(command string, executorName string, handle TaskHandle) (TaskHandle, error) {
	if handle.Status() == TERMINATED {
		exitCode, err := handle.ExitCode()
		if err != nil {
			// Something really wrong happened, print error message + logs
			log.Errorf("task %q launched on %q failed, cannot get exit code: %s", command, executorName, err.Error())
			logOutput(handle)
			return nil, errors.Errorf("task %q launched on %q failed, cannot get exit code: %s", command, executorName, err.Error())
		}
		if exitCode != 0 {
			// Task failed, log.Error exit code & stdout/err
			log.Errorf("task %q launched on %q failed: exit code %d", command, executorName, exitCode)
			logOutput(handle)
			return nil, errors.Errorf("task %q launched on %q failed exit code %d", command, executorName, exitCode)
		}

		// Exit code is zero, so task ended successfully.
		log.Debugf("task %q launched on %q has ended successfully", command, executorName)
		return handle, nil
	}

	return handle, nil
}
