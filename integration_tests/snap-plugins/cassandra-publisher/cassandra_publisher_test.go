package cassandra

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
	"github.com/vektra/errors"
	"os"
	"path"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
)


func TestCassandraPublisher(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	err := loadSnapPlugins()
	Convey("When loading snap plugins.", t, func() {
		So(err, ShouldBeNil)
	})

	err = runCassandraPublisherWorkflow()
	Convey("When running Cassadndra workflow.", t, func() {
		So(err, ShouldBeNil)
	})

	value, tags, err := getValueAndTagsFromCassandra()
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

func loadSnapPlugins() (err error) {
	defer func() {
		if r := recover(); r != nil {
			log.Errorf("Recovered from panic: %v\n", r)
			err = r.(error)
		}
	}()

	snapClient, err := client.New("http://127.0.0.1:8181", "v1", true)
	if err != nil {
		return fmt.Errorf("loadSnapPlugins: error connecting to snap: %s\n",
				   err.Error())
	}
	plugins := snap.NewPlugins(snapClient)

	if isNotCassandraPluginLoaded(plugins) {
		loadCassandraPublisherPlugin(plugins)
	}
	if isNotMockCollectorPluginLoaded(plugins) {
		loadMockCollectorPlugin(plugins)
	}
	if isNotSessionProcessorPluginLoaded(plugins) {
		loadSessionProcessorPlugin(plugins)
	}

	return err
}

func isNotCassandraPluginLoaded(pluginClient *snap.Plugins) (isLoaded bool) {
	isLoaded, err := pluginClient.IsLoaded("publisher", "cassandra")
	if err != nil {
		panic("isNotCassandraPluginLoaded: " + err.Error())
	}
	return !isLoaded
}

func isNotMockCollectorPluginLoaded(pluginClient *snap.Plugins) (isLoaded bool) {
	isLoaded, err := pluginClient.IsLoaded("collector", "mock")
	if err != nil {
		panic("isNotMockCollectorPluginLoaded: " + err.Error())
	}
	return !isLoaded
}

func isNotSessionProcessorPluginLoaded(pluginClient *snap.Plugins) (isLoaded bool) {
	isLoaded, err := pluginClient.IsLoaded("processor", "session-processor")
	if err != nil {
		panic("isNotSessionProcessorPluginLoaded: " + err.Error())
	}
	return !isLoaded
}

func loadCassandraPublisherPlugin(*snap.Plugins) (err error) {
	goPath := os.Getenv("GOPATH")
	pluginPath :=
	return nil
}

func loadMockCollectorPlugin(*snap.Plugins) (err error) {
	goPath := os.Getenv("GOPATH")
	return nil
}


func loadSessionProcessorPlugin(*snap.Plugins) (err error) {
	goPath := os.Getenv("GOPATH")
	return nil
}


func getValueAndTagsFromCassandra() (value float64, tags map[string]string, err error){
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

	session.Query(`SELECT doubleval, tags FROM snap.metrics
			WHERE ns = ? AND ver = ? AND host = ? LIMIT 1`,
				"/intel/mock/host6/baz", 0, "").
			Consistency(gocql.One).Scan(&value, &tags)
	return value, tags, err
}

func runCassandraPublisherWorkflow() (err error) {
	var errorStrings []string
	taskID, err := createSnapTaskAndGetTaskID()
	if err != nil {
		return err
	}

	runCommand("snapctl task start " + taskID)
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	} else {
		time.Sleep(2 * time.Second)
	}

	runCommand("snapctl task stop " + taskID)
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	}
	runCommand("snapctl task remove " + taskID)
	if err != nil {
		errorStrings = append(errorStrings, err.Error())
	}

	if len(errorStrings) > 0 {
		return errors.New(strings.Join(errorStrings, "\n"))
	}
	return err
}

func createSnapTaskAndGetTaskID() (taskID string, err error) {
	goPath := os.Getenv("GOPATH")
	taskPath := path.Join(goPath, "src", "github.com", "intelsdi-x", "swan",
		"integration_tests", "snap-plugins", "cassandra-publisher", "task.yml")

	_, err = os.Stat(taskPath)
	if err != nil {
		return taskID, fmt.Errorf("Stat on task file in location %s returned error %s",
			taskPath, err)
	}

	e := executor.NewLocal()
	createTask, err := e.Execute("snapctl task create -t " + taskPath)
	if err != nil {
		log.Errorf("Load task execution failed: %s\n", err.Error())
		return taskID, err
	}
	defer createTask.Clean()
	defer createTask.EraseOutput()
	createTask.Wait(0)

	exitCode, err := createTask.ExitCode()
	if exitCode != 0 {
	 	return taskID, fmt.Errorf("Create task execution returned code %d\n", exitCode)
	}

	taskID, err = getTaskIDFromTaskHandle(createTask)
	if err != nil {
		log.Errorf("Load task execution failed: %s\n", err.Error())
		return taskID, err
	}

	return taskID, err
}

func getTaskIDFromTaskHandle(loadTaskHandle executor.TaskHandle) (string, error) {
	outputFile, err := loadTaskHandle.StdoutFile()
	if err != nil {
		return "", err
	}

	output, err := ioutil.ReadAll(outputFile)
	if err != nil {
		return "", err
	}

	return getTaskIDFromSnapOutput(string(output))
}


/**
Output example:
$ snapctl task create -t task.yml
Using task manifest to create task
Task created
ID: cbf7d3f5-c4b5-4deb-a720-6212c83d3db5
Name: Task-cbf7d3f5-c4b5-4deb-a720-6212c83d3db5
State: Running
 */
func getTaskIDFromSnapOutput(snapctlOutput string) (string, error) {
	lines := strings.Split(snapctlOutput, "\n")
	getTaskIDRegex := regexp.MustCompile(`ID:\s(.*)$`)

	for _, line := range lines {
		match := getTaskIDRegex.FindStringSubmatch(line)
		if matchNotFound(match) {
			continue
		} else {
			return match[1], nil
		}
	}

	return "", fmt.Errorf("Could not find task id in string: %s\n", snapctlOutput)
}

func matchNotFound(match []string) bool {
	return match == nil || len(match) < 2 || len(match[1]) == 0
}


func runCommand(command string) (err error) {
	e := executor.NewLocal()
	commandTask, err := e.Execute(command)
	if err != nil {
		log.Errorf("Command \"%s\" invocation failed: %s\n", command, err.Error())
		return err
	}
	defer commandTask.Clean()
	defer commandTask.EraseOutput()

	commandTask.Wait(0)

	exitCode, err := commandTask.ExitCode()
	if err != nil || exitCode != 0 {
		log.Errorf("Command \"%s\" failed, err: %s; exit_code: %d\n",
			command, err.Error(), exitCode)
		return err
	}

	return err
}
