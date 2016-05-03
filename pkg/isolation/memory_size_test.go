package isolation

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"strconv"
	"testing"
)

func TestMemorySize(t *testing.T) {
	memorysize := MemorySize{cgroupName: "M", memorySize: "536870912"}

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}
	memorysize.Create()

	memorysize.Isolate(cmd.Process.Pid)

	fmt.Printf(memorysize.cgroupName)

	Convey("Should provide memorysize Create() to return and correct memory size", t, func() {
		So(memorysize.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/memory/" + memorysize.cgroupName + "/memory.limit_in_bytes")
		if err != nil {
			panic(err)
		}
		inputFmt := data[:len(data)-1]
		So(string(inputFmt), ShouldEqual, memorysize.memorySize)
	})

	Convey("Should provide memorysize Isolate() to return and correct process id", t, func() {
		So(memorysize.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/memory/" + memorysize.cgroupName + "/tasks")
		if err != nil {
			panic(err)
		}
		inputFmt := data[:len(data)-1]
		strPID := strconv.Itoa(cmd.Process.Pid)
		d := []byte(strPID)

		So(string(inputFmt), ShouldContainSubstring, string(d))

	})

	Convey("Should provide Clean() to return", t, func() {
		So(memorysize.Clean(), ShouldBeNil)
	})

}
