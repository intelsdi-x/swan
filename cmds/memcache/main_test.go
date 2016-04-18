// dummy package to make sure that we depdencies are bundled correclty
package main

import (
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestDummy(t *testing.T) {
	Convey("dummy", t, func() {
		So("dummy", ShouldEqual, "dummy")

	})

}
