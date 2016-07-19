/* manuall run

cd kubernetes
# https://github.com/kubernetes/kubernetes/blob/master/docs/devel/running-locally.md

j kubernetes
sudo -s
export GOPATH="/home/ppalucki/work/gopath"
export PATH=${PATH}:/home/ppalucki/work/gopath/src/k8s.io/kubernetes/third_party/etcd
hack/local-up-cluster.sh
hack/local-up-cluster.sh -o _output/bin

go test -run K8s -v github.com/intelsdi-x/swan/pkg/executor

cluster/kubectl.sh get pod stress1
cluster/kubectl.sh delete pod stress1
*/
package executor_test

import (
	"fmt"
	"testing"

	"github.com/intelsdi-x/swan/pkg/executor"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestK8s(t *testing.T) {

	println("BLE !!!")
	fmt.Println("client....")

	k8s, err := executor.NewKubernetesExectuor("http://127.0.0.1:8080", "stress1")
	check(err)
	println("BLE !!!")

	fmt.Printf("k8s = %+v\n", k8s)

	fmt.Println("schedule a pod....")
	th := k8s.Execute("stress -c 1")
	println("BLE !!!")

	fmt.Println("scheduled")
	fmt.Printf("th = %+v\n", th)

	th.Wait(0)
	// play with
	/*
		./cluster/kubectl.sh delete pods --all
		./cluster/kubectl.sh get pod stress1
	*/
	th.Stop()

}
