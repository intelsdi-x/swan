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
	"k8s.io/client-go/1.5/pkg/api/v1"
)

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

func getNodeListFunc(resultNodes []v1.Node, resultError error) getReadyNodesFunc {
	return func(k8sAPIAddress string) ([]v1.Node, error) {
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
		master.On("Name").Return("Master Executor")
		minion := new(mocks.Executor)
		minion.On("Name").Return("Minion Executor")

		config := DefaultConfig()
		handle := getMockedTaskHandle(outputFile)

		// Prepare Kubernetes Launcher
		var k8sLauncher k8s
		k8sLauncher = New(master, minion, config).(k8s)

		Convey("When configuration is passed to Kubernetes Launcher", func() {
			handle := &mocks.TaskHandle{}
			handle.On("Address").Return("127.0.0.1")

			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			k8sLauncher.isListening = getIsListeningFunc(true)
			k8sLauncher.getReadyNodes = getNodeListFunc([]v1.Node{}, nil)

			Convey("Privileged containers should be allowed to run by default", func() {
				k8sLauncher.config.KubeletPort = 1234
				kubeApiCommand := k8sLauncher.getKubeAPIServerCommand()
				kubeletCommand := k8sLauncher.getKubeletCommand()

				So(kubeApiCommand.raw, ShouldContainSubstring, "--allow-privileged=false")
				So(kubeApiCommand.healthCheckPort, ShouldEqual, 8080)
				So(kubeApiCommand.exec.Name(), ShouldEqual, "Master Executor")

				So(kubeletCommand.raw, ShouldContainSubstring, "--allow-privileged=false")
				So(kubeletCommand.healthCheckPort, ShouldEqual, 1234)
				So(kubeApiCommand.exec.Name(), ShouldEqual, "Master Executor")

				Convey("But they can be disallowed through configuration", func() {
					k8sLauncher.config.AllowPrivileged = true
					kubeApiCommand := k8sLauncher.getKubeAPIServerCommand()
					kubeletCommand := k8sLauncher.getKubeletCommand()

					So(kubeApiCommand.raw, ShouldContainSubstring, "--allow-privileged=true")
					So(kubeletCommand.raw, ShouldContainSubstring, "--allow-privileged=true")
				})
			})

			Convey("Default etcd server address points to http://127.0.0.1:2379", func() {
				kubeApiCommand := k8sLauncher.getKubeAPIServerCommand()
				So(kubeApiCommand.raw, ShouldContainSubstring, "--etcd-servers=http://127.0.0.1:2379")
				So(kubeApiCommand.exec.Name(), ShouldEqual, "Master Executor")

				Convey("But etcd server location can be changed to arbitrary one", func() {
					k8sLauncher.config.EtcdServers = "http://1.1.1.1:1111,https://2.2.2.2:2222"
					kubeApiCommand := k8sLauncher.getKubeAPIServerCommand()
					So(kubeApiCommand.raw, ShouldContainSubstring, "--etcd-servers="+k8sLauncher.config.EtcdServers)
					So(kubeApiCommand.exec.Name(), ShouldEqual, "Master Executor")
				})
			})
			Convey("Any parameters passed to KubeAPI Server are escaped correctly", func() {
				k8sLauncher.config.KubeAPIArgs = "--admission-control=\"AlwaysAdmit,AddToleration\""
				kubeApiCommand := k8sLauncher.getKubeAPIServerCommand()
				So(kubeApiCommand.raw, ShouldContainSubstring, " --admission-control=\"AlwaysAdmit,AddToleration\"")
			})

		})

		Convey("When everything succeed, on Launch method we should receive not-nil task handle and no error", func() {
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			k8sLauncher.isListening = getIsListeningFunc(true)
			k8sLauncher.getReadyNodes = getNodeListFunc([]v1.Node{v1.Node{}}, nil)

			resultHandle, err := k8sLauncher.Launch()
			So(err, ShouldBeNil)
			So(resultHandle, ShouldNotBeNil)
		})
		Convey("When Minion executor fails to execte, we should receive nil task handle and an error", func() {
			err := errors.New("mocked-error")
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, err)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			k8sLauncher.isListening = getIsListeningFunc(true)
			k8sLauncher.getReadyNodes = getNodeListFunc([]v1.Node{v1.Node{}}, nil)

			resultHandle, err := k8sLauncher.Launch()
			So(err, ShouldNotBeNil)
			So(resultHandle, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, err.Error())
		})

		Convey("When Master executor fails to execte, we should receive nil task handle and an error", func() {
			err := errors.New("mocked-error")
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, err)
			k8sLauncher.isListening = getIsListeningFunc(true)
			k8sLauncher.getReadyNodes = getNodeListFunc([]v1.Node{v1.Node{}}, nil)

			resultHandle, err := k8sLauncher.Launch()
			So(err, ShouldNotBeNil)
			So(resultHandle, ShouldBeNil)
			So(err.Error(), ShouldContainSubstring, err.Error())
		})

		Convey("When Launcher cannot bind TCP connection to endpoint to check if service responds, we should receive an error", func() {
			minion.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			master.On("Execute", mock.AnythingOfType("string")).Return(handle, nil)
			handle.On("Status").Return(executor.TERMINATED)
			k8sLauncher.isListening = getIsListeningFunc(false)
			k8sLauncher.getReadyNodes = getNodeListFunc([]v1.Node{v1.Node{}}, nil)

			resultHandle, err := k8sLauncher.Launch()
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
			k8sLauncher.isListening = getIsListeningFunc(true)
			k8sLauncher.getReadyNodes = getNodeListFunc(nil, err)

			resultHandle, err := k8sLauncher.Launch()
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
