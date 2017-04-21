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
	. "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os/exec"
	"os/user"
	"path"
	"strconv"
	"testing"
)

func TestCpu(t *testing.T) {
	user, err := user.Current()
	if err != nil {
		t.Fatalf("Cannot get current user")
	}

	if user.Name != "root" {
		t.Skipf("Need to be privileged user to run cgroups tests")
	}

	cpuName := "M"
	shares := 1024
	cpu := isolation.NewCPUShares(cpuName, shares)
	cmd := exec.Command("sleep", "1h")

	err = cmd.Start()

	Convey("While using TestCpu", t, func() {
		So(err, ShouldBeNil)
	})

	defer func() {
		err = cmd.Process.Kill()
		Convey("Should provide kill to return while TestCpu", t, func() {
			So(err, ShouldBeNil)
		})
	}()

	Convey("Should provide Create() to return and correct cpu shares", t, func() {
		So(cpu.Create(), ShouldBeNil)
		data, err := ioutil.ReadFile(path.Join("/sys/fs/cgroup/cpu", cpuName, "cpu.shares"))
		So(err, ShouldBeNil)

		inputFmt := string(data[:len(data)-1])
		readShares, err := strconv.Atoi(inputFmt)
		So(err, ShouldBeNil)
		So(readShares, ShouldEqual, shares)
	})

	Convey("Should provide Isolate() to return and correct process id", t, func() {
		So(cpu.Isolate(cmd.Process.Pid), ShouldBeNil)
		data, err := ioutil.ReadFile(path.Join("/sys/fs/cgroup/cpu", cpuName, "tasks"))
		So(err, ShouldBeNil)

		inputFmt := string(data[:len(data)-1])
		strPID := strconv.Itoa(cmd.Process.Pid)

		So(inputFmt, ShouldEqual, strPID)
	})

	Convey("Should provide Clean() to return", t, func() {
		So(cpu.Clean(), ShouldBeNil)
	})
}
