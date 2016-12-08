package specjbbsession_test

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/athena/pkg/snap/sessions/specjbb"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"golang.org/x/oauth2/github"
)

func TestSnapSpecJbbSession(t *testing.T) {

	var snapd *testhelpers.Snapd
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("While having Snapd running", t, func() {
		snapd = testhelpers.NewSnapd()
		err := snapd.Start()
		So(err, ShouldBeNil)


		defer func() {
			if snapd != nil {
				err := snapd.Stop()
				err2 := snapd.CleanAndEraseOutput()

				So(err, ShouldBeNil)
				So(err2, ShouldBeNil)
			}
		}()

		// Wait until snap is up.
		So(snapd.Connected(), ShouldBeTrue)

		snapdAddress := fmt.Sprintf("http://127.0.0.1:%d", snapd.Port())

		loaderConfig := snap.DefaultPluginLoaderConfig()
		loaderConfig.SnapdAddress = snapdAddress
		loader, err := snap.NewPluginLoader(loaderConfig)
		So(err, ShouldBeNil)

		Convey("We are able to connect with snapd", func() {
			Convey("Loading test publisher", func() {
				tmpFile, err := ioutil.TempFile("", "session_test")
				So(err, ShouldBeNil)
				tmpFile.Close()

				metricsFile = tmpFile.Name()
				defer os.Remove(metricsFile)

				loader.Load(snap.SessionPublisher)
				pluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)
				So(err, ShouldBeNil)

				publisher = wmap.NewPublishNode(pluginName, snap.PluginAnyVersion)
				So(publisher, ShouldNotBeNil)

				Convey("While launching SpecJbbSnapSession", func() {
					specjbbSessionConfig := specjbbsession.DefaultConfig()
					specjbbSessionConfig.SnapdAddress = snapdAddress
					specjbbSessionConfig.Publisher = publisher
					specjbbSnapSession,err := specjbbsession.NewSessionLauncher(specjbbSessionConfig)
					So(err, ShouldBeNil)

					mockedTaskInfo := new(mocks.TaskInfo)
					specjbbStdoutPath := path.Join(
						os.Getenv("GOPATH"), "src/github.com/intelsdi-x/swan/misc/snap-plugin-collector-specjbb/specjbb/specjbb.stdout")

					file, err := os.Open(specjbbStdoutPath)

					So(err, ShouldBeNil)
					defer file.Close()

					mockedTaskInfo.On("StdoutFile").Return(file, nil)

					handle, err := specjbbSnapSession.LaunchSession(mockedTaskInfo, "foo:bar")
					So(err, ShouldBeNil)

					defer func() {
						err := handle.Stop()
						So(err, ShouldBeNil)
					}()




				})

			})

		})


	})


}