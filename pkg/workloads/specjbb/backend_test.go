package specjbb

import (
	"fmt"
	"testing"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
)

// TestBackendWithMockedExecutor runs a Backend launcher with the mocked executor.
func TestBackendWithMockedExecutor(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When using Backend launcher", t, func() {

		expectedCommand := fmt.Sprintf("java -jar -Dspecjbb.controller.host=127.0.0.1 test -m backend -G GRP1 -J JVM2"+
			" -p /usr/local/share/specjbb/config/specjbb2015.props", fs.GetSwanWorkloadsPath())
		expectedHost := "127.0.0.1"

		mockedExecutor := new(mocks.Executor)
		mockedTaskHandle := new(mocks.TaskHandle)
		config := DefaultSPECjbbBackendConfig()
		config.PathToBinary = "test"
		backendLauncher := NewBackend(mockedExecutor, config)

		Convey("While simulating proper execution", func() {
			mockedExecutor.On("Execute", expectedCommand).Return(mockedTaskHandle, nil).Once()
			mockedTaskHandle.On("Address").Return(expectedHost)

			Convey("Build command should create proper command", func() {
				command := backendLauncher.buildCommand()
				So(command, ShouldEqual, expectedCommand)

				Convey("Arguments passed to Executor should be a proper command", func() {
					task, err := backendLauncher.Launch()
					So(err, ShouldBeNil)

					So(task, ShouldNotBeNil)
					So(task, ShouldEqual, mockedTaskHandle)

					Convey("Location of the returned task shall be 127.0.0.1", func() {
						address := task.Address()
						So(address, ShouldEqual, expectedHost)
					})
				})
			})
		})

	})
}
