package provisioning

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"github.com/intelsdi-x/swan/pkg/isolation"
)


// TestLocal takes fixed amount of time (6s) since it tests command execution and
// wait functionality.
func TestLocal(t *testing.T) {
	Convey("Using Local Shell with no isolation", t, func() {
		l := NewLocal("root", []isolation.Isolation{})

		Convey("When command `sleep 1` is executed and we wait for it", func() {
			start := time.Now()

			task := l.Run("sleep 1")

			task.Wait(0)

			duration := time.Since(start)
			durationsMs := duration.Nanoseconds() / 1e6

			Convey("These expectation needs to be made", func() {
				Convey("The command Duration should last longer than 1s", func() {
					So(durationsMs, ShouldBeGreaterThan, 1000)
				})

				Convey("And the exit status should be zero", func() {
					So(task.Status().code, ShouldEqual, 0)
				})
			})
		})

		Convey("When command `sleep 1` is executed and we wait for it with timeout 0.5s", func() {
			start := time.Now()

			task := l.Run("sleep 1")

			task.Wait(500)

			duration := time.Since(start)
			durationsMs := duration.Nanoseconds() / 1e6

			Convey("These expectation needs to be made", func() {
				Convey("The Duration should last less than 2s", func() {
					So(durationsMs, ShouldBeLessThan, 2000)
				})

				Convey("And the exit status should be zero", func() {
					So(task.Status().code, ShouldEqual, 0)
				})
			})
		})
	})
}
