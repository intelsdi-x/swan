package cassandra_publisher

import (
	"testing"
	"github.com/intelsdi-x/swan/pkg/executor"
	"io/ioutil"
	"time"
	"strings"
	"regexp"
	"fmt"
	. "github.com/smartystreets/goconvey/convey"
	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"k8s.io/kubernetes/third_party/golang/go/doc/testdata"
)


func TestCassandraPublisher(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	err := RunCassandraPublisherWorkflow()
	Convey("When running Cassadndra workflow.", t, func() {
		So(err, ShouldBeNil)
	})

	value, tags, err := GetValueAndTagsFromCassandra()
	Convey("When getting values from Cassadndra", t, func() {
		So(err, ShouldBeNil)
		Convey("Value should be different than zero", func() {
			So(value, ShouldNotEqual, 0)
		})
		Convey("Tags should be approprierate", func() {
			So(tags["swan_experiment"], ShouldEqual, "example-experiment")
			So(tags["swan_phase"], ShouldEqual, "example-phase")
		})
	})
}

func GetValueAndTagsFromCassandra() (value float64, tags map[string]string, err error){
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.ProtoVersion = 4
	cluster.Keyspace = "snap"
	cluster.Consistency = gocql.All
	session, _ := cluster.CreateSession()
	defer session.Close()

	// cqlsh> select doubleval, tags from snap.metrics where ns='/intel/mock/host6/baz' AND ver=0 AND host='' ORDER BY time ASC limit 1;
	//ns                    | ver | host | time                            | boolval | doubleval | strval | tags                                                                     | valtype
	//-----------------------+-----+------+---------------------------------+---------+-----------+--------+--------------------------------------------------------------------------+-----------
	///intel/mock/host6/baz |   0 |      | 2016-05-13 14:05:31.194000+0000 |    null |        89 |   null | {'swan_experiment': 'example-experiment', 'swan_phase': 'example-phase'} | doubleval

	if err := session.Query(`SELECT doubleval, tags FROM snap.metrics WHERE ns = ? AND ver = ? AND host = ? LIMIT 1`,
		"/intel/mock/host6/baz", 0, "").Consistency(gocql.One).Scan(&value, &tags); err != nil {
		return value, tags, err
	}

	return value, tags, err
}

func createSnapTaskAndGetTaskId() (taskId string, err error) {
	e := executor.NewLocal()
	loadTask, err := e.Execute("snapctl task create -t task.yml")
	if err != nil {
		log.Errorf("Load task execution failed: %s\n", err.Error())
		return taskId, err
	}
	defer loadTask.Clean()
	defer loadTask.EraseOutput()
	loadTask.Wait(0)

	exitCode, err := loadTask.ExitCode()
	if err != nil || exitCode != 0 {
		log.Errorf("Load task execution failed: %s\n", err.Error())
		return taskId, err
	}

	taskId, err = getTaskIdFromTaskHandle(loadTask)
	if err != nil {
		log.Errorf("Load task execution failed: %s\n", err.Error())
		return taskId, err
	}

	return taskId, err
}

func startSnapTask(taskId string) (err error) {

}


func RunCassandraPublisherWorkflow() error {
	taskId, err := createSnapTaskAndGetTaskId()

	runTask, err := e.Execute("snapctl task start " + taskId)
	if err != nil {
		log.Errorf("Start task execution failed: %s\n", err.Error())
		return err
	}
	defer runTask.Clean()
	defer runTask.EraseOutput()

	runTask.Wait(0)

	exitCode, err = runTask.ExitCode()
	if err != nil || exitCode != 0 {
		log.Errorf("Start task failed, err: %s; exit_code: %d\n", err.Error(), exitCode)
		return err
	}

	time.Sleep(2 * time.Second)

	stopTask, err := e.Execute("snapctl task stop " + taskId)
	if err != nil {
		log.Errorf("Stop task execution failed: %s\n", err.Error())
		return err
	}
	defer stopTask.Clean()
	defer stopTask.EraseOutput()

	stopTask.Wait(0)

	exitCode, err = stopTask.ExitCode()
	if err != nil || exitCode != 0 {
		log.Errorf("Start task failed, err: %s; exit_code: %d\n", err.Error(), exitCode)
		return err
	}

	cleanTask, err := e.Execute("snapctl task remove " + taskId)
	if err != nil {
		log.Errorf("Remove task execution failed: %s\n", err.Error())
		return err
	}
	defer cleanTask.Clean()
	defer cleanTask.EraseOutput()

	cleanTask.Wait(0)

	exitCode, err = cleanTask.ExitCode()
	if err != nil || exitCode != 0 {
		log.Errorf("Remove task execution failed: %s\n", err.Error())
		return err
	}

	return err
}

func getTaskIdFromTaskHandle(loadTaskHandle executor.TaskHandle) (string, error) {
	outputFile, err := loadTaskHandle.StdoutFile()
	if err != nil {
		return "", err
	}

	output, err := ioutil.ReadAll(outputFile)
	if err != nil {
		return "", err
	}

	return getTaskIdFromSnapOutput(string(output))
}

func matchNotFound(match []string) bool {
	return match == nil || len(match) < 2 || len(match[1]) == 0
}

/**
$ snapctl task create -t task.yml
Using task manifest to create task
Task created
ID: cbf7d3f5-c4b5-4deb-a720-6212c83d3db5
Name: Task-cbf7d3f5-c4b5-4deb-a720-6212c83d3db5
State: Running
 */
func getTaskIdFromSnapOutput(snapctlOutput string) (string, error) {
	lines := strings.Split(snapctlOutput, "\n")
	getTaskIdRegex := regexp.MustCompile(`ID:\s(.*)$`)

	for _, line := range lines {
		match := getTaskIdRegex.FindStringSubmatch(line)
		if matchNotFound(match) {
			continue
		} else {
			return match[1], nil
		}
	}

	return "", fmt.Errorf("Could not find task id in string: %s\n", snapctlOutput)
}
