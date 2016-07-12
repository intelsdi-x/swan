package kubernetes

import (
	"testing"

	"errors"
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/suite"
	"io/ioutil"
	"os"
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

func (s *KubernetesTestSuite) recursiveConveyTest(iteration int) {
	Convey(fmt.Sprintf("When %q fails to execute, we expect error", serviceNames[iteration]), func() {
		// Mock current service's executor failure.
		s.mExecutor.On(
			"Execute", mock.AnythingOfType("string")).Return(nil, errors.New("executor-fail")).Once()

		// Mock successful connection verifier for `iteration+1` iteration.
		// It will succeed for current service.
		s.isListeningIteration = 0
		s.k8sLauncher.isListening = func(string, time.Duration) bool {
			s.isListeningIteration++
			return s.isListeningIteration <= (iteration + 1)
		}

		k8sHandle, err := s.k8sLauncher.Launch()
		So(k8sHandle, ShouldBeNil)
		So(err, ShouldNotBeNil)
		// Error should wrap the initial reason.
		So(err.Error(), ShouldContainSubstring, serviceNames[iteration])
		So(err.Error(), ShouldContainSubstring, "Execution of service failed")
		So(err.Error(), ShouldContainSubstring, "executor-fail")
	})

	Convey(fmt.Sprintf("When %q is not listening on the endpoint, we expect error and it's handle stopped", serviceNames[iteration]), func() {
		// Mock current service's executor success.
		s.mExecutor.On(
			"Execute", mock.AnythingOfType("string")).Return(s.mTaskHandles[iteration], nil).Once()

		s.mTaskHandles[iteration].On("StderrFile").Return(s.outputFile, nil).Once()
		s.mTaskHandles[iteration].On("Address").Return(serviceNames[iteration]).Once()
		s.mTaskHandles[iteration].On("Stop").Return(nil).Once()
		s.mTaskHandles[iteration].On("Clean").Return(nil).Once()

		// Mock successful connection verifier for `iteration` iteration.
		// It will fail for current service.
		s.isListeningIteration = 0
		s.k8sLauncher.isListening = func(string, time.Duration) bool {
			s.isListeningIteration++
			return s.isListeningIteration <= (iteration)
		}

		k8sHandle, err := s.k8sLauncher.Launch()
		So(k8sHandle, ShouldBeNil)
		So(err, ShouldNotBeNil)
		So(err.Error(), ShouldContainSubstring, serviceNames[iteration])
		So(err.Error(), ShouldContainSubstring, "Failed to connect to service on instance.")
	})

	Convey(fmt.Sprintf("When %q execute successfully", serviceNames[iteration]), func() {
		// Mock current service's executor success.
		s.mExecutor.On(
			"Execute", mock.AnythingOfType("string")).Return(s.mTaskHandles[iteration], nil).Once()

		s.mTaskHandles[iteration].On("Address").Return(serviceNames[iteration]).Once()
		s.mTaskHandles[iteration].On("Stop").Return(nil).Once()
		s.mTaskHandles[iteration].On("Clean").Return(nil).Once()

		// Check if it is the last service.
		if iteration < len(serviceNames)-1 {
			// kube-apiserver's Address is passed to other services so we need to add that as well.
			s.mTaskHandles[0].On("Address").Return(serviceNames[iteration]).Once()

			// We did not test all of them yet. Go to another service.
			s.recursiveConveyTest(iteration + 1)
		} else {
			// It is the last service so check the launcher's successful case.
			Convey("We expect launcher to return cluster handle and no error", func() {
				// Mock successful connection verifier for `iteration+1` iteration.
				// It will succeed for current service.
				s.isListeningIteration = 0
				s.k8sLauncher.isListening = func(string, time.Duration) bool {
					s.isListeningIteration++
					return s.isListeningIteration <= (iteration + 1)
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

func (s *KubernetesTestSuite) TestKubernetesLauncher() {
	Convey("While having mocked executor", s.T(), func() {
		s.mExecutor = new(mocks.Executor)
		// Create taskHandles with labels corresponding to service names.
		s.mTaskHandles = []*mocks.TaskHandle{}
		for _, _ = range serviceNames {
			mTaskHandle := new(mocks.TaskHandle)
			// TODO(bp): Add that after https://github.com/intelsdi-x/swan/pull/263 merge.
			//mTaskHandle.Label(serviceName)

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

			// Start from 0 iteration.
			s.recursiveConveyTest(0)

			So(s.mExecutor.AssertExpectations(s.T()), ShouldBeTrue)
			for _, mTaskHandle := range s.mTaskHandles {
				So(mTaskHandle.AssertExpectations(s.T()), ShouldBeTrue)
			}
		})
	})
}

func TestKubernetesLauncherSuite(t *testing.T) {
	suite.Run(t, new(KubernetesTestSuite))
}
