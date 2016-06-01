package cgroup

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
)

// Subsys returns true if the named subsystem is mounted.
func Subsys(name string, executor executor.Executor, timeout time.Duration) (bool, error) {
	mounts, err := SubsysMounts(executor, timeout)
	if err != nil {
		return false, err
	}
	_, found := mounts[name]
	return found, nil
}

// SubsysPath returns the absolute path where the supplied subsystem is
// mounted. Returns the empty string if the subsystem is not mounted.
func SubsysPath(name string, executor executor.Executor, timeout time.Duration) (string, error) {
	mounts, err := SubsysMounts(executor, timeout)
	if err != nil {
		return "", err
	}
	return mounts[name], nil
}

// SubsysMounts returns a map of cgroup subsystem controller names to
// mount points in the file system.
func SubsysMounts(executor executor.Executor, timeout time.Duration) (map[string]string, error) {
	out, err := cmdOutput(executor, timeout, "lssubsys", "--all-mount-points")
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		var name, mount string
		_, err := fmt.Sscanf(line, "%s %s", &name, &mount)
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		// The name part may indicate co-mounted subsystems (e.g. "cpu,cpuacct").
		// Let's save them separately to make them easier to find.
		names := strings.Split(name, ",")
		for _, n := range names {
			result[n] = mount
		}
	}
	return result, nil
}
