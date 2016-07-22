package executor

import (
	"fmt"
	"io"
	"os"
	"path"
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

// KubernetesConfig describes the necessary information to connect to a Kubernetes cluster.
type KubernetesConfig struct {
	PodName    string // unique pod identifier.
	Address    string // no default must be specifed or TODO: DefaultKubenertesConfig
	CPURequest int64  // k8s/.../resource.DecimalSi unit
	CPULimit   int64  // k8s/.../resouce.DecimalSi unit
	Decorators isolation.Decorators
}

type kubernetesExecutor struct {
	config KubernetesConfig
	client *client.Client
}

// NewKubernetesExecutor returns an executor which lets the user run commands in pods in a
// kubernetes cluster.
func NewKubernetesExecutor(config KubernetesConfig) (Executor, error) {
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

// Execute creates a pod and runs the provided command in it. When the command completes, the pod
// is stopped i.e. the container is not restarted automatically.
func (k8s *kubernetesExecutor) Execute(command string) (TaskHandle, error) {
	podsAPI := k8s.client.Pods(swanKubernetesNamespace)

	for _, decorator := range k8s.config.Decorators {
		command = decorator.Decorate(command)
	}

	// See http://kubernetes.io/docs/api-reference/v1/definitions/ for definition of the pod manifest.
	pod, err := podsAPI.Create(&api.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      k8s.config.PodName,
			Namespace: swanKubernetesNamespace,
			Labels:    map[string]string{"name": k8s.config.PodName},
		},
		Spec: api.PodSpec{
			RestartPolicy: "Never",
			Containers: []api.Container{
				api.Container{
					Name:      defaultContainerName,
					Image:     "jess/stress", // replace with image of swan
					Command:   []string{"sh", "-c", command},
					Resources: prepareContainerResources(k8s.config.CPULimit, k8s.config.CPURequest),
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrapf(err, "cannot schedule pod %q ns=%q: %s", k8s.config.PodName, swanKubernetesNamespace, err)
	}

	th := newKubernetesTaskHandle(command, podsAPI, pod)

	// NOTE: We should have timeout for the amount of time we want to wait for the pod to appear.
	<-th.started

	th.setupLogs()

	return th, nil
}

// kubernetesTaskHandle implements the TaskHandle interface
type kubernetesTaskHandle struct {
	podsAPI  client.PodInterface
	pod      *api.Pod
	command  string
	stopped  chan struct{}
	started  chan struct{}
	stdout   *os.File
	stderr   *os.File
	errCh    chan error
	exitCode *int
}

func newKubernetesTaskHandle(command string, podsAPI client.PodInterface, pod *api.Pod) *kubernetesTaskHandle {
	th := &kubernetesTaskHandle{
		command: command,
		podsAPI: podsAPI,
		pod:     pod,
	}
	th.watch()
	return th
}

func (th *kubernetesTaskHandle) watch() error {
	selectorRaw := fmt.Sprintf("name=%s", th.pod.Name)
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		return errors.Wrapf(err, "cannot create an selector err:%s", selectorRaw)
	}

	// Prepare events watcher.
	watcher, err := th.podsAPI.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "cannot create a watcher over selector: %v", selector)
	}

	th.started = make(chan struct{})
	th.stopped = make(chan struct{})

	go func() {
		// var once sync.Once
		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*api.Pod)
			if ok {
				// Update with latest status.
				th.pod = pod

				switch event.Type {
				case watch.Added:
					// NOTE: Replicate switch statement below.

				case watch.Modified:
					switch pod.Status.Phase {
					case api.PodPending:
						log.Debugf("modified event: '%s' in PodPending phase", pod.Name)
					case api.PodFailed:
						log.Debugf("modified event: '%s' in PodFailed phase", pod.Name)
						exitCode := int(1)
						th.exitCode = &exitCode
						close(th.stopped)

						// Try to delete the failed pod to avoid conflicts and having to call Stop()
						// after the stopped channel has been closed.
						var GracePeriodSeconds int64
						th.podsAPI.Delete(th.pod.Name, &api.DeleteOptions{
							GracePeriodSeconds: &GracePeriodSeconds,
						})
					case api.PodSucceeded:
						log.Debugf("modified event: '%s' in PodSucceeded phase", pod.Name)
						exitCode := int(0)
						th.exitCode = &exitCode
						close(th.stopped)

					case api.PodRunning:
						log.Debugf("modified event: '%s' in PodRunning phase", pod.Name)

						if api.IsPodReady(pod) {
							close(th.started)
						}
					default:
						log.Debugf("unknown pod.Status.Phase '%d' for pod '%s'", pod.Status.Phase, pod.Name)
					}
				case watch.Deleted:
					log.Debugf("pod '%s' deleted", pod.Name)
					close(th.stopped)
				default:
					log.Debugf("unknown event.Type")
				}
			}
		}
	}()

	return nil
}

// NOTE: That setupLogs can only be called when the pod is running i.e. wait until the started
// channel has been closed by watch().
func (th *kubernetesTaskHandle) setupLogs() error {
	// Wire up logs to task handle stdout.
	logsRequest := th.podsAPI.GetLogs(th.pod.Name, &api.PodLogOptions{
		Container: defaultContainerName,
	})
	logStream, err := logsRequest.Stream()
	if err != nil {
		return errors.Wrapf(err, "cannot create a stream to get logsRequest selector: %+v", logsRequest)
	}

	// Prepare local files
	stdoutFile, stderrFile, err := createExecutorOutputFiles(th.command, "local")
	if err != nil {
		return errors.Wrapf(err, "createExecutorOutputFiles for pod '%s' failed", th.pod.Name)
	}
	log.Debug("created temporary files stdout path:  ", stdoutFile.Name(), ", stderr path:  ", stderrFile.Name())

	th.stdout = stdoutFile
	th.stderr = stderrFile
	th.errCh = make(chan error)

	// a goroutine that copies data from streamer to local files.
	go func() {
		mw := io.MultiWriter(stdoutFile, stderrFile)
		_, err := io.Copy(mw, logStream)
		if err != nil {
			th.errCh <- err
		}
	}()

	return nil
}

func (th *kubernetesTaskHandle) isTerminated() bool {
	select {
	case <-th.stopped:
		return true
	default:
		return false
	}
}

// Stop will delete the pod and block caller until done.
func (th *kubernetesTaskHandle) Stop() error {
	if th.isTerminated() {
		return nil
	}

	log.Debugf("deleting pod '%s'", th.pod.Name)

	var GracePeriodSeconds int64
	err := th.podsAPI.Delete(th.pod.Name, &api.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	})
	if err != nil {
		return errors.Wrapf(err, "cannot delete pod %s", th.pod.Name)
	}

	log.Debugf("waiting for pod '%s' to stop", th.pod.Name)
	<-th.stopped
	log.Debugf("pod '%s' stopped", th.pod.Name)

	return nil
}

// Status returns the current task state in terms of RUNNING or TERMINATED.
func (th *kubernetesTaskHandle) Status() TaskState {
	if th.isTerminated() {
		return TERMINATED
	}

	return RUNNING
}

// ExitCode returns the exit code of the container running in the pod.
func (th *kubernetesTaskHandle) ExitCode() (int, error) {
	if !th.isTerminated() {
		return 0, errors.New("task is still running")
	}

	if th.exitCode == nil {
		return 0, errors.New("exit code unknown")
	}

	return *th.exitCode, nil
}

// Wait blocks until the pod terminates _or_ if timeout is provided, will exit ealier with
// false if the pod didn't terminate before the provided timeout.
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

// Clean closes file descriptors but leaves stdout and stderr files intact.
func (th *kubernetesTaskHandle) Clean() error {
	for _, f := range []*os.File{th.stderr, th.stdout} {
		if err := f.Close(); err != nil {
			return errors.Wrapf(err, "close on file %q failed", f.Name())
		}
	}
	return nil
}

// EraseOutput deletes the stdout and stderr files.
func (th *kubernetesTaskHandle) EraseOutput() error {
	outputDir, _ := path.Split(th.stderr.Name())
	if err := os.RemoveAll(outputDir); err != nil {
		return errors.Wrapf(err, "os.RemoveAll of directory %q failed", outputDir)
	}
	return nil
}

// Address returns the host IP where the pod was scheduled.
func (th *kubernetesTaskHandle) Address() string {
	// NOTE: Could be th.pod.Status.PodIP as well.
	return th.pod.Status.HostIP
}

// StdoutFile returns a file handle to the stdout file for the pod.
func (th *kubernetesTaskHandle) StdoutFile() (*os.File, error) {
	return th.stdout, nil
}

// StderrFile returns a file handle to the stderr file for the pod.
func (th *kubernetesTaskHandle) StderrFile() (*os.File, error) {
	return th.stderr, nil
}
