package kubernetes

import (
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
)

//func recursiveConveyTest(t *testing.T) {
//	Convey("When we launch k8s cluster with default config", t, func() {
//
//	k8sHandle, err := k8sLauncher.Launch()
//	So(err, ShouldBeNil)
//
//}
//
//func TestKubernetes(t *testing.T) {
//	Convey("While having mocked executor", t, func() {
//		mExecutor := new(mocks.Executor{})
//		mTaskHandle := new(mocks.TaskHandle{})
//		Convey("When we launch k8s cluster with default config", t, func() {
//			k8sLauncher := New(mExecutor, mExecutor, DefaultConfig())
//
//
//			defer func() {
//				err := k8sHandle.Stop()
//				So(err, ShouldBeNil)
//				err = k8sHandle.Clean()
//				So(err, ShouldBeNil)
//				err = k8sHandle.EraseOutput()
//				So(err, ShouldBeNil)
//			}()
//		})
//	})
//}
