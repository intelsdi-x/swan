package isolation

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
	"strings"
)

// CPUSet describes a cgroup cpuset with core ids and numa (memory) nodes.
type CPUSet struct {
	name string
	cpus Set
	mems Set
}

// NewCPUSet creates an instance of input data.
func NewCPUSet(name string, cpus Set, mems Set) Isolation {
	return &CPUSet{
		name: name,
		cpus: cpus,
		mems: mems,
	}
}

// Decorate implements Decorator interface
func (cpuSet *CPUSet) Decorate(command string) string {
	return "cgexec -g cpuset:" + cpuSet.name + " " + command
}

// Clean removes specified cgroup.
func (cpuSet *CPUSet) Clean() error {
	cmd := exec.Command("cgdelete", "-g", "cpuset:"+cpuSet.name)
	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Create specified cgroup.
func (cpuSet *CPUSet) Create() error {
	// 1.a Create cpuset cgroup.
	cmd := exec.Command("cgcreate", "-g", "cpuset:"+cpuSet.name)
	err := cmd.Run()
	if err != nil {
		return err
	}

	// 1.b Set cpu nodes for cgroup cpus. This is a temporary change.
	// After we discover platform, we change accordingly.
	cpus := []string{}
	for cpu := range cpuSet.cpus {
		cpus = append(cpus, strconv.Itoa(cpu))
	}

	cmd = exec.Command("cgset", "-r", "cpuset.cpus="+strings.Join(cpus, ","), cpuSet.name)
	err = cmd.Run()
	if err != nil {
		return err
	}

	mems := []string{}
	for mem := range cpuSet.mems {
		mems = append(mems, strconv.Itoa(mem))
	}

	cmd = exec.Command("cgset", "-r", "cpuset.mems="+strings.Join(mems, ","), cpuSet.name)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Isolate creates specified cgroup.
func (cpuSet *CPUSet) Isolate(PID int) error {
	// Set PID to cgroups
	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	err := ioutil.WriteFile(path.Join("/sys/fs/cgroup/cpuset", cpuSet.name, "/tasks"), d, 0644)

	if err != nil {
		return err
	}

	return nil
}
