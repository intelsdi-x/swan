package provisioning

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
	"github.com/intelsdi-x/swan/pkg/isolation"
)

func TestShell(t *testing.T) {
	Convey("Creating a new shell with `sleep 3`", t, func() {
		s := NewShell()

		Convey("Should take more than three second to execute", func() {
			start := time.Now()
			status := <-s.Execute("sleep 3", "local", []isolation.Isolation{})
			duration := time.Now().Sub(start)
			durationsMs := duration.Nanoseconds() / 1e6
			So(durationsMs, ShouldBeGreaterThan, 3000)

			Convey("And the exit status should be zero", func() {
				So(status.code, ShouldEqual, 0)
			})
		})
	})
}
