package isolation

import (
	"fmt"
	"os/exec"
	"testing"
)

func TestCpu(t *testing.T) {
	cpu := CPUShares{cgroupName: "M", cpuShares: "1024"}

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	cpu.Create()

	cpu.Isolate(cmd.Process.Pid)

	fmt.Printf(cpu.cgroupName)

}
