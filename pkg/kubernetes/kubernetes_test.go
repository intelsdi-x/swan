package kubernetes

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/pkg/errors"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"k8s.io/kubernetes/pkg/api"
)

func TestKubernetesLauncherConfiguration(t *testing.T) {
	Convey("When configuration is passed to Kubernetes Launcher", t, func() {
		config := DefaultConfig()
		handle := &mocks.TaskHandle{}
		handle.On("Address").Return("127.0.0.1")

		Convey("Privileged containers should be allowed to run by default", func() {
			So(getKubeAPIServerCommand(config), ShouldContainSubstring, "--allow-privileged=false")
			So(getKubeletCommand(handle, config), ShouldContainSubstring, "--allow-privileged=false")

			Convey("But they can be disallowed through configuration", func() {
				config.AllowPrivileged = true
				So(getKubeAPIServerCommand(config), ShouldContainSubstring, "--allow-privileged=true")
				So(getKubeletCommand(handle, config), ShouldContainSubstring, "--allow-privileged=true")
			})
		})

		Convey("Default etcd server address points to http://127.0.0.1:2379", func() {
			So(getKubeAPIServerCommand(config), ShouldContainSubstring, "--etcd-servers=http://127.0.0.1:2379")

			Convey("But etcd server location can be changed to arbitrary one", func() {
				config.EtcdServers = "http://1.1.1.1:1111,https://2.2.2.2:2222"
				So(getKubeAPIServerCommand(config), ShouldContainSubstring, "--etcd-servers="+config.EtcdServers)
			})
		})
		Convey("Any parameters passed to KubeAPI Server are escaped correctly", func() {
			config.KubeAPIArgs = "--admission-control=\"AlwaysAdmit,AddToleration\""
			So(getKubeAPIServerCommand(config), ShouldContainSubstring, " --admission-control=\"AlwaysAdmit,AddToleration\"")
		})

	})
}

func getMockedTaskHandle(outputFile *os.File) *mocks.TaskHandle {
	handle := new(mocks.TaskHandle)
	handle.On("StderrFile").Return(outputFile, nil)
	handle.On("StdoutFile").Return(outputFile, nil)
	handle.On("Address").Return("127.0.0.1")
	handle.On("Stop").Return(nil)
	handle.On("Clean").Return(nil)
	handle.On("EraseOutput").Return(nil)
	handle.On("ExitCode").Return(0, nil)

	return handle
}

func getNodeListFunc(resultNodes []api.Node, resultError error) getReadyNodesFunc {
	return func(k8sAPIAddress string) ([]api.Node, error) {
		return resultNodes, resultError
	}
}

func getIsListeningFunc(result bool) func(address string, timeout time.Duration) bool {
	return func(address string, timeout time.Duration) bool {
		return result
	}
}

func TestKubernetesLauncher(t *testing.T) {
	Convey("When testing Kubernetes Launcher", t, func() {
		// Prepare mocked output file for TaskHandles
		outputFile, err := ioutil.TempFile(os.TempDir(), "k8s-ut")
		So(err, ShouldBeNil)
		defer outputFile.Close()

		// Prepare Executor Mocks
		master := new(mocks.Executor)
		master.On("Name").Return("Mocked Executor")
		minion := new(mocks.Executor)
		minion.On("Name").Return("Mocked Executor")

		config := DefaultConfig()
		handle := getMockedTaskHandle(outputFile)

		// Prepare Kubernetes Launcher
		var k8s kubernetes
		k8s = New(master, minion, config).(kubernetes)

		Convey("When everything succeed, on Launch method we should receive not-nil task handle and no error", func() {
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			k8s.isListening = getIsListeningFunc(true)
			k8s.getReadyNodes = getNodeListFunc([]api.Node{api.Node{}}, nil)

			resultHandle, err := k8s.Launch()
			So(err, ShouldBeNil)
			So(resultHandle, ShouldNotBeNil)
		})
		Convey("When Minion executor fails to execte, we should receive nil task handle and an error", func() {
			err := errors.New("mocked-error")
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, err)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			k8s.isListening = getIsListeningFunc(true)
			k8s.getReadyNodes = getNodeListFunc([]api.Node{api.Node{}}, nil)

			resultHandle, err := k8s.Launch()
			So(err, ShouldNotBeNil)
			So(resultHandle, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, err.Error())
		})

		Convey("When Master executor fails to execte, we should receive nil task handle and an error", func() {
			err := errors.New("mocked-error")
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, err)
			k8s.isListening = getIsListeningFunc(true)
			k8s.getReadyNodes = getNodeListFunc([]api.Node{api.Node{}}, nil)

			resultHandle, err := k8s.Launch()
			So(err, ShouldNotBeNil)
			So(resultHandle, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, err.Error())
		})

		Convey("When Launcher cannot bind TCP connection to endpoint to check if service responds, we should receive an error", func() {
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			handle.On("Status").Return(executor.TERMINATED)
			k8s.isListening = getIsListeningFunc(false)
			k8s.getReadyNodes = getNodeListFunc([]api.Node{api.Node{}}, nil)

			resultHandle, err := k8s.Launch()
			So(err, ShouldNotBeNil)
			So(resultHandle, ShouldBeNil)

			Convey("Assert that task hadle is properly cleaned before returning", func() {
				handle.AssertCalled(t, "Stop")
				handle.AssertCalled(t, "Clean")
				handle.AssertCalled(t, "EraseOutput")
			})
		})

		Convey("When Kubelet cannot register to Master, we should receive an error", func() {
			err := errors.New("mocked-error")
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			k8s.isListening = getIsListeningFunc(true)
			k8s.getReadyNodes = getNodeListFunc(nil, err)

			resultHandle, err := k8s.Launch()
			So(err, ShouldNotBeNil)
			So(resultHandle, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, err.Error())

			Convey("Assert that task hadle is properly cleaned before returning", func() {
				handle.AssertCalled(t, "Stop")
				handle.AssertCalled(t, "Clean")
				handle.AssertCalled(t, "EraseOutput")
			})
		})
	})
}
