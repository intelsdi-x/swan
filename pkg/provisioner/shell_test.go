package provisioner

import (
	. "github.com/smartystreets/goconvey/convey"
	"testing"
	"time"
)

func TestShell(t *testing.T) {
	Convey("Creating a new empty shell", t, func() {
		s := NewShell([]*Command{})

		Convey("Should leave zero commands in shell to execute", func() {
			So(s.LenCommands(), ShouldEqual, 0)
		})
	})

	Convey("Creating a new shell with `sleep 1`", t, func() {
		s := NewShell([]*Command{NewCommand("sleep 1")})

		Convey("Should leave one command in shell to execute", func() {
			So(s.LenCommands(), ShouldEqual, 1)
		})

		Convey("Should take more than one second to execute", func() {
			start := time.Now()
			<-s.Execute()
			duration := time.Now().Sub(start)
			duration_ms := duration.Nanoseconds() / 1e6
			So(duration_ms, ShouldBeGreaterThan, 1000)

			Convey("And the exit status should be zero", func() {
				// So(len(statuses), ShouldEqual, 1)
			})
		})
	})
}
