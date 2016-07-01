package sysctl

import (
	"io/ioutil"
	"path"
	"strings"
)

// Get returns the value of the sysctl key specified by name.
func Get(name string) (string, error) {
	// "net.ipv4.tcp_syncookies" translates into "/proc/sys/net/ipv4/tcp_syncookies"
	const sysctlRoot = "/proc/sys"
	relativeSysctlPath := strings.Replace(name, ".", "/", -1)
	sysctlPath := path.Join(sysctlRoot, relativeSysctlPath)

	byteContent, err := ioutil.ReadFile(sysctlPath)
	if err != nil {
		return "", err
	}

	// As the sys file system represent single values as files, they are
  // terminated with a newline. We trim trailing newline, if present.
	content := strings.TrimSuffix(string(byteContent), "\n")

	return content, nil
}
