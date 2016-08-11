package kubernetes

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
	"testing"
	"time"
)

var serviceNames = []string{
	"kube-apiserver",
	"kube-controller-manager",
	"kube-scheduler",
	"kube-proxy",
	"kubelet",
}

type KubernetesTestSuite struct {
	suite.Suite

	mExecutor    *mocks.Executor
	mTaskHandles []*mocks.TaskHandle
	k8sLauncher  *kubernetes

	outputFile *os.File

	isListeningIteration int
}

func (s *KubernetesTestSuite) TestKubernetesLauncher() {
	Convey("While having mocked executor", s.T(), func() {
		s.mExecutor = new(mocks.Executor)
		// Create taskHandles with labels corresponding to service names.
		s.mTaskHandles = []*mocks.TaskHandle{}
		for _ = range serviceNames {
			mTaskHandle := new(mocks.TaskHandle)
			s.mTaskHandles = append(s.mTaskHandles, mTaskHandle)
		}

		var err error
		s.outputFile, err = ioutil.TempFile(os.TempDir(), "k8s")
		if err != nil {
			s.Fail(err.Error())
			return
		}
		defer s.outputFile.Close()

		Convey("While launching k8s cluster with default configuration", func() {
			k8sLauncher, ok := New(s.mExecutor, s.mExecutor, DefaultConfig()).(kubernetes)
			So(ok, ShouldBeTrue)

			s.k8sLauncher = &k8sLauncher

			// Start from first service (service[0]).
			s.testServiceCasesRecursively(0)

			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			for _, mTaskHandle := range s.mTaskHandles {
				So(mTaskHandle.AssertExpectations(s.T()), ShouldBeTrue)
			}
		})
	})
}

// testServiceCasesRecursively tests three cases which can happen for single service during
// Launcher work:
// a) failed `Execute` execution.
// b) failed `IsListening` function (checking the case, where the service started but is not listening.
// c) successful service creation
// 		In this particular case we can move to another service, so having mocked successful case
// 		for service 1 we can run the same function `testServiceCasesRecursively` to test service1+1.
//
// Since our launcher just spawn 5 services, to save LOC we test each of them using the this function.
func (s *KubernetesTestSuite) testServiceCasesRecursively(serviceIterator int) {
	Convey(fmt.Sprintf("When %q fails to execute, we expect error", serviceNames[serviceIterator]), func() {
		// Mock current service's executor failure.
		s.mExecutor.On(
			"Execute", mock.AnythingOfType("string")).Return(nil, errors.New("executor-fail")).Once()
		s.mExecutor.On("Name").Return("Local")
		s.mExecutor.On("EraseOutput").Return(nil)

		// Mock successful connection verifier for `iteration+1` iteration.
		// It will succeed for current service.
		s.isListeningIteration = 0
		s.k8sLauncher.isListening = func(string, time.Duration) bool {
			s.isListeningIteration++
			return s.isListeningIteration <= (serviceIterator + 1)
		}

		k8sHandle, err := s.k8sLauncher.Launch()
		So(k8sHandle, ShouldBeNil)
		So(err, ShouldNotBeNil)
		// Error should wrap the initial reason.
		So(err.Error(), ShouldContainSubstring, serviceNames[serviceIterator])
		So(err.Error(), ShouldContainSubstring, "execution of command")
		So(err.Error(), ShouldContainSubstring, "executor-fail")
	})

	Convey(fmt.Sprintf("When %q is not listening on the endpoint, we expect error and it's handle stopped", serviceNames[serviceIterator]), func() {
		// Mock current service's executor success.
		s.mExecutor.On(
			"Execute", mock.AnythingOfType("string")).Return(s.mTaskHandles[serviceIterator], nil).Once()
		s.mExecutor.On("Name").Return("Local")

		s.mTaskHandles[serviceIterator].On("StderrFile").Return(s.outputFile, nil)
		s.mTaskHandles[serviceIterator].On("StdoutFile").Return(s.outputFile, nil)
		s.mTaskHandles[serviceIterator].On("Address").Return(serviceNames[serviceIterator])
		s.mTaskHandles[serviceIterator].On("Stop").Return(nil)
		s.mTaskHandles[serviceIterator].On("Clean").Return(nil)
		s.mTaskHandles[serviceIterator].On("EraseOutput").Return(nil)
		s.mTaskHandles[serviceIterator].On("ExitCode").Return(1, nil)

		// Mock successful connection verifier for `iteration` iteration.
		// It will fail for current service.
		s.isListeningIteration = 0
		s.k8sLauncher.isListening = func(string, time.Duration) bool {
			s.isListeningIteration++
			return s.isListeningIteration <= (serviceIterator)
		}

		k8sHandle, err := s.k8sLauncher.Launch()
		So(k8sHandle, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, serviceNames[serviceIterator])
		So(err.Error(), ShouldContainSubstring, "failed to connect to service")
	})

	Convey(fmt.Sprintf("When %q execute successfully", serviceNames[serviceIterator]), func() {
		// Mock current service's executor success.
		s.mExecutor.On(
			"Execute", mock.AnythingOfType("string")).Return(s.mTaskHandles[serviceIterator], nil).Once()

		s.mTaskHandles[serviceIterator].On("Address").Return(serviceNames[serviceIterator]).Once()
		s.mTaskHandles[serviceIterator].On("Stop").Return(nil).Once()
		s.mTaskHandles[serviceIterator].On("Clean").Return(nil).Once()

		// Check if it is the last service.
		if serviceIterator < len(serviceNames)-1 {
			// kube-apiserver's Address is passed to other services so we need to add that as well.
			s.mTaskHandles[0].On("Address").Return(serviceNames[serviceIterator]).Once()

			// We did not test all of them yet. Go to another service.
			s.testServiceCasesRecursively(serviceIterator + 1)
		} else {
			// It is the last service so check the launcher's successful case.
			Convey("We expect launcher to return cluster handle and no error", func() {
				// Mock successful connection verifier for `iteration+1` iteration.
				// It will succeed for current service.
				s.isListeningIteration = 0
				s.k8sLauncher.isListening = func(string, time.Duration) bool {
					s.isListeningIteration++
					return s.isListeningIteration <= (serviceIterator + 1)
				}

				k8sHandle, err := s.k8sLauncher.Launch()
				So(k8sHandle, ShouldNotBeNil)
				So(err, ShouldBeNil)

				So(k8sHandle.Stop(), ShouldBeNil)
				So(k8sHandle.Clean(), ShouldBeNil)
			})
		}
	})
}

func TestKubernetesLauncherSuite(t *testing.T) {
	suite.Run(t, new(KubernetesTestSuite))
}
