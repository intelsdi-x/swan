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

package experiment

import (
	"os"
	"path"
	"strconv"

	"github.com/pkg/errors"
)

const (
	// ExperimentKey defines the key for Snap tag.
	ExperimentKey = "swan_experiment"
	// PhaseKey defines the key for Snap tag.
	PhaseKey = "swan_phase"
	// RepetitionKey defines the key for Snap tag.
	RepetitionKey = "swan_repetition"
	// LoadPointQPSKey defines the key for Snap tag.
	LoadPointQPSKey = "swan_loadpoint_qps"
	// AggressorNameKey defines the key for Snap tag.
	AggressorNameKey = "swan_aggressor_name"

	// See /usr/include/sysexits.h for refference regarding constants below

	// ExUsage reperense command line user error exit code
	ExUsage = 64
	// ExSoftware represents internal software error exit code
	ExSoftware = 70
	// ExIOErr represents input/output error exit code
	ExIOErr = 74
)

// CreateExperimentDir creates directory structure for the experiment.
func CreateExperimentDir(uuid, appName string) (experimentDirectory string, logFile *os.File, err error) {
	experimentDirectory = createExperimentLogsDirectoryName(appName, uuid)
	err = os.MkdirAll(experimentDirectory, 0777)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "cannot create experiment directory: ", experimentDirectory)
	}
	err = os.Chdir(experimentDirectory)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "cannot chdir to experiment directory", experimentDirectory)
	}

	masterLogFilename := path.Join(experimentDirectory, "master.log")
	logFile, err = os.OpenFile(masterLogFilename, os.O_WRONLY|os.O_CREATE, 0755)
	if err != nil {
		return "", &os.File{}, errors.Wrapf(err, "could not open log file %q", masterLogFilename)
	}

	return experimentDirectory, logFile, nil
}

func createExperimentLogsDirectoryName(appName, uuid string) string {
	return path.Join(os.TempDir(), appName, uuid)
}

// CreateRepetitionDir creates folders that store repetition logs inside experiment's directory.
func CreateRepetitionDir(appName, uuid, phaseName string, repetition int) error {
	experimentDirectory := createExperimentLogsDirectoryName(appName, uuid)
	repetitionDir := path.Join(experimentDirectory, phaseName, strconv.Itoa(repetition))
	err := os.MkdirAll(repetitionDir, 0777)
	if err != nil {
		return errors.Wrapf(err, "could not create dir %q", repetitionDir)
	}

	err = os.Chdir(repetitionDir)
	if err != nil {
		return errors.Wrapf(err, "could not change to dir %q", repetitionDir)
	}

	return nil
}
