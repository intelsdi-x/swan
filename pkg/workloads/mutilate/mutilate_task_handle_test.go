package mutilate

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestMutilateTaskHandle(t *testing.T) {
	handle := MutilateTaskHandle{}
	Convey("MutilateTaskHandle implements TaskHandle", t, func() {
		So(handle, ShouldImplement, (*executor.TaskHandle)(nil))
	})
}
