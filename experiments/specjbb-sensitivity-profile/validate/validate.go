package validate

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/utils/errutil"
)

// CheckCPUPowerGovernor warn user about potential issues with performance when powersave governor is used.
// procfs: https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-devices-system-cpu
// governor path: https://www.kernel.org/doc/Documentation/cpu-freq/user-guide.txt
// performance,powersave constants: http://lxr.free-electrons.com/source/drivers/cpufreq/cpufreq.c#L484
func CheckCPUPowerGovernor() {
	const cpu0GovernorFile = "/sys/devices/system/cpu/cpu0/cpufreq/scaling_governor" // Assume at least one CPU exists!.
	if _, err := os.Stat(cpu0GovernorFile); os.IsNotExist(err) {
		logrus.Warnf("Validation of CPU power governor failed! - %q not available (check `dmesg | grep acpi_cpufreq` entry for hardware support).", cpu0GovernorFile)
		return
	}
	const performance = "performance"
	for i := 0; i < runtime.NumCPU(); i++ {
		cpuGovernorFile := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor", i)
		governorBytes, err := ioutil.ReadFile(cpuGovernorFile)
		governor := strings.TrimSuffix(string(governorBytes), "\n")
		errutil.Check(err)
		logrus.Debugf("governor cpu%d: %q", i, governor)
		if string(governor) != performance {
			logrus.Warnf("scaling_governor=%q (%q) should be set to 'performance' policy to mitigate wakeup penalty (causes variability in measurements at moderate load). You can change this value with 'cpupower frequency-set -g performance'as root.", governor, cpuGovernorFile)
		}
	}
}

// OS checks experiment local OS environment to help identify potential issues.
// Note: in case of some requirements not met, only warns user.
func OS() {
	CheckCPUPowerGovernor()
}
