// +build integration

package snap

import (
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

func TestSnap(t *testing.T) {
	Convey("Creating a Snap experiment session", t, func() {
		session, err := NewSession(
			"test-session",
			[]string{"/intel/mock/foo"},
			1*time.Second,
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
}
