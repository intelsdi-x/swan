package executor

import (
	"bufio"
	"fmt"
	"math/rand"
	"os/exec"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
	"os"
)

// LogLinesCount is the number of lines printed from stderr & stdout in case of task failure.
var LogLinesCount = conf.NewIntFlag("output_lines_count", "Number of lines printed from stderr & stdout in case of task unsucessful termination", 3)

// LogSuccessfulExecution is helper function for logging standard output and standard error
// file names
func LogSuccessfulExecution(whatWasExecuted string, whereWasExecuted string, handle TaskHandle) {
	stdoutFileName, stderrFileName := readStdoutFilenames(handle, whatWasExecuted, whereWasExecuted)

	id := rand.Intn(9999)
	logrus.Debugf("%4d Process %q on %q on %q has ended\n", id, whatWasExecuted, whereWasExecuted, handle.Address())
	logrus.Debugf("%4d Stdout stored in %q", id, stdoutFileName)
	logrus.Debugf("%4d Stderr stored in %q", id, stderrFileName)

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
	lineCount := LogLinesCount.Value()
	stdoutFileName, stderrFileName := readStdoutFilenames(handle, whatWasExecuted, whereWasExecuted)
	stdoutTail, stderrTail := readTails(stdoutFileName, stderrFileName, lineCount)

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
		logrus.Errorf("%4d Could not read exit code: %v", id, err)
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

func readStdoutFilenames(handle TaskHandle, what string, where string) (stdoutFileName string, stderrFileName string) {
	stdoutFile, err := handle.StdoutFile()
	if err != nil {
		stdoutFileName = fmt.Sprintf("Could not read stdout filename for command %s on %s: %v", what, where, err)
	} else {
		stdoutFileName = stdoutFile.Name()
	}

	stderrFile, err := handle.StderrFile()
	if err != nil {
		stdoutFileName = fmt.Sprintf("Could not read stderr filename for command %s on %s: %v", what, where, err)
	} else {
		stderrFileName = stderrFile.Name()
	}

	return stdoutFileName, stderrFileName
}

func readTails(stdoutFileName string, stderrFileName string, lineCount int) (stdoutFileTail string, stderrFileTail string) {
	stdoutTail, err := readTail(stdoutFileName, lineCount)
	if err != nil {
		stdoutTail = fmt.Sprintf("%v", err)
	}
	stderrTail, err := readTail(stderrFileName, lineCount)
	if err != nil {
		stderrTail = fmt.Sprintf("%v", err)
	}

	return stdoutTail, stderrTail
}

func readTail(filePath string, lineCount int) (tail string, err error) {
	_, err = os.Stat(filePath)
	if err != nil {
		return "", errors.New("file does not exists")
	}

	lineCountParam := fmt.Sprintf("-n %d", lineCount)
	output, err := exec.Command("tail", lineCountParam, filePath).CombinedOutput()

	if err != nil {
		return "", errors.Wrapf(err, "could not read tail of %q", filePath)
	}

	return string(output), nil
}
