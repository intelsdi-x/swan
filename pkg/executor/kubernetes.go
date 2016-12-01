package executor

import (
	"fmt"
	"io"
	"os"
	"path"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/isolation"
	"github.com/nu7hatch/gouuid"
	"github.com/pkg/errors"
	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/resource"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/kubelet/qos"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

const (
	defaultContainerImage = "jess/stress" // TODO: replace with "centos_swan_image" when available.

	dockerSockPath = "unix:///var/run/docker.sock"
)

// KubernetesConfig describes the necessary information to connect to a Kubernetes cluster.
type KubernetesConfig struct {
	// PodName vs PodNamePrefix:
	// - PodName(deprecated; by default is empty string) - when this field is
	// configured, Kubernetes executor is using it as a Pod name. Kubernetes
	// doesn't support spawning pods with same name, so this field shouldn't
	// be in use.
	// - PodNamePrefix(by default is "swan") - If PodName field is not
	// configured, this field is used as a prefix for random generated Pod
	// name.
	PodName        string
	PodNamePrefix  string
	Address        string
	Username       string
	Password       string
	CPURequest     int64
	CPULimit       int64
	MemoryRequest  int64
	MemoryLimit    int64
	Decorators     isolation.Decorators
	ContainerName  string
	ContainerImage string
	Namespace      string
	Privileged     bool
	HostNetwork    bool
	LaunchTimeout  time.Duration
}

// LaunchTimedOutError is the error type returned when launching new pods exceed
// the timeout value defined in kubernetes.Config.LaunchTimeout.
type LaunchTimedOutError struct {
	errorMessage string
}

// Error is one method needed to implement the error interface. Here, we just return
// an the error message.
func (err *LaunchTimedOutError) Error() string {
	return err.errorMessage
}

// DefaultKubernetesConfig returns a KubernetesConfig object with safe defaults.
func DefaultKubernetesConfig() KubernetesConfig {
	return KubernetesConfig{
		PodName:        "",
		PodNamePrefix:  "swan",
		Address:        "127.0.0.1:8080",
		Username:       "",
		Password:       "",
		CPURequest:     0,
		CPULimit:       0,
		MemoryRequest:  0,
		MemoryLimit:    0,
		Decorators:     isolation.Decorators{},
		ContainerName:  "swan",
		ContainerImage: defaultContainerImage,
		Namespace:      api.NamespaceDefault,
		Privileged:     false,
		HostNetwork:    false,
		LaunchTimeout:  0,
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

	// requests
	resourceListRequests := api.ResourceList{}
	if k8s.config.CPURequest > 0 {
		resourceListRequests[api.ResourceCPU] = *resource.NewMilliQuantity(k8s.config.CPURequest, resource.DecimalSI)
	}
	if k8s.config.MemoryRequest > 0 {
		resourceListRequests[api.ResourceMemory] = *resource.NewQuantity(k8s.config.MemoryRequest, resource.DecimalSI)
	}

	// limits
	resourceListLimits := api.ResourceList{}
	if k8s.config.CPULimit > 0 {
		resourceListLimits[api.ResourceCPU] = *resource.NewMilliQuantity(k8s.config.CPULimit, resource.DecimalSI)
	}
	if k8s.config.MemoryLimit > 0 {
		resourceListLimits[api.ResourceMemory] = *resource.NewQuantity(k8s.config.MemoryLimit, resource.DecimalSI)
	}

	return api.ResourceRequirements{
		Requests: resourceListRequests,
		Limits:   resourceListLimits,
	}
}

// Name returns user-friendly name of executor.
func (k8s *kubernetes) Name() string {
	return "Kubernetes Executor"
}

// generatePodName is generating pods based on KubernetesConfig struct.
// If KubernetesConfig has got non-empty PodName field then this field is
// used as a Pod name (even if selected Pod name is already in use - in
// this case pod spawning should fail).
// It returns string as a Pod name which should be used or error if cannot
// generate random suffix.
func (k8s *kubernetes) generatePodName() (string, error) {
	if k8s.config.PodName != "" {
		return k8s.config.PodName, nil
	}

	uuid, err := uuid.NewV4()
	if err != nil {
		return "", errors.Wrapf(err, "cannot generate suffix UUID")
	}

	return fmt.Sprintf("%s-%x", k8s.config.PodNamePrefix, uuid.String())[:60], nil
}

// newPod is a helper to build in-memory struture representing pod
// before sending it as request to API server. It can returns also
// error if cannot generate Pod name.
func (k8s *kubernetes) newPod(command string) (*api.Pod, error) {

	resources := k8s.containerResources()
	podName, err := k8s.generatePodName()
	if err != nil {
		return nil, errors.Wrapf(err, "cannot generate pod name")
	}

	var zero int64
	return &api.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      podName,
			Namespace: k8s.config.Namespace,
			Labels:    map[string]string{"name": podName},
		},
		Spec: api.PodSpec{
			RestartPolicy:                 "Never",
			SecurityContext:               &api.PodSecurityContext{HostNetwork: k8s.config.HostNetwork},
			TerminationGracePeriodSeconds: &zero,
			Containers: []api.Container{
				api.Container{
					Name:            k8s.config.ContainerName,
					Image:           k8s.config.ContainerImage,
					Command:         []string{"sh", "-c", command},
					Resources:       resources,
					ImagePullPolicy: api.PullIfNotPresent, // Default because swan image is not published yet.
					SecurityContext: &api.SecurityContext{Privileged: &k8s.config.Privileged},
				},
			},
		},
	}, nil
}

// Execute creates a pod and runs the provided command in it. When the command completes, the pod
// is stopped i.e. the container is not restarted automatically.
func (k8s *kubernetes) Execute(command string) (TaskHandle, error) {
	podsAPI := k8s.client.Pods(k8s.config.Namespace)
	command = k8s.config.Decorators.Decorate(command)

	// This is a workaround for kubernetes #31446 & SCE-883.
	// Make sure that at least one line of text is outputed from pod, to unblock .GetLogs() on apiserver call
	// with streamed response (when follow=true). Check SCE-883 for details or kubernetes #31446 issue.
	// https://github.com/kubernetes/kubernetes/pull/31446
	command = "echo;" + command

	// See http://kubernetes.io/docs/api-reference/v1/definitions/ for definition of the pod manifest.
	podManifest, err := k8s.newPod(command)
	if err != nil {
		log.Errorf("K8s executor: cannot create pod manifest")
		return nil, errors.Wrapf(err, "cannot create pod manifest")
	}

	pod, err := podsAPI.Create(podManifest)
	if err != nil {
		log.Errorf("K8s executor: cannot schedule pod %q with namespace %q", k8s.config.PodName, k8s.config.Namespace)
		return nil, errors.Wrapf(err, "cannot schedule pod %q with namespace %q",
			k8s.config.PodName, k8s.config.Namespace)
	}

	log.Debugf("K8s executor: pod specification = %+v\n", pod.Spec)
	log.Debugf("K8s executor: pod %q QoS class %q", pod.Name, qos.GetPodQOS(pod))

	taskWatcher := &kubernetesWatcher{
		podsAPI: podsAPI,
		pod:     pod,

		stopped:         make(chan struct{}, 1),
		started:         make(chan struct{}, 1),
		logsReady:       make(chan struct{}, 1),
		outputCopied:    make(chan struct{}, 1),
		exitCodeChannel: make(chan *int, 1),

		command: command,
	}

	taskHandle, err := taskWatcher.watch(k8s.config.LaunchTimeout)
	if err != nil {
		log.Errorf("K8s executor: cannot create task on pod %q", pod.Name)
		return nil, errors.Wrapf(err, "cannot create task on pod %q", pod.Name)
	}

	select {
	case <-taskWatcher.started:
		// Pod succesfully started.
	case <-taskWatcher.stopped:
		// Look into exit state to determine if start up failed or completed immediately.
		// TODO(skonefal): We don't have stdout & stderr when pod fails.
		exitCode, err := taskHandle.ExitCode()
		if err != nil || exitCode != 0 {
			LogUnsucessfulExecution(command, k8s.Name(), taskHandle)
		} else {
			LogSuccessfulExecution(command, k8s.Name(), taskHandle)
		}
	}

	return taskHandle, nil
}

// kubernetesTaskHandle implements the TaskHandle interface
type kubernetesTaskHandle struct {
	podsAPI   client.PodInterface
	stopped   chan struct{}
	started   chan struct{}
	logsReady chan struct{}

	stdout string
	stderr string // Kubernetes does not support separation of stderr & stdout, so this file will be empty
	logdir string

	// NOTE: Access to the exit code must be done through setExitCode()
	// and getExitCode() to avoid data races between caller and watcher routine.
	exitCode        *int
	exitCodeChannel chan *int

	podName   string
	podHostIP string
}

func (th *kubernetesTaskHandle) setFileNames(stdoutFile, stderrFile string) {
	th.stdout = stdoutFile
	th.stderr = stderrFile

	outputDir, _ := path.Split(th.stdout)
	th.logdir = outputDir
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

	log.Debugf("K8s task handle: deleting pod %q", th.podName)

	// Setting GP to zero will erase pod from API server and won't wait for it to exit.
	gracePeriodSeconds := int64(1)
	err := th.podsAPI.Delete(th.podName, &api.DeleteOptions{
		GracePeriodSeconds: &gracePeriodSeconds,
	})
	if err != nil {
		log.Errorf("K8s task handle: cannot delete pod %q", th.podName)
		return errors.Wrapf(err, "cannot delete pod %q", th.podName)
	}

	log.Debugf("K8s task handle: waiting for pod %q to stop", th.podName)
	<-th.stopped
	log.Debugf("K8s task handle: pod %q stopped", th.podName)

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

	select {
	case exitCode := <-th.exitCodeChannel:
		log.Debugf("K8s task handle: received exit code: %#v", th.exitCode)
		if th.exitCode == nil {
			log.Debug("K8s task handle: setting exit code")
			th.exitCode = exitCode
		}
	default:
		log.Debug("K8s task handle: no exit code received")
	}

	if th.exitCode == nil {
		log.Error("K8s task handle: exit code is unknown")
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

// Clean implements TaskHandle interface but is a no-action (as kubernetesWatcher manages file descriptors).
func (th *kubernetesTaskHandle) Clean() error {
	return nil
}

// EraseOutput deletes the stdout and stderr files.
// NOTE: EraseOutput will block until output directory is available.
func (th *kubernetesTaskHandle) EraseOutput() error {
	log.Debug("K8s task handle: waiting for logsReady channel to be closed")
	<-th.logsReady
	log.Debug("K8s task handle: logs ready channel must have been closed")

	directory, err := os.Lstat(th.logdir)
	if err == nil && directory.IsDir() {
		if err := os.RemoveAll(th.logdir); err != nil {
			return errors.Wrapf(err, "cannot remove directory %q", th.logdir)
		}
	} else {
		return errors.Wrapf(err, "cannot remove directory %q", th.logdir)
	}

	return nil
}

// Address returns the host IP where the pod was scheduled.
func (th *kubernetesTaskHandle) Address() string {
	return th.podHostIP
}

// StdoutFile returns a file handle to the stdout file for the pod.
// NOTE: StdoutFile will block until stdout file is available.
func (th *kubernetesTaskHandle) StdoutFile() (*os.File, error) {
	<-th.logsReady
	if th.stdout == "" {
		return nil, errors.New("stdout file has been already closed or it is not created yet")
	}

	file, err := os.Open(th.stdout)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open stdout file")
	}

	return file, nil
}

// StderrFile returns a file handle to the stderr file for the pod.
// NOTE: StderrFile will block until stderr file is available.
func (th *kubernetesTaskHandle) StderrFile() (*os.File, error) {
	<-th.logsReady
	if th.stderr == "" {
		return nil, errors.New("stderr file has been already closed or it is not created yet")
	}

	file, err := os.Open(th.stderr)
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open stderr file")
	}

	return file, nil
}

type kubernetesWatcher struct {
	podsAPI client.PodInterface
	pod     *api.Pod

	stdout *os.File
	stderr *os.File // Kubernetes does not support separation of stderr & stdout, so this file will be empty

	stopped         chan struct{}
	started         chan struct{}
	logsReady       chan struct{}
	outputCopied    chan struct{}
	exitCodeChannel chan *int

	taskHandle *kubernetesTaskHandle

	command string
}

// watch creates instance of TaskHandle and is responsible for keeping it in-sync with k8s cluster
func (kw *kubernetesWatcher) watch(timeout time.Duration) (TaskHandle, error) {
	kw.taskHandle = &kubernetesTaskHandle{
		podName:         kw.pod.Name,
		podHostIP:       kw.pod.Status.HostIP,
		stopped:         kw.stopped,
		started:         kw.started,
		logsReady:       kw.logsReady,
		podsAPI:         kw.podsAPI,
		exitCodeChannel: kw.exitCodeChannel,
	}

	selectorRaw := fmt.Sprintf("name=%s", kw.pod.Name)
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create selector %q", selector)
	}

	// Prepare events watcher.
	watcher, err := kw.podsAPI.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create watcher over selector %q", selector)
	}

	go func() {
		var onceSetupLogs, onceStopped, onceStarted sync.Once
		var hasBeenRunning bool

		started := func(pod *api.Pod) {
			if api.IsPodReady(pod) {
				onceStarted.Do(func() {
					log.Debug("K8s task watcher: pod has been started succesfully")
					close(kw.started)
				})
			}
		}

		// sendExitCode sends exit code to task handle.
		sendExitCode := func(exitCode *int) {
			kw.exitCodeChannel <- exitCode
			log.Debug("K8s task watcher: exit code sent")
			close(kw.exitCodeChannel)
		}

		// setExitCode retrieves exit code from the cluster and sends it to task handle.
		setExitCode := func(pod *api.Pod) {
			exitCode := -1

			// Look for an exit status from the container.
			// If more than one container is present, the last takes precedence.
			for _, status := range pod.Status.ContainerStatuses {
				if status.State.Terminated == nil {
					continue
				}

				exitCode = int(status.State.Terminated.ExitCode)
			}
			log.Debugf("K8s task watcher: exit code retrieved: %d", exitCode)
			sendExitCode(&exitCode)
		}

		// cleanup makes sure that output has already been copied, closes file descriptors.
		cleanup := func() {
			log.Debug("K8s task watcher: waiting for output to be copied")
			<-kw.outputCopied
			log.Debug("K8s task watcher: closing files' descriptors")
			kw.stdout.Close()
			kw.stderr.Close()
			log.Debug("K8s task watcher: task stopped succesfully")
			close(kw.stopped)
		}

		// delete removes the pod from the cluster on best effort basis.
		delete := func(pod *api.Pod) {
			log.Debugf("K8s task watcher: attempting to delete pod %q", pod.Name)
			// Try to delete the failed pod to avoid conflicts and having to call Stop()
			// after the stopped channel has been closed.
			var GracePeriodSeconds int64
			kw.podsAPI.Delete(pod.Name, &api.DeleteOptions{
				GracePeriodSeconds: &GracePeriodSeconds,
			})
		}

		var timeoutChannel <-chan time.Time
		if timeout != 0 {
			timeoutChannel = time.After(timeout)
		}
		for {
			select {
			case event := <-watcher.ResultChan():
				pod, ok := event.Object.(*api.Pod)
				if !ok {
					continue
				}

				switch event.Type {
				case watch.Added, watch.Modified:
					switch pod.Status.Phase {
					case api.PodPending:
						log.Debug("K8s task watcher: event received: api.PodPending")
					// Noop for now.
					case api.PodRunning:
						log.Debug("K8s task watcher: event received: api.PodRunning")
						onceSetupLogs.Do(func() {
							kw.setupLogs(pod)
						})

						started(pod)
						hasBeenRunning = true
					case api.PodFailed, api.PodSucceeded:
						log.Debug("K8s task watcher: event received: api.PodFailed or api.PodSucceeded")
						onceSetupLogs.Do(func() {
							kw.setupLogs(pod)
						})

						onceStopped.Do(func() {
							setExitCode(pod)
							cleanup()
							delete(pod)
						})
						hasBeenRunning = true
						return
					case api.PodUnknown:
						log.Warnf("K8s task watcher: pod %q with command %q is in unknown phase. "+
							"Probably state of the pod could not be obtained, "+
							"typically due to an error in communicating with the host of the pod", pod.Name, kw.command)
					default:
						log.Warnf("K8s task watcher: unhandled pod phase event %q for pod %q", pod.Status.Phase, pod.Name)
					}
				case watch.Deleted:
					// Pod phase will still be 'running', so we disregard the phase at this point.
					log.Debug("K8s task watcher: event received: watch.Deleted")
					onceSetupLogs.Do(func() {
						kw.setupLogs(pod)
					})

					onceStopped.Do(func() {
						setExitCode(pod)
						cleanup()
					})

					return
				case watch.Error:
					log.Errorf("K8s task watcher: kubernetes pod error event: %v", event.Object)
				default:
					log.Warnf("K8s task watcher: unhandled event type: %v", event.Type)
				}
			case <-timeoutChannel:
				// If task has been running then we need to ignore timeout
				if hasBeenRunning {
					continue
				}
				log.Errorf("K8s task watcher: timeout occured afrer %f seconds. Pod %s has not been created.", timeout.Seconds(), kw.pod.Name)
				onceSetupLogs.Do(func() {
					kw.setupLogs(kw.pod)
				})

				onceStopped.Do(func() {
					sendExitCode(nil)
					cleanup()
					delete(kw.pod)
				})

				return
			}
		}
	}()

	return kw.taskHandle, nil
}

// setupLogs creates log files and initializes goroutine that copies stream from k8s
func (kw *kubernetesWatcher) setupLogs(pod *api.Pod) error {
	log.Debugf("K8s task watcher: setting up logs for pod %q", pod.Name)

	// Wire up logs to task handle stdout.
	logStream, err := kw.podsAPI.GetLogs(pod.Name, &api.PodLogOptions{Follow: true}).Stream()
	if err != nil {
		log.Errorf("K8s task watcher: cannot create a stream: %s", err.Error())
		return errors.Wrap(err, "cannot create a stream")
	}

	// Prepare local files
	stdoutFile, stderrFile, err := createExecutorOutputFiles(kw.command, "kubernetes")

	if err != nil {
		log.Errorf("K8s task watcher: cannot create output files for pod %q: %s", pod.Name, err.Error())
		return errors.Wrapf(err, "cannot create output files for pod %q", pod.Name)
	}

	log.Debugf("K8s task watcher: created temporary files stdout path: %q stderr path: %q",
		stdoutFile.Name(), stderrFile.Name())

	go func() {
		_, err := io.Copy(stdoutFile, logStream)
		log.Debug("K8s task watcher: output copied succesfully")
		close(kw.outputCopied)
		if err != nil {
			log.Debugf("K8s task watcher: failed to copy container log stream to task output: %s", err.Error())
		}

		stdoutFile.Sync()
	}()

	// Set file handles in task handle in a synchronized manner.
	kw.taskHandle.setFileNames(stdoutFile.Name(), stderrFile.Name())
	log.Debug("K8s task watcher: logs for %q have been set up", pod.Name)
	close(kw.logsReady)

	return nil
}
