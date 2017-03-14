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
		prTask.Clean()
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
