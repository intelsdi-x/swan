package sessions

import (
	"fmt"
	"io/ioutil"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	"github.com/intelsdi-x/athena/pkg/snap"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/smartystreets/goconvey/convey"
)

// RunAndTestSnap starts snapd on random port returning clenaup function, plugin loader and string
// with snapd address
func RunAndTestSnap() (cleanup func(), loader *snap.PluginLoader, snapURL string) {
	snapd := testhelpers.NewSnapd()
	err := snapd.Start()
	convey.So(err, convey.ShouldBeNil)

	loaderConfig := snap.DefaultPluginLoaderConfig()
	snapURL = fmt.Sprintf("http://127.0.0.1:%d", snapd.Port())
	loaderConfig.SnapdAddress = snapURL

	loader, err = snap.NewPluginLoader(loaderConfig)

	convey.So(err, convey.ShouldBeNil)

	 cleanup = func() {
		err := snapd.CleanAndEraseOutput()
		convey.So(err, convey.ShouldBeNil)
		err = snapd.Stop()
		convey.So(err, convey.ShouldBeNil)}
	return
}

// PrepareAndTestPublisher creates session publisher and publisher output file, returns cleanup function,
// publisher and file name where publisher data will be stored
func PrepareAndTestPublisher(loader *snap.PluginLoader) ( cleanup func(), publisher *wmap.PublishWorkflowMapNode, publisherMetricsFile string) {

	tmpFile, err := ioutil.TempFile("", "session_test")
	convey.So(err, convey.ShouldBeNil)

	publisherMetricsFile = tmpFile.Name()
	loader.Load(snap.SessionPublisher)

	pluginName, _, err := snap.GetPluginNameAndType(snap.SessionPublisher)
	convey.So(err, convey.ShouldBeNil)

	publisher = wmap.NewPublishNode(pluginName, snap.PluginAnyVersion)
	convey.So(publisher, convey.ShouldNotBeNil)

	publisher.AddConfigItem("file", publisherMetricsFile)

	cleanup = func() {
		os.Remove(publisherMetricsFile)
	}
	return
}

// PrepareMockedTaskInfo based on provided path, creates mock task info that is used to
// configure collector
func PrepareMockedTaskInfo(outFilePath string) (cleanup func(), mockedTaskInfo *mocks.TaskInfo) {
	mockedTaskInfo = new(mocks.TaskInfo)
	file, err := os.Open(outFilePath)
	convey.So(err, convey.ShouldBeNil)
	mockedTaskInfo.On("StdoutFile").Return(file, nil)

	cleanup = func() {
		file.Close()
	}

	return
}

// ReadAndTestPublisherData reads publisher output, when data are read, function checks if we have all data,
// if we have all columns, if yes, then we compare read data against expectedMetrics.
// Function returns bool, which if true means that all data are valid, if false - data have not been read properly.
// If data do not match expected data, convey.So is "thrown"
func ReadAndTestPublisherData(dataFilePath string, expectedMetrics map[string]string) ( validData bool) {
	retries := 50
	validData = false
	expectedColumnsNum := 3
	for i := 0; i < retries; i++ {
		time.Sleep(100 * time.Millisecond)

		data, err := ioutil.ReadFile(dataFilePath)
		if err != nil {
			continue
		}

		if len(data) > 0 {
			// Check if we have all published data
			lines := strings.Split(string(data), "\n")
			if len(lines) < len(expectedMetrics) {
				convey.Printf("There should be at least ",
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
				convey.Printf("There should be at least 3 columns for all lines. ",
					"Checking again.")
				continue
			}

			// Now we are sure that we have all data so we begin validation
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
	return
}


// soMetricRowIsValid function takes 3 strings:
// namespace, tags, value - and checks if provided value is almost equal
// corresponding one from expectedMetrics
func soMetricRowIsValid(
	expectedMetrics map[string]string,
	namespace, tags, value string) {

	// Check tags.
	tagsSplitted := strings.Split(tags, ",")
	convey.So(len(tagsSplitted), convey.ShouldBeGreaterThanOrEqualTo, 1)
	convey.So("foo=bar", convey.ShouldBeIn, tagsSplitted)

	// Split namespace string
	namespaceSplitted := strings.Split(namespace, "/")

	// Read metric name (from namespace) is "key" used to search corresponding
	// expectedMetric "value"
	expectedValue, ok := expectedMetrics[namespaceSplitted[len(namespaceSplitted)-1]]
	convey.So(ok, convey.ShouldBeTrue)

	// Reduce string-encoded-float to common precision for comparison.
	expectedValueFloat, err := strconv.ParseFloat(expectedValue, 64)
	convey.So(err, convey.ShouldBeNil)
	valueFloat, err := strconv.ParseFloat(value, 64)
	convey.So(err, convey.ShouldBeNil)

	epsilon := 0.00001
	convey.So(expectedValueFloat, convey.ShouldAlmostEqual, valueFloat, epsilon)
}
