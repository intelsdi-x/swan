package executor

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
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

// KubernetesConfig describes the necessary information to connect to a Kubernetes cluster.
type KubernetesConfig struct {
	PodName        string
	Address        string
	Username       string
	Password       string
	CPURequest     int64
	CPULimit       int64
	Decorators     isolation.Decorators
	ContainerName  string
	ContainerImage string
	Namespace      string
}

// DefaultKubernetesConfig returns a KubernetesConfig object with safe defaults.
func DefaultKubernetesConfig() KubernetesConfig {
	return KubernetesConfig{
		PodName:        "swan",
		Address:        "127.0.0.1:8080",
		Username:       "",
		Password:       "",
		CPURequest:     0,
		CPULimit:       0,
		Decorators:     isolation.Decorators{},
		ContainerName:  "swan",
		ContainerImage: "jess/stress",
		Namespace:      api.NamespaceDefault,
	}
}

type kubernetes struct {
	config KubernetesConfig
	client *client.Client
}

// NewKubernetes returns an executor which lets the user run commands in pods in a
// kubernetes cluster.
func NewKubernetes(config KubernetesConfig) (Executor, error) {
	k8s := &kubernetes{
		config: config,
	}

	var err error
	k8s.client, err = client.New(&restclient.Config{
		Host:     config.Address,
		Username: config.Username,
		Password: config.Password,
	})

	if err != nil {
		return nil, errors.Wrapf(err, "can't initilize kubernetes client for host '%s'", config.Address)
	}

	return k8s, nil
}

// containerResources helper to create ResourceRequirments for the container.
func (k8s *kubernetes) containerResources() api.ResourceRequirements {
	resourceListLimits := api.ResourceList{}
	resourceListRequests := api.ResourceList{}
	if k8s.config.CPULimit > 0 {
		resourceListRequests[api.ResourceCPU] = *resource.NewQuantity(k8s.config.CPULimit, resource.DecimalSI)
	}
	if k8s.config.CPURequest > 0 {
		resourceListRequests[api.ResourceCPU] = *resource.NewQuantity(k8s.config.CPURequest, resource.DecimalSI)
	}
	return api.ResourceRequirements{
		Limits:   resourceListLimits,
		Requests: resourceListRequests,
	}
}

// Execute creates a pod and runs the provided command in it. When the command completes, the pod
// is stopped i.e. the container is not restarted automatically.
func (k8s *kubernetes) Execute(command string) (TaskHandle, error) {
	podsAPI := k8s.client.Pods(k8s.config.Namespace)
	command = k8s.config.Decorators.Decorate(command)

	// See http://kubernetes.io/docs/api-reference/v1/definitions/ for definition of the pod manifest.
	pod, err := podsAPI.Create(&api.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      k8s.config.PodName,
			Namespace: k8s.config.Namespace,
			Labels:    map[string]string{"name": k8s.config.PodName},
		},
		Spec: api.PodSpec{
			RestartPolicy: "Never",
			Containers: []api.Container{
				api.Container{
					Name:      k8s.config.ContainerName,
					Image:     k8s.config.ContainerImage,
					Command:   []string{"sh", "-c", command},
					Resources: k8s.containerResources(),
				},
			},
		},
	})

	if err != nil {
		return nil, errors.Wrapf(err, "cannot schedule pod %q with namespace %q",
			k8s.config.PodName, k8s.config.Namespace)
	}

	th := &kubernetesTaskHandle{
		command: command,
		podsAPI: podsAPI,
		pod:     pod,
	}

	th.watch()

	// NOTE: We should have timeout for the amount of time we want to wait for the pod to appear.
	select {
	case <-th.started:
		// Pod succesfully started.
	case <-th.stopped:
		// Look into exit state to determine if start up failed or completed immediately.
		exitCode, err := th.ExitCode()
		if err != nil {
			return th, errors.Errorf("failed to start pod: cannot get exit code")
		}

		if exitCode != 0 {
			return th, errors.Errorf("failed to start pod: failed with exit code %d", exitCode)
		}

		return th, nil
	}

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
	exitCode *int
}

func (th *kubernetesTaskHandle) watch() error {
	selectorRaw := fmt.Sprintf("name=%s", th.pod.Name)
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		return errors.Wrapf(err, "cannot create an selector")
	}

	// Prepare events watcher.
	watcher, err := th.podsAPI.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "cannot create watcher over selector")
	}

	th.started = make(chan struct{})
	th.stopped = make(chan struct{})

	go func() {
		var onceStarted sync.Once
		started := func(pod *api.Pod) {
			if api.IsPodReady(pod) {
				onceStarted.Do(func() {
					close(th.started)
				})
			}
		}

		var onceStopped sync.Once
		terminated := func(pod *api.Pod) {
			onceStopped.Do(func() {
				exitCode := 1

				// Look for an exit status from the container.
				// If more than one container is present, the last takes precedence.
				for _, status := range pod.Status.ContainerStatuses {
					if status.State.Terminated == nil {
						continue
					}

					exitCode = int(status.State.Terminated.ExitCode)
				}

				// NOTE: We may want to have a lock/read barrier on the exit code to ensure consistent read
				// in ExitCode().
				th.exitCode = &exitCode

				close(th.stopped)
			})

			// Try to delete the failed pod to avoid conflicts and having to call Stop()
			// after the stopped channel has been closed.
			var GracePeriodSeconds int64
			th.podsAPI.Delete(th.pod.Name, &api.DeleteOptions{
				GracePeriodSeconds: &GracePeriodSeconds,
			})
		}

		for event := range watcher.ResultChan() {
			pod, ok := event.Object.(*api.Pod)
			if !ok {
				continue
			}

			// Update with latest status.
			// NOTE: May want to make this synchronized.
			th.pod = pod

			switch event.Type {
			case watch.Added, watch.Modified:
				switch pod.Status.Phase {
				case api.PodPending:
					// Noop for now.
				case api.PodRunning:
					started(pod)
				case api.PodFailed, api.PodSucceeded:
					terminated(pod)
					return
				default:
					log.Debugf("unknown phase '%d' for pod %q", pod.Status.Phase, pod.Name)
				}

			case watch.Deleted:
				// Pod phase will still be 'running', so we disregard the phase at this point.
				terminated(pod)
				return
			default:
				log.Debugf("unknown event type")
			}
		}
	}()

	return nil
}

// NOTE: That setupLogs can only be called when the pod is running i.e. wait until the started
// channel has been closed by watch().
func (th *kubernetesTaskHandle) setupLogs() error {
	// Wire up logs to task handle stdout.
	logStream, err := th.podsAPI.GetLogs(th.pod.Name, &api.PodLogOptions{
		Container: th.pod.Spec.Containers[0].Name,
	}).Stream()
	if err != nil {
		return errors.Wrapf(err, "cannot create a stream")
	}

	// Prepare local files
	stdoutFile, stderrFile, err := createExecutorOutputFiles(th.command, "local")
	if err != nil {
		return errors.Wrapf(err, "cannot create output files for pod %q", th.pod.Name)
	}
	log.Debugf("created temporary files stdout path: %q stderr path: %q",
		stdoutFile.Name(), stderrFile.Name())

	th.stdout = stdoutFile
	// NOTE: As logs are unified in one stream in Kubernetes, we only write it to stdout.
	// Therefore, stderr will always be empty.
	th.stderr = stderrFile

	go func() {
		_, err := io.Copy(stdoutFile, logStream)
		if err != nil {
			log.Debugf("Failed to copy container log stream to task output: %s", err.Error())
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

	log.Debugf("deleting pod %q", th.pod.Name)

	var GracePeriodSeconds int64
	err := th.podsAPI.Delete(th.pod.Name, &api.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	})
	if err != nil {
		return errors.Wrapf(err, "cannot delete pod %q", th.pod.Name)
	}

	log.Debugf("waiting for pod %q to stop", th.pod.Name)
	<-th.stopped
	log.Debugf("pod %q stopped", th.pod.Name)

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
		return errors.Wrapf(err, "cannot remove directory %q", outputDir)
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
