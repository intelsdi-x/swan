// +build integration

package snap

import (
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {

	Convey("Loading plugins", t, func() {
		UnloadPlugin("collector", "session-test", 1)

		err := LoadPlugin("snap-collector-session-test")
		Convey("Shouldn't return any errors", func() {
			So(err, ShouldBeNil)
		})

		publisher := wmap.NewPublishNode("session-publisher", 1)
		publisher.AddConfigItem("file", "/tmp/swan-snap.out")

		// TODO(niklas): Test should create publisher which is configured to write to shared temporary file.

		Convey("Creating a Snap experiment session", func() {
			session, err := NewSession(
				"test-session",
				[]string{"/intel/swan/session/metric1"},
				1*time.Second,
				publisher,
			)

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("Starting a session", func() {
				err := session.Start()
				// defer session.Stop()

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

				// Convey("Stopping a session", func() {
				// 	err := session.Stop()
				//
				// 	Convey("Shouldn't return any errors", func() {
				// 		So(err, ShouldBeNil)
				// 	})
				//
				// 	Convey("And the task should not be available", func() {
				// 		_, err := session.Status()
				// 		So(err, ShouldNotBeNil)
				// 	})
				// })
			})
		})
	})
}
