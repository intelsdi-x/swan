// +build integration

package isolation

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"os/user"
	"strconv"
	"testing"
)

func TestCpuSet(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Fatalf("Cannot get current user")
	}

	if user.Name != "root" {
		t.Skipf("Need to be privileged user to run cgroups tests")
	}

	cpuset := CPUSetShares{cgroupName: "M", cpuSetShares: "0-2"}

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err = cmd.Start()

	Convey("While using TestCpu", t, func() {
		So(err, ShouldBeNil)
	})

	Convey("Should provide cpuset Create() to return and correct cpu set shares", t, func() {
		So(cpuset.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpuset/" + cpuset.cgroupName + "/cpuset.cpus")
		So(err, ShouldBeNil)

		inputFmt := data[:len(data)-1]
		So(string(inputFmt), ShouldEqual, cpuset.cpuSetShares)
	})

	Convey("Should provide cpuset Isolate() to return and correct process id", t, func() {
		So(cpuset.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpuset/" + cpuset.cgroupName + "/tasks")
		So(err, ShouldBeNil)

		inputFmt := data[:len(data)-1]
		strPID := strconv.Itoa(cmd.Process.Pid)
		d := []byte(strPID)

		So(string(inputFmt), ShouldContainSubstring, string(d))

	})

	Convey("Should provide Clean() to return", t, func() {
		So(cpuset.Clean(), ShouldBeNil)
	})

	cmd = exec.Command("sh", "-c", "kill -9 ", string(cmd.Process.Pid))

	err = cmd.Start()

	Convey("Should provide kill to return while TestCpuSet", t, func() {
		So(err, ShouldBeNil)
	})

}
