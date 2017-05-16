// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package testhelpers

import (
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strconv"
	"strings"
	"time"

	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/pkg/errors"
	"github.com/smartystreets/goconvey/convey"
)

// SwanPath is an absolute path of project.
var SwanPath = path.Join(os.Getenv("GOPATH"), "src", "github.com/intelsdi-x/swan")

// AssertFileExists makes sure that executable is available in $PATH or panics returning location of executable.
func AssertFileExists(executable string) string {
	path, err := exec.LookPath(executable)
	if err != nil {
		panic(errors.Wrapf(err, "cannot find required binary %q in $PATH", executable))
	}
	return path
}

// RunAndTestSnaptel checks snapteld on returning cleanup function, plugin loader and string
// with snapteld address
// Note: It is facade function that assumes snapteld is running all the time.
// But can be easily replaced with self-provided snapteld.
func RunAndTestSnaptel() (cleanup func(), loader *snap.PluginLoader, snaptelURL string) {

	loaderConfig := snap.DefaultPluginLoaderConfig()
	loader, err := snap.NewPluginLoader(loaderConfig)
	convey.So(err, convey.ShouldBeNil)
	snaptelURL = loaderConfig.SnapteldAddress
	cleanup = func() {}
	return
}

// PreparePublisher creates session publisher and publisher output file, returns cleanup function,
// publisher and file name where publisher data will be stored
func PreparePublisher(loader *snap.PluginLoader) (cleanup func(), publisher *wmap.PublishWorkflowMapNode, publisherMetricsFile string) {

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
		os.RemoveAll(publisherMetricsFile)
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
func ReadAndTestPublisherData(dataFilePath string, expectedMetrics map[string]string, session *snap.Session) (validData bool) {
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
				convey.Printf("There should be at least %d lines. Checking again.", len(expectedMetrics))
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
				convey.Print("There should be at least 3 columns for all lines. ",
					"Checking again.")
				continue
			}

			// Now we are sure that we have all data so we begin validation
			for i := 0; i < len(expectedMetrics); i++ {
				columns := strings.Split(lines[i], "\t")
				soMetricRowIsValid(expectedMetrics, columns[0], columns[1], columns[2], session)
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
func soMetricRowIsValid(expectedMetrics map[string]string, namespace, tags, value string, session *snap.Session) {

	// Check tags.
	tagsSplitted := strings.Split(tags, ",")
	convey.So(len(tagsSplitted), convey.ShouldBeGreaterThanOrEqualTo, 1)

	convey.So(session.GetTags(), convey.ShouldContainKey, "foo")
	convey.So(session.GetTags()["foo"], convey.ShouldEqual, "bar")

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
