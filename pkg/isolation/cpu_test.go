package isolation

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"strconv"
	"testing"
)

func TestCpu(t *testing.T) {
	cpu := CPUShares{cgroupName: "M", cpuShares: "1024"}

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	fmt.Printf(cpu.cgroupName)

	Convey("Should provide Create() to return and correct cpu shares", t, func() {
		So(cpu.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/" + cpu.cgroupName + "/cpu.shares")
		if err != nil {
			panic(err)
		}
		inputFmt := data[:len(data)-1]
		So(string(inputFmt), ShouldEqual, cpu.cpuShares)
	})

	Convey("Should provide Isolate() to return and correct process id", t, func() {
		So(cpu.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpu/" + cpu.cgroupName + "/tasks")
		if err != nil {
			panic(err)
		}
		inputFmt := data[:len(data)-1]
		strPID := strconv.Itoa(cmd.Process.Pid)
		d := []byte(strPID)

		So(string(inputFmt), ShouldEqual, string(d))

	})

	Convey("Should provide Clean() to return", t, func() {
		So(cpu.Clean(), ShouldBeNil)
	})

}
