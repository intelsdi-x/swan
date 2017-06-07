package executor

import (
	"testing"
	"time"

	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
)

const taskExecutionTime = 100 * time.Millisecond
const waitTimeout = 60 * time.Second

func TestSuccsessfulChainedTaskHandle(t *testing.T) {
	Convey("When having ChainedTaskHandle with successful execution", t, func() {
		initialTaskHandle := new(MockTaskHandle)
		initialTaskHandle.On("Wait", 0*time.Second).After(taskExecutionTime).Return(true, nil)
		initialTaskHandle.On("Stop").Return(nil)

		chainedTaskHandle := new(MockTaskHandle)
		chainedTaskHandle.On("Wait", 0*time.Second).After(2*taskExecutionTime).Return(true, nil)
		chainedTaskHandle.On("Stop").Return(nil)

		chainedLauncher := new(MockLauncher)
		chainedLauncher.On("Launch").Return(chainedTaskHandle, nil)

		Convey("Successful run should yield no error", func() {
			startTime := time.Now()
			taskHandle := NewChainedTaskHandle(initialTaskHandle, chainedLauncher)
			So(taskHandle.Status(), ShouldEqual, RUNNING)

			isTerminated, err := taskHandle.Wait(waitTimeout)
			So(time.Since(startTime), ShouldBeGreaterThanOrEqualTo, taskExecutionTime)
			So(err, ShouldBeNil)
			So(isTerminated, ShouldBeTrue)
			So(taskHandle.Status(), ShouldEqual, TERMINATED)

			Convey("Subsequent Waits should yield no error", func() {
				isTerminated, err := taskHandle.Wait(waitTimeout)
				So(err, ShouldBeNil)
				So(isTerminated, ShouldBeTrue)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)

				isTerminated, err = taskHandle.Wait(waitTimeout)
				So(err, ShouldBeNil)
				So(isTerminated, ShouldBeTrue)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})
			Convey("Subsequent Stops should yield no error", func() {
				err := taskHandle.Stop()
				So(err, ShouldBeNil)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)

				err = taskHandle.Stop()
				So(err, ShouldBeNil)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})
		})

		Convey("Immediate stop of ChainedTaskHandle should yield no error", func() {
			taskHandle := NewChainedTaskHandle(initialTaskHandle, chainedLauncher)
			err := taskHandle.Stop()
			So(err, ShouldBeNil)
			So(taskHandle.Status(), ShouldEqual, TERMINATED)

			Convey("Subsequent stop should yield no error", func() {
				err := taskHandle.Stop()
				So(err, ShouldBeNil)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})
		})

		Convey("Stop during execution of chained task should yield no error", func() {
			taskHandle := NewChainedTaskHandle(initialTaskHandle, chainedLauncher)
			taskHandle.Wait(taskExecutionTime + taskExecutionTime/4) // Wait for second task to start.
			startTime := time.Now()
			err := taskHandle.Stop()
			So(err, ShouldBeNil)
			So(time.Since(startTime), ShouldBeLessThan, taskExecutionTime)
			So(taskHandle.Status(), ShouldEqual, TERMINATED)

			Convey("Subsequent stop should yield no error", func() {
				err := taskHandle.Stop()
				So(err, ShouldBeNil)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})
		})
	})
}

func TestFailureChainedTaskHandle(t *testing.T) {
	const fixedError = "initial error"
	Convey("When having ChainedTaskHandle with successful execution", t, func() {
		initialTaskHandle := new(MockTaskHandle)
		chainedTaskHandle := new(MockTaskHandle)
		chainedLauncher := new(MockLauncher)

		Convey("When initial TaskHandle is running and returns error", func() {
			initialTaskHandle.On("Wait", 0*time.Second).After(taskExecutionTime).Return(true, errors.New(fixedError))
			initialTaskHandle.On("Stop").Return(errors.New(fixedError))

			taskHandle := NewChainedTaskHandle(initialTaskHandle, chainedLauncher)
			Convey("It should be returned on Wait()", func() {
				isTerminated, err := taskHandle.Wait(waitTimeout)
				So(isTerminated, ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, fixedError)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)

				Convey("Subsequent Wait() should return the same error", func() {
					isTerminated, err := taskHandle.Wait(waitTimeout)
					So(isTerminated, ShouldBeTrue)
					So(err.Error(), ShouldContainSubstring, fixedError)
					So(taskHandle.Status(), ShouldEqual, TERMINATED)
				})

				Convey("Subsequent Stop() should return the same error", func() {
					err := taskHandle.Stop()
					So(err.Error(), ShouldContainSubstring, fixedError)
					So(taskHandle.Status(), ShouldEqual, TERMINATED)
				})
			})

			Convey("It should be returned on Stop()", func() {
				err := taskHandle.Stop()
				So(err.Error(), ShouldContainSubstring, fixedError)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)

				Convey("Subsequent Wait() should return the same error", func() {
					isTerminated, err := taskHandle.Wait(waitTimeout)
					So(isTerminated, ShouldBeTrue)
					So(err.Error(), ShouldContainSubstring, fixedError)
					So(taskHandle.Status(), ShouldEqual, TERMINATED)
				})

				Convey("Subsequent Stop() should return the same error", func() {
					err := taskHandle.Stop()
					So(err.Error(), ShouldContainSubstring, fixedError)
					So(taskHandle.Status(), ShouldEqual, TERMINATED)
				})
			})
		})

		Convey("When Chained Launcher returns an error", func() {
			initialTaskHandle.On("Wait", 0*time.Second).After(taskExecutionTime).Return(true, nil)
			chainedLauncher.On("Launch").Return(nil, errors.New(fixedError))

			taskHandle := NewChainedTaskHandle(initialTaskHandle, chainedLauncher)
			isTerminated, err := taskHandle.Wait(waitTimeout)
			So(isTerminated, ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, fixedError)
			So(taskHandle.Status(), ShouldEqual, TERMINATED)

			Convey("Subsequent Wait() should return the same error", func() {
				isTerminated, err := taskHandle.Wait(waitTimeout)
				So(isTerminated, ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, fixedError)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})

			Convey("Subsequent Stop() should return the same error", func() {
				err := taskHandle.Stop()
				So(err.Error(), ShouldContainSubstring, fixedError)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})
		})

		Convey("When Chained TaskHandle returns an error", func() {
			initialTaskHandle.On("Wait", 0*time.Second).After(taskExecutionTime).Return(true, nil)
			chainedLauncher.On("Launch").Return(chainedTaskHandle, nil)
			chainedTaskHandle.On("Wait", 0*time.Second).After(taskExecutionTime).Return(true, errors.New(fixedError))
			chainedTaskHandle.On("Stop").Return(errors.New(fixedError))

			taskHandle := NewChainedTaskHandle(initialTaskHandle, chainedLauncher)
			isTerminated, err := taskHandle.Wait(waitTimeout)
			So(isTerminated, ShouldBeTrue)
			So(err.Error(), ShouldContainSubstring, fixedError)
			So(taskHandle.Status(), ShouldEqual, TERMINATED)

			Convey("Subsequent Wait() should return the same error", func() {
				isTerminated, err := taskHandle.Wait(waitTimeout)
				So(isTerminated, ShouldBeTrue)
				So(err.Error(), ShouldContainSubstring, fixedError)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})

			Convey("Subsequent Stop() should return the same error", func() {
				err := taskHandle.Stop()
				So(err.Error(), ShouldContainSubstring, fixedError)
				So(taskHandle.Status(), ShouldEqual, TERMINATED)
			})
		})
	})
}
