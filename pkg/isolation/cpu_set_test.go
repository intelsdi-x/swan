package isolation

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestCpuSet(t *testing.T) {
	cpuset := CPUSetShares{cgroupName: "M", cpuSetShares: "0-2"}

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	cpuset.Isolate(cmd.Process.Pid)

	fmt.Printf(cpuset.cgroupName)

}
