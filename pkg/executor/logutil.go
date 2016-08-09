package executor

import (
	"bufio"
	"fmt"
	"io"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

// LogLinesCount is the number of lines printed from stderr & stdout in case of task failure.
var LogLinesCount = conf.NewIntFlag("output_lines_count", "Number of lines printed from stderr & stdout in case of task unsucessful termination", 5)

// LogSuccessfulExecution is helper function for logging standard output and standard error
// file names
func LogSuccessfulExecution(whatWasExecuted string, whereWasExecuted string, handle TaskHandle) {
	id := rand.Intn(9999)
	logrus.Debugf("%4d Process %q on %q on %q has ended\n", id, whatWasExecuted, whereWasExecuted, handle.Address())

	exitCode, err := handle.ExitCode()
	if err != nil {
		logrus.Debugf("%4d Could not read exit code: %v", id, err)
	} else {
		logrus.Debugf("%4d Exit code: %d", id, exitCode)
	}
}

// LogUnsucessfulExecution is helper function for logging standard output and standard error
// of task handles
func LogUnsucessfulExecution(whatWasExecuted string, whereWasExecuted string, handle TaskHandle) {
	var stdoutTail, stderrTail string
	var err error

	lineCount := LogLinesCount.Value()

	stdoutFile, stdoutErr := handle.StdoutFile()
	stderrFile, stderrErr := handle.StderrFile()

	if stdoutErr == nil {
		stdoutTail, err = readTail(stdoutFile.Name(), lineCount)
		if err != nil {
			stdoutErr = err
		}
	}
	if stderrErr == nil {
		stdoutTail, err = readTail(stderrFile.Name(), lineCount)
		if err != nil {
			stderrErr = err
		}
	}

	id := rand.Intn(9999)

	logrus.Errorf("%4d Command %q might have ended prematurely on %q on address %q", id, whatWasExecuted, whereWasExecuted, handle.Address())
	if stdoutErr == nil {
		logrus.Errorf("%4d Last %d lines of stdout", id, lineCount)
		errorLogLines(strings.NewReader(stdoutTail), id)
	} else {
		logrus.Errorf("%4d could not read stdout: %v", id, stdoutErr)
	}

	if stderrErr == nil {
		logrus.Errorf("%4d Last %d lines of stderr", id, lineCount)
		errorLogLines(strings.NewReader(stderrTail), id)
	} else {
		logrus.Errorf("%4d could not read stderr: %v", id, stderrErr)
	}

	exitCode, err := handle.ExitCode()
	if err != nil {
		logrus.Errorf("%4d Could not read exit code: %v", id, err)
	} else {
		logrus.Errorf("%4d Exit code: %d", id, exitCode)
	}
}

// ErrorLogLines takes reader and some ID (eg. PID) and prints each line
// from reader in a separate log.Errorf("%4d <line>", pid, line) .
// Rationale behind this function is fact, that logrus does not support multi-line logs.
func errorLogLines(r io.Reader, logID int) {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		logrus.Errorf("%4d %s", logID, scanner.Text())
	}
	err := scanner.Err()
	if err != nil {
		logrus.Errorf("%4d Printing from reader failed: %q", logID, err.Error())
	}
}

func readTail(filePath string, lineCount int) (tail string, err error) {
	lineCountParam := fmt.Sprintf("-n %d", lineCount)
	output, err := exec.Command("tail", lineCountParam, filePath).CombinedOutput()

	if err != nil {
		return "", errors.Wrapf(err, "could not read tail of %q", filePath)
	}

	return string(output), nil
}
