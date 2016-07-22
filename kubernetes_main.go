// THIS FILE WILL BE DELETED.

package main

import (
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	log.SetFormatter(&log.TextFormatter{FullTimestamp: true})
	log.SetLevel(log.DebugLevel)

	k8s, err := executor.NewKubernetesExecutor(
		executor.KubernetesConfig{
			Address: "http://127.0.0.1:8080",
			PodName: "stress1",
		},
	)
	check(err)

	log.Println("execute...")
	th, err := k8s.Execute("sleep 2 && exit 1")
	check(err)

	log.Printf("th.Status() = %v\n", th.Status())

	log.Println("wait for 10 seconds result:", th.Wait(10*time.Second))

	log.Println("stop ...")
	// wait()
	err = th.Stop() // delete pod
	check(err)
	log.Println("stopped with status:", th.Status())

	log.Println("hostip:", th.Address())


	// missing
	code, _ := th.ExitCode()
	log.Printf("exit code: %d", code)
}
