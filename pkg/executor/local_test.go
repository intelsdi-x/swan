package executor

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

// TestLocal tests the execution of process on local machine.
func TestLocal(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using Local Shell", t, func() {
		l := NewLocal()

		Convey("The generic Executor test should pass", func() {
			TestExecutor(t, l)
		})
	})
}
