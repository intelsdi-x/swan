package errcollection

import (
	"errors"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestErrorCollection(t *testing.T) {
	Convey("When use ErrorCollection", t, func() {
		var errCollection ErrorCollection

		Convey("When no error was passed, GetErr should return nil", func() {
			So(errCollection.GetErrIfAny(), ShouldBeNil)
		})
		Convey("When nil error was passed, GetErr should return nil", func() {
			errCollection.Add(nil)
			So(errCollection.GetErrIfAny(), ShouldBeNil)
		})

		Convey("When we pass one error, GetErr should return error with exact message", func() {
			errCollection.Add(errors.New("test error"))
			So(errCollection.GetErrIfAny(), ShouldNotBeNil)
			So(errCollection.GetErrIfAny().Error(), ShouldEqual, "test error")
		})

		Convey("When we pass multiple errors, GetErr should return error with combined messages", func() {
			errCollection.Add(errors.New("test error"))
			errCollection.Add(errors.New("test error1"))
			errCollection.Add(errors.New("test error2"))
			So(errCollection.GetErrIfAny(), ShouldNotBeNil)
			So(errCollection.GetErrIfAny().Error(), ShouldEqual, "test error;\n test error1;\n test error2")
		})
	})
}
