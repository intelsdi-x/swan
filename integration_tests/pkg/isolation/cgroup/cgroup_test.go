// +build parallel

package integration

import (
	"os"
	"os/exec"
	pth "path"
	"strings"
	"testing"

	. "github.com/intelsdi-x/swan/pkg/isolation/cgroup"
	. "github.com/smartystreets/goconvey/convey"
)

func uuidgen(t *testing.T) string {
	cmd := exec.Command("uuidgen")
	out, err := cmd.Output()
	if err != nil {
		t.Error(err)
	}
	return strings.TrimSpace(string(out))
}

// AbsPath(controller string) (string, error)
func TestCgroupAbsPath(t *testing.T) {
	Convey("After constructing a cgroup", t, func() {
		cg, _ := NewCgroup([]string{"cpuset"}, "foo")
		Convey("It should have a reasonable absolute path", func() {
			So(cg.AbsPath("cpuset"), ShouldNotBeNil)
		})
	})
}

// Create() error
func TestCgroupCreate(t *testing.T) {
	Convey("After creating a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpu"
		cg, _ := NewCgroup([]string{controller}, path)
		err := cg.Create()
		defer cg.Destroy(true)

		Convey("The returned error should be nil", func() {
			So(err, ShouldBeNil)
		})
		Convey("And the absolute path should exist", func() {
			abs := cg.AbsPath(controller)
			info, err := os.Stat(abs)
			So(err, ShouldBeNil)
			So(info, ShouldNotBeNil)
			So(info.IsDir(), ShouldBeTrue)
		})
	})
}

// Exists() (bool, error)
func TestCgroupExists(t *testing.T) {
	Convey("After constructing a new cgroup", t, func() {
		path := uuidgen(t)
		cg, _ := NewCgroup([]string{"cpu", "cpuset"}, path)

		Convey("It should not exist before being created", func() {
			ok, err := cg.Exists()
			So(err, ShouldBeNil)
			So(ok, ShouldBeFalse)
		})

		cg.Create()
		defer cg.Destroy(true)

		Convey("It should exist after being created", func() {
			ok, err := cg.Exists()
			So(err, ShouldBeNil)
			So(ok, ShouldBeTrue)
		})
		Convey("And the directories should exist for each controller", func() {
			for _, ctrl := range cg.Controllers() {
				abs := cg.AbsPath(ctrl)
				info, err := os.Stat(abs)
				So(err, ShouldBeNil)
				So(info, ShouldNotBeNil)
				So(info.IsDir(), ShouldBeTrue)
			}
		})
	})
}

// Parent() Cgroup
func TestCgroupParent(t *testing.T) {
	Convey("After creating a new nested cgroup", t, func() {
		uuid1 := uuidgen(t)
		uuid2 := uuidgen(t)
		path := pth.Join(uuid1, uuid2)
		controller := "cpu"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Parent().Destroy(true)

		Convey("The nested cgroup should exist", func() {
			ok, _ := cg.Exists()
			So(ok, ShouldBeTrue)
		})
		Convey("The path should be nested correctly", func() {
			So(cg.Path(), ShouldEqual, "/"+uuid1+"/"+uuid2)
		})
		Convey("The parent should also exist", func() {
			So(cg.Parent().Path(), ShouldEqual, "/"+uuid1)
			ok, _ := cg.Parent().Exists()
			So(ok, ShouldBeTrue)
		})
		Convey("And the grand-parent should be the root (and exist)", func() {
			So(cg.Parent().Parent().Path(), ShouldEqual, "/")
			ok, _ := cg.Parent().Parent().Exists()
			So(ok, ShouldBeTrue)
		})
	})
}

// Destroy(recursive bool) error
func TestCgroupDestroy(t *testing.T) {
	Convey("After creating a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpu,cpuset"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Destroy(true)

		Convey("The cgroup should exist until it is destroyed", func() {
			ok, _ := cg.Exists()
			So(ok, ShouldBeTrue)

			cg.Destroy(true)
			ok, _ = cg.Exists()
			So(ok, ShouldBeFalse)
			for _, ctrl := range cg.Controllers() {
				abs := cg.AbsPath(ctrl)
				_, err := os.Stat(abs)
				So(err, ShouldNotBeNil)
			}
		})
	})

	Convey("After creating a new nested cgroup", t, func() {
		uuid1 := uuidgen(t)
		uuid2 := uuidgen(t)
		path := pth.Join(uuid1, uuid2)
		controller := "cpu"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Parent().Destroy(true)

		Convey("The nested cgroup should exist", func() {
			ok, _ := cg.Exists()
			So(ok, ShouldBeTrue)

			Convey("And destroying the parent non-recursively should fail", func() {
				err := cg.Parent().Destroy(false)
				So(err, ShouldNotBeNil)
			})
		})
	})
}

// Get(name string) (string, error)
func TestCgroupGet(t *testing.T) {
	Convey("After creating a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpuset"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Destroy(true)
		ok, _ := cg.Exists()
		So(ok, ShouldBeTrue)

		Convey("The cgroup's attributes should be gettable", func() {
			value, err := cg.Get("cpuset.cpu_exclusive")
			So(err, ShouldBeNil)
			So(value, ShouldEqual, "0")
		})
	})
}

// Set(name string, value string) error
func TestCgroupSet(t *testing.T) {
	Convey("After creating a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpuset"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Destroy(true)
		ok, _ := cg.Exists()
		So(ok, ShouldBeTrue)

		Convey("The cgroup's attributes should be settable", func() {
			// Check initial cpuset mems setting
			value, _ := cg.Get("cpuset.mems")
			So(value, ShouldEqual, "")

			// Set cpuset mems
			err := cg.Set("cpuset.mems", "0")
			So(err, ShouldBeNil)
			value, _ = cg.Get("cpuset.mems")
			So(value, ShouldEqual, "0")

			// Check initial cpuset exclusivity setting
			value, _ = cg.Get("cpuset.cpu_exclusive")
			So(value, ShouldEqual, "0")

			// Set cpuset exclusivity
			err = cg.Set("cpuset.cpu_exclusive", "1")
			So(err, ShouldBeNil)
			value, _ = cg.Get("cpuset.cpu_exclusive")
			So(value, ShouldEqual, "1")

			// Unset cpuset exclusivity
			err = cg.Set("cpuset.cpu_exclusive", "0")
			So(err, ShouldBeNil)
			value, _ = cg.Get("cpuset.cpu_exclusive")
			So(value, ShouldEqual, "0")
		})
	})
}

// Tasks(controller string) ([]uint32, error)
func TestCgroupTasks(t *testing.T) {
	Convey("After creating a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpu"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Destroy(true)
		ok, _ := cg.Exists()
		So(ok, ShouldBeTrue)

		Convey("When sleeping in the cgroup", func() {
			cmd := exec.Command("cgexec", "-g", "cpu:"+cg.Path(), "sleep", "10")
			cmd.Start()
			defer cmd.Process.Kill()
			Convey("The tasks set should contain the sleeping process id", func() {
				tasks, err := cg.Tasks(controller)
				So(err, ShouldBeNil)
				So(len(tasks), ShouldEqual, 1)
				So(tasks.Contains(cmd.Process.Pid), ShouldBeTrue)
			})
		})
	})
}

// Decorate(command string) string
func TestCgroupDecorate(t *testing.T) {
	Convey("After constructing a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpu"
		cg, _ := NewCgroup([]string{controller}, path)

		Convey("It should decorate commands", func() {
			result := cg.Decorate("sleep 10")
			So(result, ShouldEqual, "cgexec -g cpu:"+cg.Path()+" sleep 10")
		})
	})
}

// Isolate(PID int) error
func TestCgroupIsolate(t *testing.T) {
	Convey("After creating a new cgroup", t, func() {
		path := uuidgen(t)
		controller := "cpu"
		cg, _ := NewCgroup([]string{controller}, path)
		cg.Create()
		defer cg.Destroy(true)
		ok, _ := cg.Exists()
		So(ok, ShouldBeTrue)

		Convey("When sleeping outside of the cgroup", func() {
			cmd := exec.Command("sleep", "10")
			cmd.Start()
			defer cmd.Process.Kill()
			tasks, _ := cg.Tasks(controller)
			So(tasks.Contains(cmd.Process.Pid), ShouldBeFalse)

			Convey("After isolating the sleeping process", func() {
				err := cg.Isolate(cmd.Process.Pid)
				So(err, ShouldBeNil)

				Convey("The tasks set should contain the sleeping process id", func() {
					tasks, _ := cg.Tasks(controller)
					So(tasks.Contains(cmd.Process.Pid), ShouldBeTrue)
				})
			})
		})
	})
}
