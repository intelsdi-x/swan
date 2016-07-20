package executor

import (
	"fmt"
	"io"
	"os"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/pkg/errors"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	swanKubernetesNamespace = api.NamespaceDefault
	defaultContainerName    = "swan"
)

// -----------------------------
// kuberentes and Pod config
type KubernetesConfig struct {
	PodName    string // unique pod identifier.
	Address    string // no default must be specifed or TODO: DefaultKubenertesConfig
	CpuRequest int64  // k8s/.../resource.DecimalSi unit
	CpuLimit   int64  // k8s/.../resouce.DecimalSi unit
	Decorators isolation.Decorators
}

// -----------------------------
// Executor
type kubernetesExecutor struct {
	config KubernetesConfig
	client *client.Client
}

func NewKubernetesExectuor(config KubernetesConfig) (*kubernetesExecutor, error) {

	// new
	k8s := &kubernetesExecutor{
		config: config,
	}

	restClientConfig := &restclient.Config{
		Host: config.Address,
		// Username: "test", // TODO authorization
		// Password: "password",
	}
	var err error
	k8s.client, err = client.New(restClientConfig)
	if err != nil {
		return nil, errors.Wrapf(err, "can't initilize kubernetes Client for host = %q", config.Address)
	}

	return k8s, nil
}

// prepareContainerResources helper to create ResourceRequirments for container.
func prepareContainerResources(CPULimit, CPURequest int64) api.ResourceRequirements {
	resourceListLimits := api.ResourceList{}
	resourceListRequests := api.ResourceList{}
	if CPULimit > 0 {
		resourceListRequests[api.ResourceCPU] = *resource.NewQuantity(CPULimit, resource.DecimalSI)
	}
	if CPURequest > 0 {
		resourceListRequests[api.ResourceCPU] = *resource.NewQuantity(CPURequest, resource.DecimalSI)
	}
	return api.ResourceRequirements{
		Limits:   resourceListLimits,
		Requests: resourceListRequests,
	}

}

func (k8s *kubernetesExecutor) Execute(command string) (*kubernetesTaskHandle, error) {

	podsAPI := k8s.client.Pods(swanKubernetesNamespace)

	for _, decorator := range k8s.config.Decorators {
		command = decorator.Decorate(command)
	}

	// Definition of Pod http://kubernetes.io/docs/api-reference/v1/definitions/.
	pod, err := podsAPI.Create(&api.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      k8s.config.PodName,
			Namespace: swanKubernetesNamespace,
			Labels:    map[string]string{"name": k8s.config.PodName},
		},
		Spec: api.PodSpec{
			Containers: []api.Container{
				api.Container{
					Name:      defaultContainerName,
					Image:     "jess/stress", // replace with image of swan
					Command:   []string{"sh", "-c", command},
					Resources: prepareContainerResources(k8s.config.CpuLimit, k8s.config.CpuRequest),
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrapf(err, "cannot schedule pod %q ns=%q: %s", k8s.config.PodName, swanKubernetesNamespace, err)
	}

	// req, err := labels.NewRequirement("name", labels.EqualsOperator, sets.NewString("jebac"))
	// selector := labels.NewSelector()
	// selector.Add(*req)
	//// for some reasons it doesn't work

	selectorRaw := fmt.Sprintf("name=%s", k8s.config.PodName)
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create an selector err:%s", selectorRaw)
	}

	// Prepare events watcher.
	watcher, err := podsAPI.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create a watcher over selector: %v", selector)
	}

	// One goroutine to watch over events.
	started, stopped := startWatching(watcher)
	<-started

	logsRequest := podsAPI.GetLogs(k8s.config.PodName, &api.PodLogOptions{
		Container: defaultContainerName,
	})
	logStream, err := logsRequest.Stream()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create a stream to get logsRequest selector: %+v", logsRequest)
	}

	// TODO: move that to constructor
	th := &kubernetesTaskHandle{
		podsAPI: podsAPI,
		pod:     pod,
		stopped: stopped}

	// Prepare local files
	stdoutFile, stderrFile, err := createExecutorOutputFiles(command, "local")
	if err != nil {
		return nil, errors.Wrapf(err, "createExecutorOutputFiles for command %q failed", command)
	}
	log.Debug("Created temporary files ",
		"stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	th.stdout = stdoutFile
	th.stderr = stderrFile
	th.errCh = make(chan error)

	// a gourtine that copies data from streamer to local files.
	go func() {
		mw := io.MultiWriter(stdoutFile, stderrFile)
		_, err := io.Copy(mw, logStream)
		if err != nil {
			th.errCh <- err
		}
	}()

	return th, nil
}

// ------------------------------------------------------
// TaskHandle
type kubernetesTaskHandle struct {
	podsAPI client.PodInterface
	pod     *api.Pod
	stopped chan struct{} // channel that is closed when Pod finished execution.
	stdout  *os.File
	stderr  *os.File
	errCh   chan error
}

func (th *kubernetesTaskHandle) isTerminated() bool {
	select {
	case <-th.stopped:
		return true
	default:
		return false
	}
}

func (th *kubernetesTaskHandle) Stop() error {
	if th.isTerminated() {
		return nil
	}

	var GracePeriodSeconds int64 = 0 // need an address

	log.Debugf("delete pod...")
	err := th.podsAPI.Delete(th.pod.Name, &api.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	})
	if err != nil {
		return errors.Wrapf(err, "cannot delete pod %q", th.pod.Name)
	}
	<-th.stopped
	return nil
}

func (th *kubernetesTaskHandle) Status() TaskState {
	panic("not implemented")
}

func (th *kubernetesTaskHandle) ExitCode() (int, error) {
	panic("not implemented")
}

func (th *kubernetesTaskHandle) Wait(timeout time.Duration) bool {
	if th.isTerminated() {
		return true
	}

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}
	select {
	case <-th.stopped:
		return true
	case <-timeoutChannel:
		return false
	}
}

func (th *kubernetesTaskHandle) Clean() error {
	panic("not implemented")
}

func (th *kubernetesTaskHandle) EraseOutput() error {
	panic("not implemented")
}

func (th *kubernetesTaskHandle) Address() string {
	panic("not implemented")
}

// ---------------------
// task info

// StdoutFile return local file from stream
func (th *kubernetesTaskHandle) StdoutFile() (*os.File, error) {
	panic("not implemented")
}

func (th *kubernetesTaskHandle) StderrFile() (*os.File, error) {
	panic("not implemented")
}

// ---------------------
// status watcher
func startWatching(watcher watch.Interface) (started, stopped chan struct{}) {
	started, stopped = make(chan struct{}), make(chan struct{})

	go func() {
		var startedAlreadyClosed bool //guard to not close "started" channel twice.
		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*api.Pod)
			// http://kubernetes.io/docs/user-guide/pod-states/
			if ok {
				log.Debugf("pod event: %v", pod)
				switch pod.Status.Phase {
				case api.PodPending:
					log.Debugf("pod %q is pending", pod.Name)
				case api.PodRunning:
					log.Debugf("pod %q running (closed=%v)", pod.Name, startedAlreadyClosed)
					if !startedAlreadyClosed { // we get two events about running, but I can close channel only once.
						close(started)
					} else {
						log.Fatal("WTF?")
					}
				case api.PodFailed, api.PodSucceeded, api.PodUnknown:
					log.Debugf("pod %q stopped: with status: %+v", pod.Name, pod.Status)
					close(stopped)
					return
				}
			}
		}
	}()
	return started, stopped
}
