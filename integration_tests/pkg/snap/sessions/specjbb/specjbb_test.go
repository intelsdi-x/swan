package specjbb

import (
	"path"
	"testing"

	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/specjbb"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapSpecJbbSession(t *testing.T) {

	Convey("When testing SpecJbbSnaptelSession ", t, func() {
		Convey("We have snapteld running ", func() {

			cleanupSnaptel, loader, snapteldAddress := testhelpers.RunAndTestSnaptel()
			defer cleanupSnaptel()

			Convey("And we loaded publisher plugin", func() {

				cleanupMerticsFile, publisher, publisherDataFilePath := testhelpers.PreparePublisher(loader)
				defer cleanupMerticsFile()

				Convey("Then we prepared and launch specjbb session", func() {

					specjbbSessionConfig := specjbbsession.DefaultConfig()
					specjbbSessionConfig.SnapteldAddress = snapteldAddress
					specjbbSessionConfig.Publisher = publisher
					specjbbSnaptelSession, err := specjbbsession.NewSessionLauncher(specjbbSessionConfig)
					So(err, ShouldBeNil)

					cleanupMockedFile, mockedTaskInfo := testhelpers.PrepareMockedTaskInfo(path.Join(
						testhelpers.SwanPath, "misc/snap-plugin-collector-specjbb/specjbb/specjbb.stdout"))
					defer cleanupMockedFile()

					handle, err := specjbbSnaptelSession.LaunchSession(mockedTaskInfo, "foo:bar")
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

							dataValid := testhelpers.ReadAndTestPublisherData(publisherDataFilePath, expectedMetrics)
							So(dataValid, ShouldBeTrue)
						})
					})
				})
			})
		})
	})
}
