package isolation

import (
	"syscall"
	"testing"

	. "github.com/smartystreets/goconvey/convey"
)

func TestNamespaceDecorator(t *testing.T) {
	Convey("When I pass invalid namespace definition then decorator should not be instantiated", t, func() {
		decorator, err := NewNamespace(666)
		So(decorator, ShouldBeNil)
		So(err, ShouldNotBeNil)
	})

	Convey("When I pass valid namespace definition then decorator should be instantiated", t, func() {
		decorator, err := NewNamespace(syscall.CLONE_NEWPID)
		So(decorator, ShouldNotBeNil)
		So(err, ShouldBeNil)

		Convey("When I decorate with PID namespace then I should see correct string prepended to the command", func() {
			command := "Some random command"
			decorated := decorator.Decorate(command)
			So(decorated, ShouldEqual, "unshare --fork --pid --mount-proc Some random command")
		})
	})

	Convey("Wnen I pass definition of multiple namespace then decorator should be instantiated", t, func() {
		namespace := syscall.CLONE_NEWIPC | syscall.CLONE_NEWNET | syscall.CLONE_NEWNS | syscall.CLONE_NEWPID | syscall.CLONE_NEWUSER | syscall.CLONE_NEWUTS
		decorator, err := NewNamespace(namespace)
		So(decorator, ShouldNotBeNil)
		So(err, ShouldBeNil)

		Convey("When I decorate with multiple namespaces then decorator should work as expected", func() {
			command := "All the namespaces are all I want!"
			decorated := decorator.Decorate(command)
			So(decorated, ShouldContainSubstring, " --fork --pid --mount-proc ")
			So(decorated, ShouldContainSubstring, " --ipc ")
			So(decorated, ShouldContainSubstring, " --mount ")
			So(decorated, ShouldContainSubstring, " --uts ")
			So(decorated, ShouldContainSubstring, " --net ")
			So(decorated, ShouldContainSubstring, " --user ")
			So(decorated, ShouldStartWith, "unshare ")
			So(decorated, ShouldEndWith, " All the namespaces are all I want!")
		})
	})

	Convey("When I run multiple decorators then I should receive decorated command with expected ordering", t, func() {
		isolationNet, _ := NewNamespace(syscall.CLONE_NEWNET)
		isolationIpc, _ := NewNamespace(syscall.CLONE_NEWIPC)
		var decorators Decorators
		decorators = append(decorators, isolationNet, isolationIpc)

		decorated := decorators.Decorate("Some random command")
		So(decorated, ShouldEqual, "unshare --ipc unshare --net Some random command")
	})

	Convey("When I use empty array of decorators then I should receive unmodified command", t, func() {
		var decorators Decorators

		decorated := decorators.Decorate("This comand should be left intact")
		So(decorated, ShouldEqual, "This comand should be left intact")
	})

}
