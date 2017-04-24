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
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
	"github.com/pkg/errors"
)

// LogLinesCount is the number of lines printed from stderr & stdout in case of task failure.
var LogLinesCount = conf.NewIntFlag("output_lines_count", "Number of lines printed from stderr & stdout in case of task unsucessful termination", 5)

const outputFilePrivileges = os.FileMode(0644)

func getBinaryNameFromCommand(command string) (string, error) {
	argsSplit := strings.Split(command, " ")
	if len(argsSplit) == 0 {
		return "", errors.Errorf("failed to extract command name from %q", command)
	}
	_, name := path.Split(argsSplit[0])
	return name, nil
}

// createOutputDirectory creates directory for executor output and returns path to it when successful, or error if not.
func createOutputDirectory(command string, prefix string) (createdDirectoryPath string, err error) {
	if len(command) == 0 {
		return "", errors.New("empty command string")
	}

	commandName, err := getBinaryNameFromCommand(command)
	if err != nil {
		return "", err
	}
	directoryPrivileges := os.FileMode(0755)

	pwd, err := os.Getwd()
	if err != nil {
		return "", errors.Wrap(err, "failed to get working directory")
	}
	createdDirectoryPath, err = ioutil.TempDir(pwd, prefix+"_"+commandName+"_")
	if err != nil {
		return "", errors.Wrapf(err, "failed to create output directory for %q", commandName)
	}
	if err = os.Chmod(createdDirectoryPath, directoryPrivileges); err != nil {
		os.RemoveAll(createdDirectoryPath)
		return "", errors.Wrapf(err, "failed to set privileges for dir %q", createdDirectoryPath)
	}

	return createdDirectoryPath, nil
}

func createExecutorOutputFiles(outputDir string) (stdoutFile, stderrFile *os.File, err error) {
	stdoutFileName := path.Join(outputDir, "stdout")
	stdoutFile, err = os.OpenFile(stdoutFileName, os.O_WRONLY|os.O_SYNC|os.O_EXCL|os.O_CREATE, outputFilePrivileges)
	if err != nil {
		return nil, nil, errors.Wrapf(err, "creating %q failed", stdoutFileName)
	}

	stderrFileName := path.Join(outputDir, "stderr")
	stderrFile, err = os.OpenFile(stderrFileName, os.O_WRONLY|os.O_SYNC|os.O_EXCL|os.O_CREATE, outputFilePrivileges)
	if err != nil {
		// Clean created stdout.
		stdoutFile.Close()
		os.Remove(stdoutFileName)
		return nil, nil, errors.Wrapf(err, "os.Create failed for path %q", stderrFileName)
	}

	return stdoutFile, stderrFile, nil
}

func syncAndClose(file *os.File) error {
	var errCol errcollection.ErrorCollection
	err := file.Sync()
	if err != nil {
		errCol.Add(err)
		log.Errorf("Cannnot sync stdout file: %s", err.Error())
	}
	err = file.Close()
	if err != nil {
		errCol.Add(err)
		log.Errorf("Cannot close stdout file: %s", err.Error())
	}
	return errCol.GetErrIfAny()
}

func openFile(fileName string) (*os.File, error) {
	if _, err := os.Stat(fileName); err != nil {
		return nil, errors.Wrapf(err, "unable to stat file at %q", fileName)
	}

	file, err := os.Open(fileName)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to open file at %q", fileName)
	}

	return file, nil
}

// removeDirectory removes directory if exists.
func removeDirectory(directory string) error {
	if _, err := os.Stat(directory); os.IsNotExist(err) {
		return nil
	}

	if err := os.RemoveAll(directory); err != nil {
		return errors.Wrapf(err, "os.RemoveAll of directory %q failed", directory)
	}
	return nil
}

func logOutput(th TaskHandle) error {
	lines := LogLinesCount.Value()
	file, err := th.StdoutFile()
	if err == nil {
		stat, err := file.Stat()
		if err != nil {
			log.Errorf("Could not stat stdout file: %q", err.Error())
		}

		if stat.Size() == 0 {
			log.Errorf("Stdout file is empty")
		} else {
			stdout, err := tailFile(file.Name(), lines)
			if err != nil {
				log.Errorf("Tailing stdout file failed: %q", err.Error())
			}

			log.Errorf("Last %d lines of stdout: %s", lines, stdout)
		}
	} else {
		log.Errorf("Cannot retrieve stdout file: %q", err.Error())
	}

	file, err = th.StderrFile()
	if err == nil {
		stat, err := file.Stat()
		if err != nil {
			log.Errorf("Could not stat stdout file: %q", err.Error())
		}

		if stat.Size() == 0 {
			log.Errorf("Stderr file is empty")
		} else {
			stderr, err := tailFile(file.Name(), lines)
			if err != nil {
				log.Errorf("Tailing stderr file failed: %q", err.Error())
			}

			log.Errorf("Last %d lines of stderr: %s", lines, stderr)
		}
	} else {
		log.Errorf("Cannot retrieve stderr file: %q", err.Error())
	}

	return nil
}

func tailFile(filePath string, lineCount int) (tail string, err error) {
	lineCountParam := fmt.Sprintf("-n %d", lineCount)
	output, err := exec.Command("tail", lineCountParam, filePath).CombinedOutput()

	if err != nil {
		return "", errors.Wrapf(err, "could not read tail of %q", filePath)
	}

	return "\n" + string(output), nil
}
