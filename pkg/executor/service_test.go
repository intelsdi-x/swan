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

package executor

import (
	"errors"
	"io/ioutil"
	"os"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

var errStopFailed = errors.New("Stopping failed")

type stoppedTaskHandle struct {
	TaskHandle
	output *os.File
}

// Status implements TaskHandle interface,
func (th stoppedTaskHandle) Status() TaskState {
	return TERMINATED
}

func (th stoppedTaskHandle) Stop() error {
	return nil
}

// Wait implements TaskHandle interface.
func (th stoppedTaskHandle) Wait(duration time.Duration) (bool, error) {
	return true, nil
}

func (th stoppedTaskHandle) StderrFile() (*os.File, error) {
	return th.output, nil
}

func (th stoppedTaskHandle) StdoutFile() (*os.File, error) {
	return th.output, nil
}

func (th stoppedTaskHandle) String() string {
	return "command"
}

type runningTaskHandle struct {
	TaskHandle
}

// Status implements TaskHandle interface,
func (th runningTaskHandle) Status() TaskState {
	return RUNNING
}

// Stop implements TaskHandle interface.
func (th runningTaskHandle) Stop() error {
	return nil
}

// Wait implements TaskHandle interface.
func (th runningTaskHandle) Wait(duration time.Duration) (bool, error) {
	return true, nil
}

func (th runningTaskHandle) String() string {
	return "command"
}

type erroneousTaskHandle struct {
	TaskHandle
}

// Stop implements TaskHandle interface.
func (th erroneousTaskHandle) Stop() error {
	return errStopFailed
}

// Status implements TaskHandle interface,
func (th erroneousTaskHandle) Status() TaskState {
	return RUNNING
}

func (th erroneousTaskHandle) String() string {
	return "command"
}

func TestServiceTaskHandle(t *testing.T) {

	Convey("Calling Stop() on terminated task should fail", t, func() {
		output, err := ioutil.TempFile(os.TempDir(), "serviceTests")
		So(err, ShouldBeNil)
		Reset(func() {
			os.Remove(output.Name())
		})

		s := NewServiceHandle(stoppedTaskHandle{output: output})

		err = s.Stop()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, s.String())

		Convey("Another Stop() should return the same error", func() {
			err = s.Stop()
			So(err, ShouldNotBeNil)
			So(err.Error(), ShouldContainSubstring, s.String())
		})
	})

	Convey("Calling Stop() after Wait() should return same error", t, func() {
		output, err := ioutil.TempFile(os.TempDir(), "serviceTests")
		So(err, ShouldBeNil)
		Reset(func() {
			os.Remove(output.Name())
		})

		s := NewServiceHandle(stoppedTaskHandle{output: output})

		terminated, err := s.Wait(0)
		So(terminated, ShouldBeTrue)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, s.String())

		err = s.Stop()
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, s.String())
	})

	Convey("Calling Stop() on running task should succeed", t, func() {
		s := NewServiceHandle(runningTaskHandle{})

		err := s.Stop()
		So(err, ShouldBeNil)

		Convey("Calling Stop() on one more time should succeed", func() {
			err := s.Stop()
			So(err, ShouldBeNil)
		})
	})

	Convey("Calling Wait() on running task should succeed", t, func() {
		s := NewServiceHandle(runningTaskHandle{})

		status, _ := s.Wait(0)
		So(status, ShouldBeTrue)

		Convey("Calling Wait() on one more time should succeed", func() {
			err := s.Stop()
			So(err, ShouldBeNil)
		})
	})

	Convey("Calling Stop() on running task should fail if embedded TaskHandle.Stop() fails", t, func() {
		s := NewServiceHandle(erroneousTaskHandle{})

		err := s.Stop()
		So(err, ShouldEqual, errStopFailed)
	})
}

var errLaunchFailed = errors.New("Where the senses fail us, reason must step in")

type successfulLauncher struct {
	Launcher
}

// Launch implements Launcher interface.
func (sl successfulLauncher) Launch() (TaskHandle, error) {
	return runningTaskHandle{}, nil
}

// String implements Launcher interface.
func (sl successfulLauncher) String() string {
	return "Underlying name"
}

type failedLauncher struct {
	Launcher
}

// Launch implements Launcher interface.
func (fl failedLauncher) Launch() (TaskHandle, error) {
	return nil, errLaunchFailed
}

func TestServiceLauncher(t *testing.T) {
	Convey("When call to embedded Launcher.Launch() succeeds then running EndlessTaskHandle should be returned", t, func() {
		l := ServiceLauncher{successfulLauncher{}}

		th, err := l.Launch()
		So(err, ShouldBeNil)
		So(th.Stop(), ShouldBeNil)
	})

	Convey("When call to embedded Launcher.Launch() fails then EndlessLauncher.Launch() should fail too", t, func() {
		l := ServiceLauncher{failedLauncher{}}

		th, err := l.Launch()
		So(th, ShouldBeNil)
		So(err, ShouldEqual, errLaunchFailed)
	})

	Convey("Launcher name should contain of embedded Launcher name so that it is transparent", t, func() {
		l := ServiceLauncher{successfulLauncher{}}

		name := l.String()
		So(name, ShouldEqual, "Underlying name")
	})
}
