// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package isolation

import (
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/pivotal-golang/bytefmt"
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
	memorysizeInBytes := int(64 * bytefmt.MEGABYTE)
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

		inputFmt := string(data[:len(data)-1])
		readMemoryInBytes, err := strconv.Atoi(inputFmt)
		So(err, ShouldBeNil)
		So(readMemoryInBytes, ShouldEqual, memorysizeInBytes)
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
