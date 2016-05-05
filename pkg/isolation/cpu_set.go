package isolation

import "os/exec"
import "io/ioutil"
import "strconv"

// CPUSetShares input definition.
type CPUSetShares struct {
	cgroupName   string
	cpuSetShares string
}

// NewCPUSetShares creates an instance of input data.
func NewCPUSetShares(nameOfTheCgroup string, cpuSets string) *CPUSetShares {
	return &CPUSetShares{cgroupName: nameOfTheCgroup, cpuSetShares: cpuSets}
}

// Clean removes specified cgroup.
func (cpuSet *CPUSetShares) Clean() error {

	cmd := exec.Command("sh", "-c", "cgdelete -g cpuset:"+cpuSet.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Create specified cgroup.
func (cpuSet *CPUSetShares) Create() error {

	// 1.a Create cpuset cgroup.

	cmd := exec.Command("sh", "-c", "cgcreate -g cpuset:"+cpuSet.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// 1.b Set cpu nodes for cgroup cpus. This is a temporary change. After we discover platform, we change accordingly.

	cgCPUNodes := "0-1"

	cmd = exec.Command("sh", "-c", "cgset -r cpuset.mems="+cgCPUNodes+" "+cpuSet.cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

	// 1.c Set cpuset cgroup cpus.

	cmd = exec.Command("sh", "-c", "cgset -r cpuset.cpus="+cpuSet.cpuSetShares+" "+cpuSet.cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Isolate creates specified cgroup.
func (cpuSet *CPUSetShares) Isolate(PID int) error {

	// Set PID to cgroups
	// cgclassify & cgexec seem to exit with error so temporarily using file io

	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	err := ioutil.WriteFile("/sys/fs/cgroup/cpuset"+"/"+cpuSet.cgroupName+"/tasks", d, 0644)

	if err != nil {
		return err
	}

	return nil
}
