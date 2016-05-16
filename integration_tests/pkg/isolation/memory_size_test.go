package isolation

import (
	"github.com/intelsdi-x/swan/pkg/isolation"
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"testing"
)

func TestMemorySize(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Fatalf("Cannot get current user")
	}

	if user.Name != "root" {
		t.Skipf("Need to be privileged user to run cgroups tests")
	}

	memoryName := "M"
	memorysizeInBytes := "536870912"
	memorysize := isolation.NewMemorySize(memoryName, memorysizeInBytes)

	cmd := exec.Command("sh", "-c", "sleep 1h")
	err = cmd.Start()

	Convey("While using TestCpu", t, func() {
		So(err, ShouldBeNil)
	})

	defer func() {
		err = cmd.Process.Kill()
		Convey("Should provide kill to return while  TestMemorySize", t, func() {
			So(err, ShouldBeNil)
		})
	}()

	Convey("Should provide memorysize Create() to return and correct memory size", t, func() {
		So(memorysize.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile(path.Join("/sys/fs/cgroup/memory", memoryName, "memory.limit_in_bytes"))

		So(err, ShouldBeNil)

		inputFmt := data[:len(data)-1]
		So(string(inputFmt), ShouldEqual, memorysizeInBytes)
	})

	Convey("Should provide memorysize Isolate() to return and correct process id", t, func() {
		So(memorysize.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile(path.Join("/sys/fs/cgroup/memory", memoryName, "tasks"))

		So(err, ShouldBeNil)

		inputFmt := data[:len(data)-1]
		strPID := strconv.Itoa(cmd.Process.Pid)
		d := []byte(strPID)

		So(string(inputFmt), ShouldContainSubstring, string(d))

	})

	Convey("Should provide Clean() to return", t, func() {
		So(memorysize.Clean(), ShouldBeNil)
	})
}
