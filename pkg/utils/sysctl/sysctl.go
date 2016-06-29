package sysctl

import (
	"fmt"
	"io/ioutil"
	"runtime"
	"strings"
)

const (
	sysctlDir = "/proc/sys/"
)

// Get returns the value from a sysctl key with the specified name.
func Get(name string) (string, error) {
	if runtime.GOOS != "linux" {
		return "", fmt.Errorf("sysctl only supported on linux: '%s' found", runtime.GOOS)
	}
	path := sysctlDir + strings.Replace(name, ".", "/", -1)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("Could not find key with name '%s'", name)
	}
	return strings.TrimSpace(string(data)), nil
}

// Set the value from a sysctl key with the specified name.
func Set(name string, value string) error {
	if runtime.GOOS != "linux" {
		return fmt.Errorf("sysctl only supported on linux: %s found", runtime.GOOS)
	}
	path := sysctlDir + strings.Replace(name, ".", "/", -1)
	err := ioutil.WriteFile(path, []byte(value), 0644)
	return err
}
