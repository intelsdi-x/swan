// +build integration

package isolation

import (
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"strconv"
	"testing"
)

func TestCpuSet(t *testing.T) {
	cpuset := CPUSetShares{cgroupName: "M", cpuSetShares: "0-2"}

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err := cmd.Start()
	if err != nil {
		panic(err)
	}

	fmt.Printf(cpuset.cgroupName)

	Convey("Should provide cpuset Create() to return and correct cpu set shares", t, func() {
		So(cpuset.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpuset/" + cpuset.cgroupName + "/cpuset.cpus")
		if err != nil {
			panic(err)
		}
		inputFmt := data[:len(data)-1]
		So(string(inputFmt), ShouldEqual, cpuset.cpuSetShares)
	})

	Convey("Should provide cpuset Isolate() to return and correct process id", t, func() {
		So(cpuset.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpuset/" + cpuset.cgroupName + "/tasks")
		if err != nil {
			panic(err)
		}
		inputFmt := data[:len(data)-1]
		strPID := strconv.Itoa(cmd.Process.Pid)
		d := []byte(strPID)

		So(string(inputFmt), ShouldContainSubstring, string(d))

	})

	Convey("Should provide Clean() to return", t, func() {
		So(cpuset.Clean(), ShouldBeNil)
	})

}
