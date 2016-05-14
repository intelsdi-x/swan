package isolation

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
)

// CPUShares defines data needed for CPU controller.
type CPUShares struct {
	name   string
	shares int
}

// NewShares instance creation.
func NewShares(name string, shares int) *CPUShares {
	return &CPUShares{name: name, shares: shares}
}

// Controller returns the cgroup controller name.
func (cpu *CPUShares) Controller() string {
	return "cpu"
}

// Path returns the path relative to the controller root.
func (cpu *CPUShares) Path() string {
	return cpu.name
}



// Clean removes the specified cgroup
func (cpu *CPUShares) Clean() error {

	cmd := exec.Command("sh", "-c", "cgdelete -g cpu"+":"+cpu.name)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Create specified cgroup.
func (cpu *CPUShares) Create() error {

	// 1 Create cpu cgroup

	cmd := exec.Command("cgcreate", "-g", "cpu:"+cpu.name)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// 2 Set cpu cgroup shares

	cmd = exec.Command("cgset", "-r", "cpu.shares="+strconv.Itoa(cpu.shares), cpu.name)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Isolate associates specified pid to the cgroup.
func (cpu *CPUShares) Isolate(PID int) error {

	// Associate task with the specified cgroup.
	// cgclassify and cgexec seem to exit with error so temporarily using file io.

	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	err := ioutil.WriteFile(path.Join("/sys/fs/cgroup/cpu", cpu.name, "tasks"), d, 0644)

	if err != nil {
		return err
	}

	return nil
}
