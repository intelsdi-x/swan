package isolation

import "os/exec"
import "io/ioutil"
import "strconv"

// MemorySize defines input data
type MemorySize struct {
	cgroupName string
	memorySize string
}

// NewMemorySize creates an instance of input data.
func NewMemorySize(nameOfTheCgroup string, memorySizeShares string) *MemorySize {
	return &MemorySize{cgroupName: nameOfTheCgroup, memorySize: memorySizeShares}
}

// Clean removes specified cgroup.
func (memorysize *MemorySize) Clean() error {

	cmd := exec.Command("sh", "-c", "cgdelete -g memory:"+memorysize.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Create specified cgroup.
func (memorysize *MemorySize) Create() error {

	// 1.a Create memory size cgroup.

	cmd := exec.Command("sh", "-c", "cgcreate -g memory:"+memorysize.cgroupName)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// 1.b Set cgroup memory size.

	cmd = exec.Command("sh", "-c", "cgset -r memory.limit_in_bytes="+memorysize.memorySize+" "+memorysize.cgroupName)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Isolate create specified cgroup and associates specified process id
func (memorysize *MemorySize) Isolate(PID int) error {

	// Set PID to cgroups.
	// cgclassify and cgexec seem to exit with error so temporarily using file io.

	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	err := ioutil.WriteFile("/sys/fs/cgroup/memory/"+memorysize.cgroupName+"/tasks", d, 0644)

	if err != nil {
		return err
	}

	return nil
}
