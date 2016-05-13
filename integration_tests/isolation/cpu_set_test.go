// +build integration

package integration

import (
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"os/user"
	"strconv"
	"testing"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

func TestCpuSet(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Fatalf("Cannot get current user")
	}

	if user.Name != "root" {
		t.Skipf("Need to be privileged user to run cgroups tests")
	}

	const cgroupName = "M"
	const cpuSetShares = "0-2"
	const cgCPUNodes = "0"

	cpuset := isolation.NewCPUSetShares(cgroupName, cpuSetShares, cgCPUNodes)

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err = cmd.Start()

	Convey("While using TestCpuSet", t, func() {
		So(err, ShouldBeNil)
	})

	defer func() {
		err = cmd.Process.Kill()
		Convey("Should provide kill to return while  TestCpuSet", t, func() {
			So(err, ShouldBeNil)
		})
	}()

	Convey("Should provide cpuset Create() to return and correct cpu set shares", t, func() {
		So(cpuset.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpuset/" + cgroupName + "/cpuset.cpus")
		So(err, ShouldBeNil)

		inputFmt := data[:len(data)-1]
		So(string(inputFmt), ShouldEqual, cpuSetShares)
	})

	Convey("Should provide cpuset Isolate() to return and correct process id", t, func() {
		So(cpuset.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile("/sys/fs/cgroup/cpuset/" + cgroupName + "/tasks")
		So(err, ShouldBeNil)

		inputFmt := data[:len(data)-1]
		strPID := strconv.Itoa(cmd.Process.Pid)
		d := []byte(strPID)

		So(string(inputFmt), ShouldContainSubstring, string(d))

	})

	Convey("Should provide Clean() to return", t, func() {
		So(cpuset.Clean(), ShouldBeNil)
	})
}
