package executor_test

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestExecutorFactory(t *testing.T) {
	Convey("Using Factory Executor", t, func() {
		Convey("When we point to the localhost ip, it should return Local Executor", func() {
			exec, err := executor.CreateExecutor("localhost")
			So(err, ShouldBeNil)

			_, ok := exec.(executor.Local)
			So(ok, ShouldBeTrue)
		})

		Convey("When we point to the 127.0.0.1 ip, it should return Local Executor", func() {
			exec, err := executor.CreateExecutor("127.0.0.1")
			So(err, ShouldBeNil)

			_, ok := exec.(executor.Local)
			So(ok, ShouldBeTrue)
		})

		Convey("When we point to the external ip, it should return Remote Executor", func() {
			exec, err := executor.CreateExecutor("255.255.255.255")
			if err != nil {
				t.Skip(err)
			}

			So(err, ShouldBeNil)

			_, ok := exec.(executor.Remote)
			So(ok, ShouldBeTrue)
		})
	})
}
