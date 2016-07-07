package main

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
	"syscall"

	"github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/utils/sysctl"
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

// checkCPUPower warn user about potential issues with performance when powersave governor is used.
// procfs: https://www.kernel.org/doc/Documentation/ABI/testing/sysfs-devices-system-cpu
// governor path: https://www.kernel.org/doc/Documentation/cpu-freq/user-guide.txt
// performance,powersave constants: http://lxr.free-electrons.com/source/drivers/cpufreq/cpufreq.c#L484
func checkCPUPowerGovernor() {
	const PERFORMANCE = "performance"
	for i := 0; i < runtime.NumCPU(); i++ {
		cpuGovernorFile := fmt.Sprintf("/sys/devices/system/cpu/cpu%d/cpufreq/scaling_governor", i)
		governorBytes, err := ioutil.ReadFile(cpuGovernorFile)
		governor := strings.TrimSuffix(string(governorBytes), "\n")
		check(err)
		logrus.Debugf("governor cpu%d: %q", i, governor)
		if string(governor) != PERFORMANCE {
			logrus.Warnf("scaling_governor=%q (%q) should be set to 'performance' policy to mitigate wakeup penalty (causes variability in measurements at moderate load). You can change this value with 'cpupower frequency-set -g performance'as root.", governor, cpuGovernorFile)
		}
	}
}

// checkMaximumNumberOfOpenDescriptors check maximum file descriptor number that can be opened by this process.
// Swan require at least to handle remote connections for mutilate cluster, but also it is inherited by workloads.
// Expect more than default 1024.
// http://man7.org/linux/man-pages/man2/setrlimit.2.html
func checkMaximumNumberOfOpenDescriptors() {
	rlimit := &syscall.Rlimit{}
	err := syscall.Getrlimit(syscall.RLIMIT_NOFILE, rlimit)
	check(err)
	logrus.Debugf("maximum file descriptor number: cur=%d (max=%d)", rlimit.Cur, rlimit.Max)
	if rlimit.Cur <= 1024 {
		logrus.Warnf("Maximum number of open file descriptors is low = %d. You can change this value eg. ulimit -n 100000 or modifying /etc/security/limits.conf.", rlimit.Cur)
	}

}

// validateOS check experiment local OS environment to help identify potential issues.
// Note: in case of some requirements not met, only warns user.
func validateOS() {
	checkTCPSyncookies()
	checkCPUPowerGovernor()
	checkMaximumNumberOfOpenDescriptors()
}
