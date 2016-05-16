package isolation

import (
	"io/ioutil"
	"os/exec"
	"path"
	"strconv"
)

// MemorySize defines input data
type MemorySize struct {
	name string
	size string
}

// NewMemorySize creates an instance of input data.
func NewMemorySize(name string, size string) Isolation {
	return &MemorySize{name: name, size: size}
}

// Controller returns the cgroup controller name.
func (memorySize *MemorySize) Controller() string {
	return "mem"
}

// Path returns the path relative to the controller root.
func (memorySize *MemorySize) Path() string {
	return memorySize.name
}


// Clean removes specified cgroup.
func (memorySize *MemorySize) Clean() error {
	cmd := exec.Command("cgdelete", "-g", "memory:"+memorySize.name)

	err := cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Create specified cgroup.
func (memorySize *MemorySize) Create() error {

	// 1.a Create memory size cgroup.

	cmd := exec.Command("cgcreate", "-g", "memory:"+memorySize.name)

	err := cmd.Run()
	if err != nil {
		return err
	}

	// 1.b Set cgroup memory size.

	cmd = exec.Command("cgset", "-r", "memory.limit_in_bytes="+memorySize.size, memorySize.name)

	err = cmd.Run()
	if err != nil {
		return err
	}

	return nil
}

// Isolate create specified cgroup and associates specified process id
func (memorySize *MemorySize) Isolate(PID int) error {

	// Set PID to cgroups.
	// cgclassify and cgexec seem to exit with error so temporarily using file io.

	strPID := strconv.Itoa(PID)
	d := []byte(strPID)
	err := ioutil.WriteFile(path.Join("/sys/fs/cgroup/memory", memorySize.name, "tasks"), d, 0644)

	if err != nil {
		return err
	}

	return nil
}
