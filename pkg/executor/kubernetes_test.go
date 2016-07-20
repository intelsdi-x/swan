/* manuall run

cd kubernetes
# https://github.com/kubernetes/kubernetes/blob/master/docs/devel/running-locally.md

j kubernetes
sudo -s GOPATH=$GOPATH PATH=$PATH
export PATH=${PATH}:/home/ppalucki/work/gopath/src/k8s.io/kubernetes/third_party/etcd
hack/local-up-cluster.sh -o _output/bin

go test -run K8s -v github.com/intelsdi-x/swan/pkg/executor

cluster/kubectl.sh get pod stress1
cluster/kubectl.sh delete pod stress1
cluster/kubectl.sh delete pods --all
cluster/kubectl.sh get pod stress1
*/
package executor_test

import (
	"fmt"
	"io"
	"io/ioutil"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func TestK8s(t *testing.T) {
	log.SetFormatter(&log.TextFormatter{})
	log.SetLevel(log.DebugLevel)

	fmt.Println("new executor")
	k8s, err := executor.NewKubernetesExectuor(
		executor.KubernetesConfig{
			Address: "http://127.0.0.1:8080",
			PodName: "stress1",
		},
	)
	check(err)
	log.Printf("k8s = %+v\n", k8s)

	log.Println("exectute...")
	th, err := k8s.Execute("stress -c 1 -t 15")
	check(err)
	log.Printf("th = %+v\n", th)

	log.Println("wait...")

	log.Println("wait result:", th.Wait(5*time.Second))

	err = th.Stop() // delete pod
	check(err)

	// just helper function
	pr := func(r io.Reader, err error) string {
		check(err)
		data, err := ioutil.ReadAll(r)
		check(err)
		return string(data)
	}

	log.Println("stderr: ", pr(th.StderrFile()))
	log.Println("stdout: ", pr(th.StdoutFile()))

	// missing
	th.ExitCode()
}
