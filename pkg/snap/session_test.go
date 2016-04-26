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
			[]string{"/intel/swan/dummy/metric1", "/intel/swan/dummy/metric2"},
			1*time.Second,
		)

		Convey("Shouldn't return any errors", func() {
			So(err, ShouldBeNil)
		})

		Convey("When listing running session", func() {
			sessions, err := ListSessions()

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})

			Convey("There should be zero session", func() {
				So(len(sessions), ShouldEqual, 0)
			})
		})

		Convey("Starting a session", func() {
			err := session.Start()

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})
		})
	})
}
