package specjbbsessiontest

import (
	"github.com/intelsdi-x/athena/pkg/snap/sessions/specjbb"
	"github.com/intelsdi-x/swan/integration_tests/pkg/snap/sessions"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	"path"
	"testing"
)

func TestAaaSnapSpecJbbSession(t *testing.T) {

	Convey("When testing SpecJbbSnapSession ", t, func() {
		Convey("We have snapd running ", func() {

			cleanaupSnap, loader, snapdAddress := sessions.RunAndTestSnap()
			defer cleanaupSnap()

			Convey("And we loaded publisher plugin", func() {

				clenupMerticsFile, publisher, publisherDataFilePath := sessions.PrepareAndTestPublisher(loader)
				defer clenupMerticsFile()

				Convey("Then we prepared and launch specjbb session", func() {

					specjbbSessionConfig := specjbbsession.DefaultConfig()
					specjbbSessionConfig.SnapdAddress = snapdAddress
					specjbbSessionConfig.Publisher = publisher
					specjbbSnapSession, err := specjbbsession.NewSessionLauncher(specjbbSessionConfig)
					So(err, ShouldBeNil)

					cleanupMockedFile, mockedTaskInfo := sessions.PrepareMockedTask(path.Join(
						fs.GetSwanPath(), "misc/snap-plugin-collector-specjbb/specjbb/specjbb.stdout"))
					defer cleanupMockedFile()

					handle, err := specjbbSnapSession.LaunchSession(mockedTaskInfo, "foo:bar")
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()

					Convey("Later we checked if task is running", func() {
						So(handle.IsRunning(), ShouldBeTrue)

						// These are results from test output file
						// in "src/github.com/intelsdi-x/swan/misc/
						// snap-plugin-collector-specjbb/specjbb/specjbb.stdout"
						expectedMetrics := map[string]string{
							"min":             "300",
							"50th":            "3100",
							"90th":            "21000",
							"95th":            "89000",
							"99th":            "517000",
							"max":             "640000",
							"qps":             "4007",
							"issued_requests": "4007",
						}

						Convey("In order to read and test published data", func() {

							dataValid := sessions.ReadAndTestPublisherData(publisherDataFilePath, expectedMetrics, t)
							So(dataValid, ShouldBeTrue)
						})
					})
				})
			})
		})
	})
}
