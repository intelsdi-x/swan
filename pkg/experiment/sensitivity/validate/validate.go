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

package validate

import (
	"fmt"
	"io/ioutil"
	"os"
	"runtime"
	"strconv"
	"strings"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/sysctl"
)

const (
	// The minimal value for the maximum number of open file descriptors that will
	// be enough to handle distributed mutilate cluster when generating sensible load
	// and enough to run production tasks like memcached that handle a lot of
	// number simultaneous connections.
	minimalNOFILERequirement = 10 * 1024
)

// checkTCPSyncookies warn user about potential issue with SYN flooding of victim machine.
func checkTCPSyncookies() {
	value, err := sysctl.Get("net.ipv4.tcp_syncookies")
	if err != nil {
		logrus.Debug("Could not read net.ipv4.tcp_syncookies sysctl key: " + err.Error())
	} else if value == "1" {
		logrus.Warn("net.ipv4.tcp_syncookies is enabled on the memcached target and may lead to SYN flooding detection closing mutilate connections (you can change this by 'echo 0 > /proc/sys/net/ipv4/tcp_syncookies' as root).")
	}
	logrus.Debugf("net.ipv4.tcp_syncookies sysctl value: %q ", value)
}

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

// checkNOFILE checks if the number of maximum file descriptors
// opened by a process is more than minimum requested.
// The name NOFILE is based on "limits.conf" and definition from setrlimit.
func checkNOFILE(nofile, minimum int) {
	if nofile <= minimum {
		logrus.Warnf("Maximum number of open file descriptors (%d) is lower than required (%d). You can change this value eg. ulimit -n 10000 or modifying /etc/security/limits.conf.", nofile, minimum)
	}

}

// OS checks experiment local OS environment to help identify potential issues.
// Note: in case of some requirements not met, only warns user.
func OS() {
	checkTCPSyncookies()
	CheckCPUPowerGovernor()
	checkNOFILE(
		getNOFILE(executor.NewLocal()),
		minimalNOFILERequirement,
	)
}

// ExecutorsNOFILELimit validates if environment provided by executors can run
// distributed application that requires large number of open file descriptors.
func ExecutorsNOFILELimit(executors []executor.Executor) {
	for _, executor := range executors {
		checkNOFILE(
			getNOFILE(executor),
			minimalNOFILERequirement,
		)
	}
}

// getNOFILE is helper to retrieve resource limit for NOFILE using given executor.
func getNOFILE(executor executor.Executor) int {

	// Run ulimit and wait.
	taskHandle, err := executor.Execute("ulimit -n")
	errutil.Check(err)
	defer taskHandle.EraseOutput()
	taskHandle.Wait(0)

	// Retrieve output.
	outFile, err := taskHandle.StdoutFile()
	errutil.Check(err)
	output, err := ioutil.ReadAll(outFile)
	errutil.Check(err)

	// Parse and return.
	nofile, err := strconv.Atoi(strings.Trim(string(output), "\n\r"))
	errutil.Check(err)

	return nofile
}
