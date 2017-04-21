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
	"bufio"
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"regexp"

	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

const (
	// CPUModelNameKey defines a key in the platform metrics map
	CPUModelNameKey = "cpu_model"
	// KernelVersionKey defines a key in the platform metrics map
	KernelVersionKey = "kernel_version"
	// CentOSVersionKey defines a key in the platform metrics map
	CentOSVersionKey = "centos_version"
	// CPUTopologyKey defines a key in the platform metrics map
	CPUTopologyKey = "cpu_topology"
	// DockerVersionKey defines a key in the platform metrics map
	DockerVersionKey = "docker_version"
	// SnapteldVersionKey defines a key in the platform metrics map
	SnapteldVersionKey = "snapteld_version"
	// PowerGovernorKey defines a key in the platform metrics map
	PowerGovernorKey = "power_governor"
	// IRQAffinityKey defines a key in the platform metrics map
	IRQAffinityKey = "irq_affinity"
	// EtcdVersionKey defines a key in the platform metrics map
	EtcdVersionKey = "etcd_version"
)

// GetPlatformMetrics returns map of strings with platform metrics.
// If metric could not be retreived value for the key is empty string.
func GetPlatformMetrics() (platformMetrics map[string]string) {
	platformMetrics = make(map[string]string)
	item, err := CPUModelName()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", CPUModelNameKey, err.Error()))
	}
	platformMetrics[CPUModelNameKey] = item

	item, err = KernelVersion()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", KernelVersionKey, err.Error()))
	}
	platformMetrics[KernelVersionKey] = item

	item, err = CentOSVersion()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", CentOSVersionKey, err.Error()))
	}
	platformMetrics[CentOSVersionKey] = item

	item, err = CPUTopology()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", CPUTopologyKey, err.Error()))
	}
	platformMetrics[CPUTopologyKey] = item

	item, err = DockerVersion()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", DockerVersionKey, err.Error()))
	}
	platformMetrics[DockerVersionKey] = item

	item, err = SnapteldVersion()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", SnapteldVersionKey, err.Error()))
	}
	platformMetrics[SnapteldVersionKey] = item

	item, err = PowerGovernor()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", PowerGovernorKey, err.Error()))
	}
	platformMetrics[PowerGovernorKey] = item

	item, err = IRQAffinity()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", IRQAffinityKey, err.Error()))
	}
	platformMetrics[IRQAffinityKey] = item

	item, err = EtcdVersion()
	if err != nil {
		logrus.Warn(fmt.Sprintf("GetPlatformMetrics: Failed to get %s metric. Skipping. Error: %s", EtcdVersionKey, err.Error()))
	}
	platformMetrics[EtcdVersionKey] = item

	return platformMetrics
}

// CPUModelName reads /proc/cpuinfo and returns line 'model name' line.
// Note that it returns only first occurrence of the model since mixed cpu models
// in > 2 CPUs are not supported
// In case of an error empty string is returned.
func CPUModelName() (string, error) {
	file, err := os.Open("/proc/cpuinfo")
	if err != nil {
		return "", errors.Wrapf(err, "Cannot open /proc/cpuinfo file.")
	}
	defer file.Close()

	procScanner := bufio.NewScanner(file)

	for procScanner.Scan() {
		line := procScanner.Text()
		chunks := strings.Split(line, ":")
		key := strings.TrimSpace(chunks[0])
		value := strings.TrimSpace(chunks[1])
		if key == "model name" {
			return value, nil
		}
	}
	// Return error from scanner or newly created one.
	err = procScanner.Err()
	if err == nil {
		err = errors.New("Did not find phrase 'model name' in /proc/cpuinfo")
	}
	return "", err
}

// KernelVersion return kernel version as stated in /proc/version
// In case of an error empty string is returned
func KernelVersion() (string, error) {
	return readContents("/proc/version")
}

// CentOSVersion returns OS version as stated in /etc/redhat-release
// In case of an error empty string is returned
func CentOSVersion() (string, error) {
	return readContents("/etc/redhat-release")
}

// CPUTopology returns CPU topology returned by 'lscpu -e' command.
// The whole output of the command is returned.
// In case of an error empty string is returned
func CPUTopology() (string, error) {
	cmd := exec.Command("lscpu", "-e")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get output from lscpu -e")
	}
	return strings.TrimSpace(string(output)), nil
}

// DockerVersion returns docker version as returned by 'docker version' command.
// In case of an error empty string is returned
func DockerVersion() (string, error) {
	cmd := exec.Command("docker", "version")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get output from docker version")
	}
	return strings.TrimSpace(string(output)), nil
}

// SnapteldVersion returns snapteld version as returned by 'snapteld -v' command.
// In case of an error empty string is returned
func SnapteldVersion() (string, error) {
	cmd := exec.Command("snapteld", "-v")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get output from snapteld version")
	}
	return strings.TrimSpace(string(output)), nil

}

// PowerGovernor returns a comma separated list of CPU:power_policy.
// Example (snippet):
//    "performance,1:performance,10:performance,11:performance"
// In case of an error empty string is returned
func PowerGovernor() (string, error) {
	dir := "/sys/devices/system/cpu"
	files, err := ioutil.ReadDir(dir)

	if err != nil {
		return "", errors.Wrapf(err, "Failed to scan sysfs for CPU devices")
	}

	re := regexp.MustCompile("cpu[0-9]+")
	output := []string{}
	for _, file := range files {
		if file.IsDir() && re.MatchString(file.Name()) {
			cpufreq := path.Join(dir, file.Name(), "cpufreq/scaling_governor")

			// Just try to read it. Don't try to be smart here. Failure is OK.
			gov, err := readContents(cpufreq)
			if err != nil {
				return "", err
			}
			item := fmt.Sprintf("%s:%s", strings.TrimLeft(file.Name(), "cpu"), gov)
			output = append(output, item)
		}
	}

	return strings.Join(output, ","), nil
}

// IRQAffinity returns semicolon (;) separated list of pairs iface {comma separated
// list of pairs queue:affinity}
// Example:
//   enp0s31f6 {134:6};enp3s0 {129:5,130:4,131:3,132:2,133:2}
// In case of an error empty string is returned
func IRQAffinity() (string, error) {
	dir := "/sys/class/net"
	ifaces, err := ioutil.ReadDir(dir)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to scan sysfs for ifaces")
	}
	output := []string{}
	// Enumerate all network interfaces in the OS
	for _, iface := range ifaces {
		// for each network interface check for 'device' directory.
		// Note: local interfaces (lo) doesn't have this so on err is to skip.
		device := path.Join(dir, iface.Name(), "device/msi_irqs")
		info, err := os.Stat(device)
		if err == nil && info.IsDir() {
			queues, err := getIfaceQueuesAffinity(iface.Name())
			if err != nil {
				return "", errors.Wrapf(err, "Failed to get %s queues affinity", iface.Name())
			}
			item := fmt.Sprintf("%s {%s}", iface.Name(), queues)
			output = append(output, item)
		}
	}
	return strings.Join(output, ";"), nil
}

// EtcdVersion returns etcd version as returned by 'etcd --version'.
// In case of an error empty string is returned
func EtcdVersion() (string, error) {
	cmd := exec.Command("etcd", "--version")
	output, err := cmd.Output()
	if err != nil {
		return "", errors.Wrapf(err, "Failed to get etcd version")
	}
	return strings.TrimSpace(string(output)), nil
}

func readContents(name string) (string, error) {
	content, err := ioutil.ReadFile(name)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to read %s", name)
	}
	return strings.TrimSpace(string(content)), nil
}

func getIfaceQueuesAffinity(iface string) (string, error) {
	ifaceDir := path.Join("/sys/class/net", iface, "device/msi_irqs")
	queues, err := ioutil.ReadDir(ifaceDir)
	if err != nil {
		return "", errors.Wrapf(err, "Failed to read %s directory", ifaceDir)
	}
	output := []string{}

	for _, queue := range queues {
		smpAffinityFile := path.Join("/proc/irq", queue.Name(), "smp_affinity_list")
		smpAffinity, err := readContents(smpAffinityFile)
		if err != nil {
			return "", err
		}
		item := fmt.Sprintf("%s:%s", queue.Name(), smpAffinity)
		output = append(output, item)
	}
	return strings.Join(output, ","), nil
}
