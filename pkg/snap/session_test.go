// +build integration

package snap

import (
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {
	Convey("Connecting to snapd", t, func() {

		client, err := client.New("http://localhost:8181", "v1", true)
		Convey("Shouldn't return any errors", func() {
			So(err, ShouldBeNil)
		})

		Convey("Unloading collectors", func() {
			plugins := NewPlugins(client)

			Convey("Loading collectors", func() {
				plugins.Load("snap-collector-session-test")
			})

			Convey("Loading publisher", func() {
				plugins.Load("snap-publisher-session-test")

				time.Sleep(1 * time.Second)

				// TODO(niklas): Block until plugins are indeed loaded. May be a race if plugin load is slow.

				publisher := wmap.NewPublishNode("session-test", 1)
				publisher.AddConfigItem("file", "/tmp/swan-snap.out")

				Convey("Creating a Snap experiment session", func() {
					session, err := NewSession(
						"test-session",
						[]string{"/intel/swan/session/metric1"},
						1*time.Second,
						client,
						publisher,
					)

					Convey("Shouldn't return any errors", func() {
						So(err, ShouldBeNil)
					})

					Convey("Starting a session", func() {
						err := session.Start()
						defer session.Stop()

						Convey("Shouldn't return any errors", func() {
							So(err, ShouldBeNil)
						})

						Convey("Contacting snap to get the task status", func() {
							status, err := session.Status()

							Convey("Shouldn't return any errors", func() {
								So(err, ShouldBeNil)
							})

							Convey("And the task should be running", func() {
								So(status, ShouldEqual, "Running")
							})
						})

						// TODO(niklas): Make sure metrics are processed. Introduce WaitUntil("hits", 1).
						time.Sleep(5 * time.Second)

						// TODO(niklas): Verify that the labels have swan session label.

						Convey("Stopping a session", func() {
							err := session.Stop()

							Convey("Shouldn't return any errors", func() {
								So(err, ShouldBeNil)
							})

							Convey("And the task should not be available", func() {
								_, err := session.Status()
								So(err, ShouldNotBeNil)
							})
						})
					})
				})
			})
		})
	})
}
