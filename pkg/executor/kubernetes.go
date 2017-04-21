/*
Kubernetes executor under the hood is using these are three components:
- kubernetesExecutor:
	- talks to kubernetes cluster to create new pod,
	- starts watcher (by calling Execute)
	- runs in main goroutine,
	- gives back control to user by returning newly created taskHandle
- watcher:
	- creates "log copier" via setupLogs() funcion invocation
	- responsible for monitoring state of Pod and passing to information to taskHandle,
	- also in case of failure or part of cleaning up or when asked directly by taskHandle - deletes pod,
- copier:
	- Resides in setupLogs() function
	- It is responsible for copying logs from streamed kubernetes response
	- Closes logsCopyFinished channel when logs finishes streaming or failed to create stream


Actually pod transitions by those phases which maps to those handles:
- Pending: do nothing, just log,
- Running and Ready: calls whenPodReady() handler
- Success or Failed: calls whenPodFinishes() handler and more importantly deletePod() action.
- Deleted: whenPodDeleted - to signal taskHandler


Note:
- whenPodFinished always calls whenPodReady - if pod finished it had to be running before - to setup logs,
- whenPodFinished may be skipped at all then outputCopied never closed

Communication with taskHandle is through taskWatcher events:
- whenPodReady it signals to "started" to taskHandle (Execute method is unblocked)
- whenPodFinished it just examines the state of pod and stores it exit code you can read taskHandle.ExitCode,
- whenPodDeleted signals by closing "stopped" channel - taskHandle.Status() return terminated,


The logs are prepared ("copier" goroutine) with setupLogs() and happens at every occasion (but only once).

Every "handler" (when*) and every action like setupLogs() and deletePod() can happen only once.

Additionally every handler or action is protected by sync.Once to make sure that is run only once.
*/

package executor

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/intelsdi-x/swan/pkg/k8sports"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	"github.com/pkg/errors"
	"k8s.io/client-go/1.5/kubernetes"
	corev1 "k8s.io/client-go/1.5/kubernetes/typed/core/v1"
	"k8s.io/client-go/1.5/pkg/api"
	"k8s.io/client-go/1.5/pkg/api/resource"
	"k8s.io/client-go/1.5/pkg/api/unversioned"
	"k8s.io/client-go/1.5/pkg/api/v1"
	"k8s.io/client-go/1.5/pkg/labels"
	"k8s.io/client-go/1.5/pkg/watch"
	"k8s.io/client-go/1.5/rest"
	"k8s.io/client-go/1.5/tools/clientcmd"
)

const (
	defaultContainerImage = "centos_swan_image"
)

var (
	kubeconfigFlag                = conf.NewStringFlag("kubernetes_kubeconfig", "(optional) Absolute path to the kubeconfig file. Overrides pod configuration passed through flags. ", "")
	kubernetesPrivilegedPodsFlag  = conf.NewBoolFlag("kubernetes_privileged_pods", "Kubernetes containers will be run as privileged.", false)
	kubernetesPodLunchTimeoutFlag = conf.NewDurationFlag("kubernetes_pod_launch_timeout", "Kubernetes Pod launch timeout.", 30*time.Second)
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
	NodeName       string
	Address        string
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
		NodeName:       "",
		PodName:        "",
		PodNamePrefix:  "swan",
		Address:        "127.0.0.1:8080",
		CPURequest:     0,
		CPULimit:       0,
		MemoryRequest:  0,
		MemoryLimit:    0,
		Decorators:     isolation.Decorators{},
		ContainerName:  "swan",
		ContainerImage: defaultContainerImage,
		Namespace:      v1.NamespaceDefault,
		Privileged:     kubernetesPrivilegedPodsFlag.Value(),
		HostNetwork:    false,
		LaunchTimeout:  kubernetesPodLunchTimeoutFlag.Value(),
	}
}

type k8s struct {
	config    KubernetesConfig
	clientset *kubernetes.Clientset
}

// NewKubernetes returns an executor which lets the user run commands in pods in a
// kubernetes cluster.
func NewKubernetes(config KubernetesConfig) (Executor, error) {
	k8s := &k8s{
		config: config,
	}

	var err error
	kubeConfigPath := kubeconfigFlag.Value()
	if kubeConfigPath == "" {
		k8s.clientset, err = kubernetes.NewForConfig(&rest.Config{
			Host: config.Address,
		})
	} else {
		var kubeconfig *rest.Config
		kubeconfig, err = clientcmd.BuildConfigFromFlags("", kubeConfigPath)
		k8s.clientset, err = kubernetes.NewForConfig(kubeconfig)
	}

	if err != nil {
		return nil, errors.Wrapf(err, "can't initilize kubernetes clientset for host '%s'", config.Address)
	}

	return k8s, nil
}

// containerResources helper to create ResourceRequirments for the container.
func (k8s *k8s) containerResources() v1.ResourceRequirements {

	// requests
	resourceListRequests := v1.ResourceList{}
	if k8s.config.CPURequest > 0 {
		resourceListRequests[v1.ResourceCPU] = *resource.NewMilliQuantity(k8s.config.CPURequest, resource.DecimalSI)
	}
	if k8s.config.MemoryRequest > 0 {
		resourceListRequests[v1.ResourceMemory] = *resource.NewQuantity(k8s.config.MemoryRequest, resource.DecimalSI)
	}

	// limits
	resourceListLimits := v1.ResourceList{}
	if k8s.config.CPULimit > 0 {
		resourceListLimits[v1.ResourceCPU] = *resource.NewMilliQuantity(k8s.config.CPULimit, resource.DecimalSI)
	}
	if k8s.config.MemoryLimit > 0 {
		resourceListLimits[v1.ResourceMemory] = *resource.NewQuantity(k8s.config.MemoryLimit, resource.DecimalSI)
	}

	return v1.ResourceRequirements{
		Requests: resourceListRequests,
		Limits:   resourceListLimits,
	}
}

// Name returns user-friendly name of executor.
func (k8s *k8s) Name() string {
	return "Kubernetes Executor"
}

// generatePodName is generating pods based on KubernetesConfig struct.
// If KubernetesConfig has got non-empty PodName field then this field is
// used as a Pod name (even if selected Pod name is already in use - in
// this case pod spawning should fail).
// It returns string as a Pod name which should be used or error if cannot
// generate random suffix.
func (k8s *k8s) generatePodName() string {
	if k8s.config.PodName != "" {
		return k8s.config.PodName
	}
	return fmt.Sprintf("%s-%s", k8s.config.PodNamePrefix, uuid.New()[:8])
}

// newPod is a helper to build in-memory struture representing pod
// before sending it as request to API server. It can returns also
// error if cannot generate Pod name.
func (k8s *k8s) newPod(command string) (*v1.Pod, error) {

	resources := k8s.containerResources()
	podName := k8s.generatePodName()

	var zero int64
	return &v1.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: v1.ObjectMeta{
			Name:      podName,
			Namespace: k8s.config.Namespace,
			Labels:    map[string]string{"name": podName},
		},
		Spec: v1.PodSpec{
			NodeName:                      k8s.config.NodeName,
			DNSPolicy:                     "Default",
			RestartPolicy:                 "Never",
			HostNetwork:                   k8s.config.HostNetwork,
			TerminationGracePeriodSeconds: &zero,
			Containers: []v1.Container{
				{
					Name:            k8s.config.ContainerName,
					Image:           k8s.config.ContainerImage,
					Command:         []string{"sh", "-c", command},
					Resources:       resources,
					ImagePullPolicy: v1.PullIfNotPresent, // Default because swan image is not published yet.
					SecurityContext: &v1.SecurityContext{Privileged: &k8s.config.Privileged},
				},
			},
		},
	}, nil
}

// Execute creates a pod and runs the provided command in it. When the command completes, the pod
// is stopped i.e. the container is not restarted automatically.
func (k8s *k8s) Execute(command string) (TaskHandle, error) {
	podsAPI := k8s.clientset.Core().Pods(k8s.config.Namespace)
	command = k8s.config.Decorators.Decorate(command)

	// This is a workaround for kubernetes #31446
	// Make sure that at least one line of text is outputed from pod, to unblock .GetLogs() on apiserver call
	// with streamed response (when follow=true). Check kubernetes #31446 issue for more details.
	// https://github.com/kubernetes/kubernetes/pull/31446
	wrappedCommand := "echo;" + command

	// See http://kubernetes.io/docs/api-reference/v1/definitions/ for definition of the pod manifest.
	podManifest, err := k8s.newPod(wrappedCommand)
	if err != nil {
		log.Errorf("K8s executor: cannot create pod manifest")
		return nil, errors.Wrapf(err, "cannot create pod manifest")
	}
	log.Debugf("Starting '%s' pod=%s node=%s QoSclass=%s on kubernetes", command, podManifest.ObjectMeta.Name, podManifest.Spec.NodeName, k8sports.GetPodQOS(podManifest))

	pod, err := podsAPI.Create(podManifest)
	if err != nil {
		log.Errorf("K8s executor: cannot schedule pod %q with namespace %q", k8s.config.PodName, k8s.config.Namespace)
		return nil, errors.Wrapf(err, "cannot schedule pod %q with namespace %q",
			k8s.config.PodName, k8s.config.Namespace)
	}

	// Prepare local files
	outputDirectory, err := createOutputDirectory(command, "kubernetes")
	if err != nil {
		log.Errorf("Kubernetes Execute: cannot create output directory for command %q: %s", command, err.Error())
		return nil, err
	}

	stdoutFile, stderrFile, err := createExecutorOutputFiles(outputDirectory)
	if err != nil {
		removeDirectory(outputDirectory)
		log.Errorf("Kubernetes Execute: cannot create output files for command %q: %s", command, err.Error())
		return nil, err
	}
	stdoutFileName := stdoutFile.Name()
	stderrFileName := stderrFile.Name()
	stdoutFile.Close()
	stderrFile.Close()

	// After client-go supports GetPodQOS change to qos.GetPodQOS(pod)
	log.Debugf("K8s executor: pod %q QoS class %q", pod.Name, k8sports.GetPodQOS(pod))
	taskHandle := &k8sTaskHandle{
		podName:         pod.Name,
		command:         command,
		stdoutFilePath:  stdoutFileName,
		stderrFilePath:  stderrFileName,
		started:         make(chan struct{}),
		stopped:         make(chan struct{}),
		requestDelete:   make(chan struct{}, 1),
		exitCodeChannel: make(chan int, 1),
	}

	taskWatcher := &k8sWatcher{
		podsAPI:    podsAPI,
		pod:        pod,
		taskHandle: taskHandle,
		command:    wrappedCommand,

		stdoutFilePath: stdoutFileName,

		logsCopyFinished: make(chan struct{}, 1),

		started:         taskHandle.started,
		stopped:         taskHandle.stopped,
		requestDelete:   taskHandle.requestDelete,
		exitCodeChannel: taskHandle.exitCodeChannel,
	}

	err = taskWatcher.watch(k8s.config.LaunchTimeout)
	if err != nil {
		removeDirectory(outputDirectory)
		log.Errorf("K8s executor: cannot create task on pod %q", pod.Name)
		return nil, errors.Wrapf(err, "cannot create task on pod %q", pod.Name)
	}

	select {
	case <-taskWatcher.started:
		break
	case <-taskWatcher.stopped:
		break
	}

	// Best effort potential way to check if binary is started properly.
	taskHandle.Wait(100 * time.Millisecond)
	err = checkIfProcessFailedToExecute(command, k8s.Name(), taskHandle)
	if err != nil {
		return nil, err
	}
	return taskHandle, nil
}

// k8sTaskHandle implements the TaskHandle interface
type k8sTaskHandle struct {
	podsAPI       corev1.PodInterface
	stopped       chan struct{}
	started       chan struct{}
	requestDelete chan struct{}

	stdoutFilePath string
	stderrFilePath string // Kubernetes does not support separation of stderr & stdout, so this file will be empty

	// Use pointer with nil to indicate exitCode wasn't recevied.
	exitCode        *int
	exitCodeChannel chan int

	podName   string
	podHostIP string

	// Command requested by user. This is how this TaskHandle presents.
	command string
}

func (th *k8sTaskHandle) isTerminated() bool {
	select {
	case <-th.stopped:
		return true
	default:
		return false
	}
}

// Stop will delete the pod and block caller until done.
func (th *k8sTaskHandle) Stop() error {
	if th.isTerminated() {
		return nil
	}

	log.Debugf("K8s task handle: delete pod %q", th.podName)

	// Ask "watcher" to do deletion (watcher exists until pod was actually deleted!).
	// Ignore if request already done.
	select {
	case th.requestDelete <- struct{}{}:
	default:
	}

	log.Debugf("K8s task handle: waiting for pod %q to stop", th.podName)
	<-th.stopped
	log.Debugf("K8s task handle: pod %q stopped", th.podName)

	return nil
}

// Status returns the current task state in terms of RUNNING or TERMINATED.
func (th *k8sTaskHandle) Status() TaskState {
	if th.isTerminated() {
		return TERMINATED
	}

	return RUNNING
}

// ExitCode returns the exit code of the container running in the pod.
func (th *k8sTaskHandle) ExitCode() (int, error) {
	if !th.isTerminated() {
		return 0, errors.New("task is still running")
	}

	select {
	case exitCode, ok := <-th.exitCodeChannel:
		if ok {
			log.Debugf("K8s task handle: received exit code: %#v", exitCode)
			th.exitCode = &exitCode
		}
		log.Debug("K8s task handle: exitCode channel is already closed")
	default:
		log.Debug("K8s task handle: no exit code received (channel no closed yet)")
	}

	// Examine just or previously recevied exitCode.
	if th.exitCode == nil {
		log.Error("K8s task handle: exit code is unknown")
		return 0, errors.New("exit code unknown")
	}

	return *th.exitCode, nil
}

// Wait blocks until the pod terminates _or_ if timeout is provided, will exit ealier with
// false if the pod didn't terminate before the provided timeout.
func (th *k8sTaskHandle) Wait(timeout time.Duration) bool {
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

// EraseOutput deletes the directory where stdout file resides.
func (th *k8sTaskHandle) EraseOutput() error {
	outputDir := filepath.Dir(th.stdoutFilePath)
	return removeDirectory(outputDir)
}

func (th *k8sTaskHandle) Name() string {
	return fmt.Sprintf("Kubernetes pod named %q with command %q on %q", th.podName, th.command, th.Address())
}

// Address returns the host IP where the pod was scheduled.
func (th *k8sTaskHandle) Address() string {
	return th.podHostIP
}

// StdoutFile returns a file handle to the stdout file for the pod.
// NOTE: StdoutFile will block until stdout file is available.
func (th *k8sTaskHandle) StdoutFile() (*os.File, error) {
	return openFile(th.stdoutFilePath)
}

// StderrFile returns a file handle to the stderr file for the pod.
// NOTE: StderrFile will block until stderr file is available.
func (th *k8sTaskHandle) StderrFile() (*os.File, error) {
	return openFile(th.stderrFilePath)
}

type k8sWatcher struct {
	podsAPI corev1.PodInterface
	pod     *v1.Pod

	stdoutFilePath string

	started          chan struct{}
	stopped          chan struct{}
	logsCopyFinished chan struct{}
	requestDelete    chan struct{}
	exitCodeChannel  chan int

	taskHandle *k8sTaskHandle

	command string

	// one time events
	oncePodReady, oncePodFinished, oncePodDeleted sync.Once
	// one time actions
	onceDeletePod, onceSetupLogs sync.Once
	hasBeenRunning               bool
}

// watch creates instance of TaskHandle and is responsible for keeping it in-sync with k8s cluster
func (kw *k8sWatcher) watch(timeout time.Duration) error {
	selectorRaw := fmt.Sprintf("name=%s", kw.pod.Name)
	selector, err := labels.Parse(selectorRaw)
	if err != nil {
		return errors.Wrapf(err, "cannot create selector %q", selector)
	}

	// Prepare events watcher.
	watcher, err := kw.podsAPI.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		return errors.Wrapf(err, "cannot create watcher over selector %q", selector)
	}

	go func() {
		var timeoutChannel <-chan time.Time
		if timeout != 0 {
			timeoutChannel = time.After(timeout)
		}
		for {
			select {
			case event, ok := <-watcher.ResultChan():
				if !ok {
					// In some cases sender may close watcher channel. Usually it is safe to recreate it.
					log.Warnf("Pod %s: watcher event channel was unexpectly closed!", kw.pod.Name)
					var err error
					log.Debugf("Pod %s: recreating watcher stream", kw.pod.Name)
					watcher, err = kw.podsAPI.Watch(api.ListOptions{LabelSelector: selector})
					if err != nil {
						// We do not know what to do when error occurs.
						log.Panicf("Pod %s: cannot recreate watcher stream - %q", kw.pod.Name, err)
					}
					continue
				}

				// Don't care about anything else than pods.
				pod, ok := event.Object.(*v1.Pod)
				if !ok {
					continue
				}
				log.Debugf("K8s task watcher: event type=%v phase=%v", event.Type, pod.Status.Phase)

				switch event.Type {
				case watch.Added, watch.Modified:
					switch pod.Status.Phase {

					case v1.PodPending:
						continue

					case v1.PodRunning:
						// After client-go supports it change to
						// v1.IsPodReady(pod)
						if k8sports.IsPodReady(pod) {
							kw.whenPodReady()
						} else {
							log.Debug("K8s task watcher: Running but not ready")
						}

					case v1.PodFailed, v1.PodSucceeded:
						kw.whenPodFinished(pod)
						kw.deletePod()

					case v1.PodUnknown:
						log.Warnf("K8s task watcher: pod %q with command %q is in unknown phase. "+
							"Probably state of the pod could not be obtained, "+
							"typically due to an error in communicating with the host of the pod", pod.Name, kw.command)
					default:
						log.Warnf("K8s task watcher: unhandled pod phase event %q for pod %q", pod.Status.Phase, pod.Name)

					}
				case watch.Deleted:
					kw.whenPodDeleted()
					return // then only place we'are going to leave loop!

				case watch.Error:
					log.Errorf("K8s task watcher: kubernetes pod error event: %v", event.Object)

				default:
					log.Warnf("K8s task watcher: unhandled event type: %v", event.Type)

				}

			case <-kw.requestDelete:
				kw.deletePod()

			case <-timeoutChannel:
				// If task has been running then we need to ignore timeout
				if kw.hasBeenRunning {
					continue
				}
				log.Warnf("K8s task watcher: timeout occured afrer %f seconds. Pod %s has not been created.", timeout.Seconds(), kw.pod.Name)
				kw.deletePod()
			}
		}
	}()

	return nil
}

// ----------------------------- when pod ready
// closeStarted close started channel to indicate that pod has just started.
func (kw *k8sWatcher) whenPodReady() {
	kw.oncePodReady.Do(func() {
		log.Debug("K8s task watcher: Pod ready handler - mark Pod as running.")
		kw.hasBeenRunning = true
		kw.setupLogs()
		log.Debug("K8s task watcher: pod started [started]")
		close(kw.started)
	})
}

// whenPodFinished handler to acquire exit code from pod and pass it taskHandle.
// Additionally call whenPodReady handler to setupLogs and mark pod as running.
func (kw *k8sWatcher) whenPodFinished(pod *v1.Pod) {
	kw.oncePodFinished.Do(func() {
		kw.whenPodReady()
		kw.setExitCode(pod)
		log.Debug("K8s task watcher: pod finished")
	})
}

// setExitCode retrieves exit code from the cluster and sends it to task handle.
// Used only by whenPodFinished handler.
func (kw *k8sWatcher) setExitCode(pod *v1.Pod) {

	// Send exit code and close channel.
	sendExitCode := func(exitCode int) {
		kw.exitCodeChannel <- exitCode
		log.Debug("K8s task watcher: exit code sent [exitCodeChannel]")
		close(kw.exitCodeChannel)
	}

	// Look for an exit status from the container.
	// If more than one container is present, the last takes precedence.
	exitCode := -1
	for _, status := range pod.Status.ContainerStatuses {
		if status.State.Terminated == nil {
			continue
		}
		exitCode = int(status.State.Terminated.ExitCode)
	}
	if pod.Status.Phase == v1.PodFailed {
		log.Errorf("K8s task watcher: pod %q failed with exit code %d", pod.Name, exitCode)
	} else {
		log.Debugf("K8s task watcher: exit code retrieved: %d", exitCode)
	}
	sendExitCode(exitCode)
}

// whenPodDeleted handler on when watcher receives "deleted" event to signal taskHandle
// that pod is "stopped" and optionally setup the logs.
func (kw *k8sWatcher) whenPodDeleted() {
	kw.oncePodDeleted.Do(func() {
		kw.setupLogs()
		log.Debug("K8s task watcher: pod stopped [stopped]")
		// wait for logs to finish copying before announcing pod termination.
		<-kw.logsCopyFinished
		close(kw.stopped)

	})
}

// deletePod action that body is executed only once to ask kubernetes to deleted pod.
func (kw *k8sWatcher) deletePod() {
	kw.onceDeletePod.Do(func() {

		// Setting gracePeriodSeconds to zero will erase pod from API server and won't wait for it to exit.
		// Setting it to 1 second leaves responsibility of deleting the pod to kubelet.
		gracePeriodSeconds := int64(1)
		log.Debugf("deleting pod %q", kw.pod.Name)
		err := kw.podsAPI.Delete(kw.pod.Name, &api.DeleteOptions{
			GracePeriodSeconds: &gracePeriodSeconds,
		})
		if err != nil {
			log.Warnf("unsucessfull attemp to delete pod %q", kw.pod.Name)
		}
		log.Debugf("K8s task watcher: delete pod %q", kw.pod.Name)
	})
}

// setupLogs action creates log files and initializes goroutine that copies stream from kubernetes.
func (kw *k8sWatcher) setupLogs() {
	kw.onceSetupLogs.Do(func() {
		log.Debugf("K8s task watcher: setting up logs for pod %q", kw.pod.Name)

		// Wire up logs to task handle stdout.
		logStream, err := kw.podsAPI.GetLogs(kw.pod.Name, &v1.PodLogOptions{Follow: true}).Stream()
		if err != nil {
			log.Warnf("K8s task watcher: cannot create log stream: %s ", err.Error())
			close(kw.logsCopyFinished)
			return
		}

		// Start "copier" goroutine for copying logs api to local files.
		go func() {
			defer close(kw.logsCopyFinished)

			stdoutFile, err := os.OpenFile(kw.stdoutFilePath, os.O_WRONLY|os.O_SYNC, outputFilePrivileges)
			if err != nil {
				log.Errorf("K8s copier: cannot open file to copy logs: %s", err.Error())
				return
			}
			defer syncAndClose(stdoutFile)

			_, err = io.Copy(stdoutFile, logStream)
			if err != nil {
				log.Errorf("K8s copier: failed to copy container log stream to task output: %s", err.Error())
				return
			}

			log.Debugf("K8s copier: log copy and sync done for pod %q", kw.pod.Name)
		}()
	})
}
