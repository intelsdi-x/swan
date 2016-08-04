package executor

import (
	"bufio"
	"fmt"
	"math/rand"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/fs"
)

// LogSuccessfulExecution is helper function for logging standard output and standard error
// file names
func LogSuccessfulExecution(whatWasExecuted string, whereWasExecuted string, handle TaskHandle) {
	id := rand.Intn(9999)

	var stdoutFileName string
	var stderrFileName string

	stdoutFile, err := handle.StdoutFile()
	if err != nil {
		logrus.Errorf("Could not read stdout filename for command %s on %s", whatWasExecuted, whereWasExecuted)
		stdoutFileName = fmt.Sprintf("%v", err)
	} else {
		stdoutFileName = stdoutFile.Name()
	}

	stderrFile, err := handle.StderrFile()
	if err != nil {
		logrus.Errorf("Could not read stderr filename for command %s on %s", whatWasExecuted, whereWasExecuted)
		stderrFileName = fmt.Sprintf("%v", err)
	} else {
		stderrFileName = stderrFile.Name()
	}

	logrus.Debugf("%4d Process %q on %q on %q has ended\n", id, whatWasExecuted, whereWasExecuted, handle.Address())
	logrus.Debugf("%4d Stdout stored in %q", id, stdoutFileName)
	logrus.Debugf("%4d Stderr stored in %q", id, stderrFileName)

	exitCode, err := handle.ExitCode()
	if err != nil {
		logrus.Debugf("%4d Could not read exit code: %v", err)
	} else {
		logrus.Debugf("%4d Exit code: %d", id, exitCode)
	}
}

// LogUnsucessfulExecution is helper function for logging standard output and standard error
// of task handles
func LogUnsucessfulExecution(whatWasExecuted string, whereWasExecuted string, handle TaskHandle) {
	var stdoutFileName string
	var stderrFileName string

	stdoutFile, err := handle.StdoutFile()
	if err != nil {
		logrus.Errorf("Could not read stdout filename for command %s on %s", whatWasExecuted, whereWasExecuted)
		stdoutFileName = fmt.Sprintf("%v", err)
	} else {
		stdoutFileName = stdoutFile.Name()
	}

	stderrFile, err := handle.StderrFile()
	if err != nil {
		logrus.Errorf("Could not read stderr filename for command %s on %s", whatWasExecuted, whereWasExecuted)
		stderrFileName = fmt.Sprintf("%v", err)
	} else {
		stderrFileName = stderrFile.Name()
	}

	lineCount := 3
	stdoutTail, err := fs.ReadTail(stdoutFileName, lineCount)
	if err != nil {
		stdoutTail = fmt.Sprintf("%v", err)
	}
	stderrTail, err := fs.ReadTail(stderrFileName, lineCount)
	if err != nil {
		stderrTail = fmt.Sprintf("%v", err)
	}

	id := rand.Intn(9999)
	logrus.Errorf("%4d Command %q might have ended prematurely on %q on address %q", id, whatWasExecuted, whereWasExecuted, handle.Address())
	logrus.Errorf("%4d Stdout stored in %q", id, stdoutFileName)
	logrus.Errorf("%4d Stderr stored in %q", id, stderrFileName)
	logrus.Errorf("%4d Last %d lines of stdout", id, lineCount)
	ErrorLogLines(strings.NewReader(stdoutTail), id)
	logrus.Errorf("%4d Last %d lines of stderr", id, lineCount)
	ErrorLogLines(strings.NewReader(stderrTail), id)

	exitCode, err := handle.ExitCode()
	if err != nil {
		logrus.Errorf("%4d Could not read exit code: %v", err)
	} else {
		logrus.Errorf("%4d Exit code: %d", id, exitCode)
	}
}

// ErrorLogLines takes reader and some ID (eg. PID) and prints each line
// from reader in a separate log.Errorf("%4d <line>", pid, line) .
// Rationale behind this function is fact, that logrus does not support multi-line logs.
func ErrorLogLines(r *strings.Reader, logID int) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logrus.Errorf("%4d %s", logID, scanner.Text())
	}
	err := scanner.Err()
	if err != nil {
		logrus.Errorf("%4d Printing from reader failed: %q", logID, err.Error())
	}
}
