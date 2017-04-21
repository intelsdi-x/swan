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
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

// GetPeakLoad runs tuning in order to determine the peak load.
func GetPeakLoad(hpLauncher executor.Launcher, loadGenerator executor.LoadGenerator, slo int) (peakLoad int, err error) {
	prTask, err := hpLauncher.Launch()
	if err != nil {
		return 0, errors.Wrap(err, "tunning: cannot launch high-priority task backend")
	}
	defer func() {
		// If function terminated with error then we do not want to overwrite it with any errors in defer.
		errStop := prTask.Stop()
		if err == nil {
			err = errStop
		}
	}()

	err = loadGenerator.Populate()
	if err != nil {
		return 0, errors.Wrap(err, "tunning: cannot populate high-priority task with data")
	}

	peakLoad, _, err = loadGenerator.Tune(slo)
	if err != nil {
		return 0, errors.Wrap(err, "tuning failed")
	}

	return
}
