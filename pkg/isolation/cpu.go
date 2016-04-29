package isolation

import "os/exec"
import "io/ioutil"
import "strconv"

// CPUShares defines data needed for CPU controller
type CPUShares struct {
	cgroupName string
	cpuShares  string
}

// NewCPUShares instance creation
func NewCPUShares(nameOfTheCgroup string, NumShares string) *CPUShares {
	return &CPUShares{cgroupName: nameOfTheCgroup, cpuShares: NumShares}
}

// Clean removes the specified cgroup
func (cpu *CPUShares) Clean() error {

	cmd := exec.Command("sh", "-c", "cgdelete -g cpu"+":"+cpu.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Create specified cgroup
func (cpu *CPUShares) Create() error {

	// 1 Create cpu cgroup

	cmd := exec.Command("sh", "-c", "cgcreate -g "+"cpu"+":"+cpu.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// 2 Set cpu cgroup shares

	cmd = exec.Command("sh", "-c", "cgset -r cpu.shares="+cpu.cpuShares+" "+cpu.cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Isolate associates specified pid to the cgroup
func (cpu *CPUShares) Isolate(PID int) error {

	// Associate task with the specified cgroup
	// cgclassify and cgexec seem to exit with error so temporarily using file io

	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	err := ioutil.WriteFile("/sys/fs/cgroup/"+"cpu/"+cpu.cgroupName+"/tasks", d, 0644)

	if err != nil {
		return err
	}

	return nil
}
