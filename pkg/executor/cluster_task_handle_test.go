package executor_test

import (
	"errors"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

const (
	testMasterAddress = "testMasterAddress"
)

func TestClusterTaskHandle(t *testing.T) {
	Convey("While using MasterAgentTaskHandle", t, func() {
		Convey("When we have a master task handle", func() {
			mMasterHandle := new(mocks.TaskHandle)

			// ClusterTaskHandle with only one master should behave the same as master handle itself.
			Convey("And create ClusterTaskHandle from it", func() {
				handle := executor.NewClusterTaskHandle(
					mMasterHandle, []executor.TaskHandle{})
				Convey("It should implement TaskHandle", func() {
					So(handle, ShouldImplement, (*executor.TaskHandle)(nil))
				})

				Convey("StdoutFile() should return master StdoutFile result.", func() {
					mMasterHandle.On("StdoutFile").Return(nil, nil).Once()
					_, err := handle.StdoutFile()
					So(err, ShouldBeNil)

					mMasterHandle.On("StdoutFile").Return(nil, errors.New("test")).Once()
					_, err = handle.StdoutFile()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")
				})

				Convey("StderrFile() should return master StderrFile result.", func() {
					mMasterHandle.On("StderrFile").Return(nil, nil).Once()
					_, err := handle.StderrFile()
					So(err, ShouldBeNil)

					mMasterHandle.On("StderrFile").Return(nil, errors.New("test")).Once()
					_, err = handle.StderrFile()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")
				})

				Convey("Stop() should stop master", func() {
					mMasterHandle.On("Stop").Return(nil).Once()
					So(handle.Stop(), ShouldBeNil)

					mMasterHandle.On("Stop").Return(errors.New("test")).Once()
					err := handle.Stop()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")
				})

				Convey("Status() should return status of the master", func() {
					mMasterHandle.On("Status").Return(executor.RUNNING).Once()
					So(handle.Status(), ShouldEqual, executor.RUNNING)
				})

				Convey("ExitCode() should return exit code of the master", func() {
					mMasterHandle.On("ExitCode").Return(1234, nil).Once()
					code, err := handle.ExitCode()
					So(code, ShouldEqual, 1234)
					So(err, ShouldBeNil)

					mMasterHandle.On("ExitCode").Return(0, errors.New("test")).Once()
					code, err = handle.ExitCode()
					So(code, ShouldEqual, 0)
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")
				})

				Convey("Wait() should wait for master", func() {
					mMasterHandle.On("Wait", 0*time.Microsecond).Return(true).Once()
					So(handle.Wait(0), ShouldEqual, true)

					mMasterHandle.On("Wait", 0*time.Microsecond).Return(false).Once()
					So(handle.Wait(0), ShouldEqual, false)
				})

				Convey("Clean() should clean master's resources", func() {
					mMasterHandle.On("Clean").Return(nil).Once()
					So(handle.Clean(), ShouldBeNil)

					mMasterHandle.On("Clean").Return(errors.New("test")).Once()
					err := handle.Clean()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")
				})

				Convey("EraseOutput() should erase master's output", func() {
					mMasterHandle.On("EraseOutput").Return(nil).Once()
					So(handle.EraseOutput(), ShouldBeNil)

					mMasterHandle.On("EraseOutput").Return(errors.New("test")).Once()
					err := handle.EraseOutput()
					So(err, ShouldNotBeNil)
					So(err.Error(), ShouldEqual, "test")
				})

				Convey("Address() should return master address.", func() {
					mMasterHandle.On("Address").Return(testMasterAddress).Once()
					So(handle.Address(), ShouldEqual, testMasterAddress)
				})
			})

			// Testing multi agent mode.
			Convey("And have additional 2 agents", func() {
				mAgent1Handle := new(mocks.TaskHandle)
				mAgent2Handle := new(mocks.TaskHandle)

				Convey("And create ClusterTaskHandle from it", func() {
					handle := executor.NewClusterTaskHandle(
						mMasterHandle, []executor.TaskHandle{
							mAgent1Handle, mAgent2Handle,
						},
					)

					Convey("It should implement TaskHandle", func() {
						So(handle, ShouldImplement, (*executor.TaskHandle)(nil))
					})

					Convey("StdoutFile() should return master StdoutFile result.", func() {
						mMasterHandle.On("StdoutFile").Return(nil, nil).Once()
						_, err := handle.StdoutFile()
						So(err, ShouldBeNil)

						mMasterHandle.On("StdoutFile").Return(nil, errors.New("test")).Once()
						_, err = handle.StdoutFile()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")
					})

					Convey("StderrFile() should return master StderrFile result.", func() {
						mMasterHandle.On("StderrFile").Return(nil, nil).Once()
						_, err := handle.StderrFile()
						So(err, ShouldBeNil)

						mMasterHandle.On("StderrFile").Return(nil, errors.New("test")).Once()
						_, err = handle.StderrFile()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")
					})

					Convey("Stop() should stop master and agents", func() {
						mMasterHandle.On("Stop").Return(nil).Once()
						mAgent1Handle.On("Stop").Return(nil).Once()
						mAgent2Handle.On("Stop").Return(nil).Once()

						So(handle.Stop(), ShouldBeNil)

						mMasterHandle.On("Stop").Return(errors.New("test")).Once()
						mAgent1Handle.On("Stop").Return(nil).Once()
						mAgent2Handle.On("Stop").Return(nil).Once()
						err := handle.Stop()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")

						mMasterHandle.On("Stop").Return(errors.New("test")).Once()
						mAgent1Handle.On("Stop").Return(nil).Once()
						mAgent2Handle.On("Stop").Return(errors.New("test2")).Once()
						err = handle.Stop()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test; test2")

						mMasterHandle.On("Stop").Return(nil).Once()
						mAgent1Handle.On("Stop").Return(errors.New("test1")).Once()
						mAgent2Handle.On("Stop").Return(nil).Once()
						err = handle.Stop()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test1")
					})

					Convey("Status() should return status of the master", func() {
						mMasterHandle.On("Status").Return(executor.RUNNING).Once()
						So(handle.Status(), ShouldEqual, executor.RUNNING)

						mMasterHandle.On("Status").Return(executor.RUNNING).Once()
						So(handle.Status(), ShouldEqual, executor.RUNNING)

						mMasterHandle.On("Status").Return(executor.TERMINATED).Once()
						So(handle.Status(), ShouldEqual, executor.TERMINATED)
					})

					Convey("ExitCode() should return exitCode of the master", func() {
						mMasterHandle.On("ExitCode").Return(1234, nil).Once()
						code, err := handle.ExitCode()
						So(code, ShouldEqual, 1234)
						So(err, ShouldBeNil)

						mMasterHandle.On("ExitCode").Return(0, errors.New("test")).Once()
						code, err = handle.ExitCode()
						So(code, ShouldEqual, 0)
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")
					})

					Convey("Wait() should wait for master and stop agents", func() {
						mMasterHandle.On("Wait", 0*time.Microsecond).Return(true).Once()
						mAgent1Handle.On("Stop").Return(nil).Once()
						mAgent2Handle.On("Stop").Return(nil).Once()
						So(handle.Wait(0), ShouldEqual, true)

						// Having errors during one stop should not break stopping all agents.
						mMasterHandle.On("Wait", 0*time.Microsecond).Return(true).Once()
						mAgent1Handle.On("Stop").Return(errors.New("test")).Once()
						mAgent2Handle.On("Stop").Return(nil).Once()
						So(handle.Wait(0), ShouldEqual, true)

						// While master is not terminated - should not stop agents.
						mMasterHandle.On("Wait", 0*time.Microsecond).Return(false).Once()
						So(handle.Wait(0), ShouldEqual, false)
					})

					Convey("Clean() should clean master's and agents' resources", func() {
						mMasterHandle.On("Clean").Return(nil).Once()
						mAgent1Handle.On("Clean").Return(nil).Once()
						mAgent2Handle.On("Clean").Return(nil).Once()
						So(handle.Clean(), ShouldBeNil)

						mMasterHandle.On("Clean").Return(errors.New("test")).Once()
						mAgent1Handle.On("Clean").Return(nil).Once()
						mAgent2Handle.On("Clean").Return(nil).Once()
						err := handle.Clean()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")

						mMasterHandle.On("Clean").Return(errors.New("test")).Once()
						mAgent1Handle.On("Clean").Return(errors.New("test1")).Once()
						mAgent2Handle.On("Clean").Return(errors.New("test2")).Once()
						err = handle.Clean()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test; test1; test2")

						mMasterHandle.On("Clean").Return(nil).Once()
						mAgent1Handle.On("Clean").Return(nil).Once()
						mAgent2Handle.On("Clean").Return(errors.New("test2")).Once()
						err = handle.Clean()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test2")
					})

					Convey("EraseOutput() should erase master's and agents' output", func() {
						mMasterHandle.On("EraseOutput").Return(nil).Once()
						mAgent1Handle.On("EraseOutput").Return(nil).Once()
						mAgent2Handle.On("EraseOutput").Return(nil).Once()
						So(handle.EraseOutput(), ShouldBeNil)

						mMasterHandle.On("EraseOutput").Return(errors.New("test")).Once()
						mAgent1Handle.On("EraseOutput").Return(nil).Once()
						mAgent2Handle.On("EraseOutput").Return(nil).Once()
						err := handle.EraseOutput()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test")

						mMasterHandle.On("EraseOutput").Return(errors.New("test")).Once()
						mAgent1Handle.On("EraseOutput").Return(errors.New("test1")).Once()
						mAgent2Handle.On("EraseOutput").Return(errors.New("test2")).Once()
						err = handle.EraseOutput()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test; test1; test2")

						mMasterHandle.On("EraseOutput").Return(nil).Once()
						mAgent1Handle.On("EraseOutput").Return(nil).Once()
						mAgent2Handle.On("EraseOutput").Return(errors.New("test2")).Once()
						err = handle.EraseOutput()
						So(err, ShouldNotBeNil)
						So(err.Error(), ShouldEqual, "test2")
					})

					Convey("Address() should return master address.", func() {
						mMasterHandle.On("Address").Return(testMasterAddress).Once()
						So(handle.Address(), ShouldEqual, testMasterAddress)
					})

				})

				So(mAgent1Handle.AssertExpectations(t), ShouldBeTrue)
				So(mAgent2Handle.AssertExpectations(t), ShouldBeTrue)
			})

			So(mMasterHandle.AssertExpectations(t), ShouldBeTrue)
		})
	})
}
