package executor

import (
	"fmt"
	"os"
	"time"

	"github.com/intelsdi-x/swan/pkg/isolation"
	"github.com/pkg/errors"

	"k8s.io/kubernetes/pkg/api"
	"k8s.io/kubernetes/pkg/api/unversioned"
	"k8s.io/kubernetes/pkg/client/restclient"
	client "k8s.io/kubernetes/pkg/client/unversioned"
	"k8s.io/kubernetes/pkg/labels"
	"k8s.io/kubernetes/pkg/watch"
)

// -------------------- options -------------------------
// http://dave.cheney.net/2014/10/17/functional-options-for-friendly-apis
type kubernetesExecutorOption func(_ *kubernetesExecutor)

func CPURequest(request int) kubernetesExecutorOption {
	return func(k *kubernetesExecutor) {
		k.cpuRequest = request
	}
}

func CPULimit(limit int) kubernetesExecutorOption {
	return func(k *kubernetesExecutor) {
		k.cpuLimit = limit
	}
}

func Decorator(decorator isolation.Decorator) kubernetesExecutorOption {
	return func(k *kubernetesExecutor) {
		k.decorators = append(k.decorators, decorator)
	}
}

// ----------------------- executor ------------------------

type kubernetesExecutor struct {
	podName    string // unique pod identifier.
	address    string
	cpuRequest int
	cpuLimit   int
	decorators isolation.Decorators

	client *client.Client
}

func NewKubernetesExectuor(address, podName string, options ...kubernetesExecutorOption) (*kubernetesExecutor, error) {

	// new
	k8s := &kubernetesExecutor{address: address, podName: podName}

	// apply options
	for _, option := range options {
		option(k8s)
	}

	config := &restclient.Config{
		Host: address,
		// Username: "test", // TODO authorization
		// Password: "password",
	}
	var err error
	k8s.client, err = client.New(config)
	if err != nil {
		return nil, errors.Wrapf(err, "can't initilize kubernetes Client for host = %q", address)
	}

	return k8s, nil
}

func (k8s *kubernetesExecutor) Execute(command string) *kubernetesTaskHandle {

	// TODO: do we want own namespace ?
	ns := api.NamespaceDefault

	podsapi := k8s.client.Pods(ns)

	pod, err := podsapi.Create(&api.Pod{
		TypeMeta: unversioned.TypeMeta{},
		ObjectMeta: api.ObjectMeta{
			Name:      k8s.podName,
			Namespace: ns,
			Labels:    map[string]string{"name": k8s.podName},
		},

		Spec: api.PodSpec{
			Containers: []api.Container{
				api.Container{
					Name:    k8s.podName,
					Image:   "jess/stress", // replace with image of swan
					Command: []string{"sh", "-c", command},
				},
			},
		},
	})

	if err != nil {
		panic(fmt.Sprintf("can't schedule pod %q ns=%q: %s", k8s.podName, ns, err))
	}

	// Wait.
	// req, err := labels.NewRequirement("name", labels.EqualsOperator, sets.NewString("jebac"))
	// selector := labels.NewSelector()
	// selector.Add(*req)
	//// for some reasons it doesn't work

	selector, err := labels.Parse(fmt.Sprintf("name=%s", k8s.podName))
	if err != nil {
		panic(err)
	}
	watch, err := podsapi.Watch(api.ListOptions{LabelSelector: selector})
	if err != nil {
		panic(err)
	}

	for event := range watch.ResultChan() {
		// fmt.Printf("event = %+v\n", event)
		// log.Println(event.Object)

		pod, ok := event.Object.(*api.Pod)
		if ok {
			if pod.Status.Phase == api.PodRunning {
				break
			}
		}
	}

	return &kubernetesTaskHandle{pod, watch}
}

// --------------------  task handle -------------------------
type kubernetesTaskHandle struct {
	pod   *api.Pod
	watch watch.Interface // TODO: use that to spot our PodTerminated (in background gorountien)
}

func (k8s *kubernetesTaskHandle) Stop() error {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) Status() TaskState {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) ExitCode() (int, error) {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) Wait(timeout time.Duration) bool {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) Clean() error {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) EraseOutput() error {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) Address() string {
	panic("not implemented")
}

// --------------------- task info ----------------------------------
func (k8s *kubernetesTaskHandle) StdoutFile() (*os.File, error) {
	panic("not implemented")
}

func (k8s *kubernetesTaskHandle) StderrFile() (*os.File, error) {
	panic("not implemented")
}
