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

// Wait implements TaskHandle interface.
func (th stoppedTaskHandle) Wait(duration time.Duration) bool {
	return false
}

func (th stoppedTaskHandle) StderrFile() (*os.File, error) {
	return th.output, nil
}

func (th stoppedTaskHandle) StdoutFile() (*os.File, error) {
	return th.output, nil
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
func (th runningTaskHandle) Wait(duration time.Duration) bool {
	return true
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

func TestServiceTaskHandle(t *testing.T) {

	Convey("Calling Stop() or Wait() on terminated task should fail", t, func() {
		output, err := ioutil.TempFile(os.TempDir(), "serviceTests")
		So(err, ShouldBeNil)
		Reset(func() {
			os.Remove(output.Name())
		})

		s := service{stoppedTaskHandle{output: output}}

		success := s.Wait(0)
		So(success, ShouldBeFalse)
		err = s.Stop()
		So(err, ShouldEqual, ErrServiceStopped)
	})

	Convey("Calling Stop() on running task should succeed", t, func() {
		s := service{runningTaskHandle{}}

		err := s.Stop()
		So(err, ShouldBeNil)
	})

	Convey("Calling Wait() on running task should succeed", t, func() {
		s := service{runningTaskHandle{}}

		status := s.Wait(0)
		So(status, ShouldBeTrue)
	})

	Convey("Calling Stop() on running task should fail if embedded TaskHandle.Stop() fails", t, func() {
		s := service{erroneousTaskHandle{}}

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

// Name implements Launcher interface.
func (sl successfulLauncher) Name() string {
	return "My Name Is"
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

	Convey("Launcher name should contain of embedded Launcher name and indicate that EndlessLauncher is used", t, func() {
		l := ServiceLauncher{successfulLauncher{}}

		name := l.Name()
		So(name, ShouldContainSubstring, "My Name Is")
		So(name, ShouldContainSubstring, "Service")
	})
}
