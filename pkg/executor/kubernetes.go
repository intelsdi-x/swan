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
	"github.com/intelsdi-x/athena/pkg/utils/err_collection"
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
		return nil, err
	}

	return &api.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      podName,
			Namespace: k8s.config.Namespace,
			Labels:    map[string]string{"name": podName},
		},
		Spec: api.PodSpec{
			RestartPolicy:   "Never",
			SecurityContext: &api.PodSecurityContext{HostNetwork: k8s.config.HostNetwork},
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

	// See http://kubernetes.io/docs/api-reference/v1/definitions/ for definition of the pod manifest.
	podManifest, err := k8s.newPod(command)
	if err != nil {
		return nil, errors.Wrapf(err, "cannot create pod manifest")
	}

	pod, err := podsAPI.Create(podManifest)

	log.Debugf("pod.Spec = %+v\n", pod.Spec)
	log.Debugf("pod %q QoS class:", qos.GetPodQOS(pod))

	if err != nil {
		return nil, errors.Wrapf(err, "cannot schedule pod %q with namespace %q",
			k8s.config.PodName, k8s.config.Namespace)
	}

	taskHandle := &kubernetesTaskHandle{
		command:       command,
		podsAPI:       podsAPI,
		pod:           pod,
		podMutex:      &sync.Mutex{},
		exitCodeMutex: &sync.Mutex{},
		outputMutex:   &sync.Mutex{},
	}

	taskHandle.watch(k8s.config.LaunchTimeout)

	select {
	case <-taskHandle.started:
		// Pod succesfully started.
	case <-taskHandle.stopped:
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
	command   string
	stopped   chan struct{}
	started   chan struct{}
	logsReady chan struct{}

	// NOTE: Access to the pod structure must be done through setPod()
	// and getPod() to avoid data races between caller and watcher routine.
	pod      *api.Pod
	podMutex *sync.Mutex

	// NOTE: Access to stdout and stderr must be done through setFileHandles()
	// and getFileHandles() to avoid data races between caller and watcher
	// routine.
	stdout      *os.File
	stderr      *os.File // Kubernetes does not support separation of stderr & stdout, so this file will be empty
	logdir      string
	outputMutex *sync.Mutex

	// NOTE: Access to the exit code must be done through setExitCode()
	// and getExitCode() to avoid data races between caller and watcher routine.
	exitCode      *int
	exitCodeMutex *sync.Mutex
}

// getPod provides a thread safe way to get the most recent pod structure.
func (th *kubernetesTaskHandle) getPod() *api.Pod {
	th.podMutex.Lock()
	defer th.podMutex.Unlock()

	return th.pod
}

// setPod provides a thread safe way to set the active pod structure.
func (th *kubernetesTaskHandle) setPod(pod *api.Pod) {
	th.podMutex.Lock()
	defer th.podMutex.Unlock()

	th.pod = pod
}

// getExitCode provides a thread safe way to get the most recent exit code.
func (th *kubernetesTaskHandle) getExitCode() *int {
	th.exitCodeMutex.Lock()
	defer th.exitCodeMutex.Unlock()

	return th.exitCode
}

// setExitCode provides a thread safe way to set the active exit code.
func (th *kubernetesTaskHandle) setExitCode(exitCode *int) {
	th.exitCodeMutex.Lock()
	defer th.exitCodeMutex.Unlock()

	th.exitCode = exitCode
}

// setFileHandles provides a thread safe way to set the file descriptors for
// the stdout and stderr files.
func (th *kubernetesTaskHandle) setFileHandles(stdoutFile *os.File, stderrFile *os.File) {
	th.outputMutex.Lock()
	defer th.outputMutex.Unlock()

	th.stdout = stdoutFile
	th.stderr = stderrFile

	outputDir, _ := path.Split(th.stdout.Name())
	th.logdir = outputDir
}

// setFileHandles provides a thread safe way to get the file descriptors for
// the stdout and stderr files.
func (th *kubernetesTaskHandle) getFileHandles() (*os.File, *os.File) {
	th.outputMutex.Lock()
	defer th.outputMutex.Unlock()

	return th.stdout, th.stderr
}

func (th *kubernetesTaskHandle) setupLogs(pod *api.Pod) {
	log.Debugf("Setting up logs for pod %q", pod.Name)

	// Wire up logs to task handle stdout.
	logStream, err := th.podsAPI.GetLogs(pod.Name, &api.PodLogOptions{}).Stream()
	if err != nil {
		log.Debug("cannot create a stream: %s", err.Error())
		return
	}

	// Prepare local files
	stdoutFile, stderrFile, err := createExecutorOutputFiles(th.command, "kubernetes")

	if err != nil {
		log.Debug("cannot create output files for pod %q: %s", pod.Name, err.Error())
		return
	}

	log.Debugf("created temporary files stdout path: %q stderr path: %q",
		stdoutFile.Name(), stderrFile.Name())

	go func() {
		th.outputMutex.Lock()
		defer th.outputMutex.Unlock()
		_, err := io.Copy(stdoutFile, logStream)
		if err != nil {
			log.Debugf("Failed to copy container log stream to task output: %s", err.Error())
		}

		stdoutFile.Sync()
	}()

	// Set file handles in task handle in a synchronized manner.
	th.setFileHandles(stdoutFile, stderrFile)

	close(th.logsReady)
}

func (th *kubernetesTaskHandle) watch(timeout time.Duration) error {
	pod := th.getPod()

	selectorRaw := fmt.Sprintf("name=%s", pod.Name)
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		return errors.Wrapf(err, "cannot create selector %q", selector)
	}

	// Prepare events watcher.
	watcher, err := th.podsAPI.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "cannot create watcher over selector %q", selector)
	}

	th.started = make(chan struct{})
	th.stopped = make(chan struct{})
	th.logsReady = make(chan struct{})

	go func() {
		var onceSetupLogs, onceStopped, onceStarted sync.Once
		var hasBeenRunning bool

		started := func(pod *api.Pod) {
			if api.IsPodReady(pod) {
				onceStarted.Do(func() {
					close(th.started)
				})
			}
		}

		terminated := func(pod *api.Pod) {
			onceStopped.Do(func() {
				exitCode := -1

				// Look for an exit status from the container.
				// If more than one container is present, the last takes precedence.
				for _, status := range pod.Status.ContainerStatuses {
					if status.State.Terminated == nil {
						continue
					}

					exitCode = int(status.State.Terminated.ExitCode)
				}

				th.setExitCode(&exitCode)

				th.stdout.Sync()
				th.stderr.Sync()

				close(th.stopped)
			})

			// Try to delete the failed pod to avoid conflicts and having to call Stop()
			// after the stopped channel has been closed.
			var GracePeriodSeconds int64
			th.podsAPI.Delete(pod.Name, &api.DeleteOptions{
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

				th.setPod(pod)

				switch event.Type {
				case watch.Added, watch.Modified:
					switch pod.Status.Phase {
					case api.PodPending:
					// Noop for now.
					case api.PodRunning:
						onceSetupLogs.Do(func() {
							th.setupLogs(pod)
						})

						started(pod)
						hasBeenRunning = true
					case api.PodFailed, api.PodSucceeded:
						onceSetupLogs.Do(func() {
							th.setupLogs(pod)
						})

						terminated(pod)
						hasBeenRunning = true
						return
					case api.PodUnknown:
						log.Warnf("Pod %q with command %q is in unknown phase. "+
							"Probably state of the pod could not be obtained, "+
							"typically due to an error in communicating with the host of the pod", pod.Name, th.command)
					default:
						log.Warnf("Unhandled pod phase event %q for pod %q", pod.Status.Phase, pod.Name)
					}
				case watch.Deleted:
					// Pod phase will still be 'running', so we disregard the phase at this point.
					onceSetupLogs.Do(func() {
						th.setupLogs(pod)
					})

					terminated(pod)
					return
				case watch.Error:
					log.Errorf("Kubernetes pod error event: %v", event.Object)
				default:
					log.Warnf("Unhandled event type: %v", event.Type)
				}
			case <-timeoutChannel:
				// If task has been running then we need to ignore timeout
				if hasBeenRunning {
					continue
				}
				log.Errorf("Timeout occured afrer %f seconds. Pod %s has not been created.", timeout.Seconds(), th.getPod().Name)
				onceSetupLogs.Do(func() {
					th.setupLogs(pod)
				})
				terminated(pod)
				th.setExitCode(nil)
				return
			}
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

	pod := th.getPod()

	log.Debugf("deleting pod %q", pod.Name)

	var GracePeriodSeconds int64
	err := th.podsAPI.Delete(pod.Name, &api.DeleteOptions{
		GracePeriodSeconds: &GracePeriodSeconds,
	})
	if err != nil {
		return errors.Wrapf(err, "cannot delete pod %q", pod.Name)
	}

	log.Debugf("waiting for pod %q to stop", pod.Name)
	<-th.stopped
	log.Debugf("pod %q stopped", pod.Name)

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

	exitCode := th.getExitCode()

	if exitCode == nil {
		return 0, errors.New("exit code unknown")
	}

	return *exitCode, nil
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
	th.outputMutex.Lock()
	defer th.outputMutex.Unlock()
	var errs errcollection.ErrorCollection
	for _, f := range []*os.File{th.stderr, th.stdout} {
		if f != nil {
			if err := f.Close(); err != nil {
				errs.Add(errors.Wrapf(err, "close of file %q failed", f.Name()))
			}
		}
	}

	return errs.GetErrIfAny()
}

// EraseOutput deletes the stdout and stderr files.
// NOTE: EraseOutput will block until output directory is available.
func (th *kubernetesTaskHandle) EraseOutput() error {
	<-th.logsReady

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
	// NOTE: Could be pod.Status.PodIP as well.
	pod := th.getPod()
	return pod.Status.HostIP
}

// StdoutFile returns a file handle to the stdout file for the pod.
// NOTE: StdoutFile will block until stdout file is available.
func (th *kubernetesTaskHandle) StdoutFile() (*os.File, error) {
	<-th.logsReady

	stdout, _ := th.getFileHandles()

	if stdout == nil {
		return nil, errors.New("stdout file has been already closed or it is not created yet")
	}

	// We need to create yet another file in order to avoid races and issues with seeking.
	// stdout will point at the very end of file as it is being used for writing.
	file, err := os.Open(stdout.Name())
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open stdout file")
	}

	return file, nil
}

// StderrFile returns a file handle to the stderr file for the pod.
// NOTE: StderrFile will block until stderr file is available.
func (th *kubernetesTaskHandle) StderrFile() (*os.File, error) {
	<-th.logsReady

	_, stderr := th.getFileHandles()

	if stderr == nil {
		return nil, errors.New("stderr file has been already closed or it is not created yet")
	}

	// See comments to StdoutFile()
	file, err := os.Open(stderr.Name())
	if err != nil {
		return nil, errors.Wrap(err, "Unable to open stderr file")
	}

	return file, nil
}
