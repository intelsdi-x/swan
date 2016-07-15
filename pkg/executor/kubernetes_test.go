package executor

import "testing"

func TestK8s(t *testing.T) {

	/* setup
	https://github.com/kubernetes/kubernetes/blob/master/docs/devel/running-locally.md
	./hack/local-up-cluster.sh
	*/

	k8s := NewKubernetesExectuor("http://127.0.0.1:8080", "stress1")
	th := k8s.Execute("stress -c 1")
	th.Wait(0)
	// play with
	/*
		./cluster/kubectl.sh delete pods --all
		./cluster/kubectl.sh get pod stress1
	*/
	th.Stop()

}
