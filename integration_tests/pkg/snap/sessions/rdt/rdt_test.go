package specjbb

import (
	"io/ioutil"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/rdt"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnaptelSpecJbbSession(t *testing.T) {

	Convey("When testing RDT Session ", t, func() {
		Convey("We have snapteld running ", func() {

			cleanupSnaptel, loader, snapteldAddress := testhelpers.RunAndTestSnaptel()
			defer cleanupSnaptel()

			Convey("And we loaded publisher plugin", func() {

				tmpFile, err := ioutil.TempFile("", "rdt-session-test")
				So(err, ShouldBeNil)

				publisherMetricsFile := tmpFile.Name()
				loader.Load(snap.FilePublisher)

				pluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)
				So(err, ShouldBeNil)

				publisher := wmap.NewPublishNode(pluginName, snap.PluginAnyVersion)
				So(publisher, ShouldNotBeNil)

				publisher.AddConfigItem("file", publisherMetricsFile)

				Convey("Then we prepared and launch specjbb session", func() {

					rdtSessionConfig := rdt.DefaultConfig()
					rdtSessionConfig.SnapteldAddress = snapteldAddress
					rdtSessionConfig.Publisher = publisher
					rdtSession, err := rdt.NewSessionLauncher(rdtSessionConfig)
					So(err, ShouldBeNil)

					handle, err := rdtSession.LaunchSession(nil, "foo:bar")
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()

					time.Sleep(5 * time.Second)
					Convey("Later we checked if task is running", func() {
						So(handle.IsRunning(), ShouldBeTrue)

						Convey("In order to read and test published data", func() {
							content, err := ioutil.ReadFile(tmpFile.Name())
							So(err, ShouldBeNil)
							So(content, ShouldNotBeEmpty)
						})
					})
				})
			})
		})
	})
}
