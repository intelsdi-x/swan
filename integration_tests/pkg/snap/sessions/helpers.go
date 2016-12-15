package sessions

import (
	"fmt"
	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	So "github.com/smartystreets/goconvey/convey"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

// RunAndTestSnap starts snapd on random port returning clenaup function, plugin loader and string
// with snapd address
func RunAndTestSnap() (func(), *snap.PluginLoader, string) {
	snapd := testhelpers.NewSnapd()
	err := snapd.Start()
	So.So(err, So.ShouldBeNil)

	loaderConfig := snap.DefaultPluginLoaderConfig()
	snapdAddress := fmt.Sprintf("http://127.0.0.1:%d", snapd.Port())
	loaderConfig.SnapdAddress = snapdAddress

	loader, err := snap.NewPluginLoader(loaderConfig)

	So.So(err, So.ShouldBeNil)

	return func() {
		err := snapd.Stop()
		err2 := snapd.CleanAndEraseOutput()
		So.So(err, So.ShouldBeNil)
		So.So(err2, So.ShouldBeNil)
	}, loader, snapdAddress
}

// PrepareAndTestPublisher creates session publisher and publisher output file, returns clenaup function
// publisher and file name where publisher data will be stored
func PrepareAndTestPublisher(loader *snap.PluginLoader) (func(), *wmap.PublishWorkflowMapNode, string) {

	tmpFile, err := ioutil.TempFile("", "session_test")
	So.So(err, So.ShouldBeNil)

	publisherMetricsFile := tmpFile.Name()
	loader.Load(snap.SessionPublisher)

	pluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)
	So.So(err, So.ShouldBeNil)

	publisher := wmap.NewPublishNode(pluginName, snap.PluginAnyVersion)
	So.So(publisher, So.ShouldNotBeNil)

	publisher.AddConfigItem("file", publisherMetricsFile)

	return func() {
		os.Remove(publisherMetricsFile)
	}, publisher, publisherMetricsFile
}

// PrepareMockedTask based on provided path, creates mock task that is used to
// configure collector
func PrepareMockedTask(outFilePath string) (func(), *mocks.TaskInfo) {
	mockedTaskInfo := new(mocks.TaskInfo)
	file, err := os.Open(outFilePath)
	So.So(err, So.ShouldBeNil)
	mockedTaskInfo.On("StdoutFile").Return(file, nil)

	return func() {
		file.Close()
	}, mockedTaskInfo
}

// ReadAndTestPublisherData reads publisher output, when data are read, function checks if we have all data,
// if we have all columns, if yes, then we compare read data against expectedMetrics.
// Function returns bool indicating if read data are the same as expected data
func ReadAndTestPublisherData(dataFilePath string, expectedMetrics map[string]string, t *testing.T) bool {
	retries := 50
	validData := false
	expectedColumnsNum := 3
	for i := 0; i < retries; i++ {
		time.Sleep(100 * time.Millisecond)

		dat, err := ioutil.ReadFile(dataFilePath)
		if err != nil {
			continue
		}

		if len(dat) > 0 {
			// Check if we have all published data
			lines := strings.Split(string(dat), "\n")
			if len(lines) < len(expectedMetrics) {
				t.Log("There should be at least ",
					len(expectedMetrics),
					" lines. Checking again.")
				continue
			}

			allLinesHaveAllColumns := true
			// All lines should have expectedColumnsNum (3) columns.
			for i := 0; i < len(expectedMetrics); i++ {
				columns := strings.Split(lines[i], "\t")
				if len(columns) < expectedColumnsNum {
					allLinesHaveAllColumns = false
					break
				}
			}

			if !allLinesHaveAllColumns {
				t.Log("There should be at least 3 columns for all lines. ",
					"Checking again.")
				continue
			}

			for i := 0; i < len(expectedMetrics); i++ {
				columns := strings.Split(lines[i], "\t")
				soMetricRowIsValid(
					expectedMetrics,
					columns[0], columns[1], columns[2])
			}
			validData = true
			break
		}
	}
	return validData
}

func soMetricRowIsValid(
	expectedMetrics map[string]string,
	namespace, tags, value string) {

	// Check tags.
	tagsSplitted := strings.Split(tags, ",")
	So.So(len(tagsSplitted), So.ShouldBeGreaterThanOrEqualTo, 1)
	So.So("foo=bar", So.ShouldBeIn, tagsSplitted)

	// Check namespace & values.
	namespaceSplitted := strings.Split(namespace, "/")
	expectedValue, ok := expectedMetrics[namespaceSplitted[len(namespaceSplitted)-1]]
	So.So(ok, So.ShouldBeTrue)

	// Reduce string-encoded-float to common precision for comparison.
	expectedValueFloat, err := strconv.ParseFloat(expectedValue, 64)
	So.So(err, So.ShouldBeNil)
	valueFloat, err := strconv.ParseFloat(value, 64)
	So.So(err, So.ShouldBeNil)

	epsilon := 0.00001
	So.So(expectedValueFloat, So.ShouldAlmostEqual, valueFloat, epsilon)
}
