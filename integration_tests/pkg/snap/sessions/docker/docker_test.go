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

package sessions

import (
	"fmt"
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/kubernetes"
	"github.com/intelsdi-x/swan/pkg/snap"
	"github.com/intelsdi-x/swan/pkg/snap/sessions/docker"
	. "github.com/smartystreets/goconvey/convey"
)

func TestSnapDockerSession(t *testing.T) {
	Convey("Preparing Snap and Kubernetes enviroment", t, func() {
		cleanup, loader, snapteldAddr := testhelpers.RunAndTestSnaptel()
		defer cleanup()

		err := loader.Load(snap.DockerCollector, snap.FilePublisher)
		So(err, ShouldBeNil)
		publisherPluginName, _, err := snap.GetPluginNameAndType(snap.FilePublisher)
		So(err, ShouldBeNil)

		resultsFile, err := ioutil.TempFile("", "session_test")
		So(err, ShouldBeNil)
		resultsFileName := resultsFile.Name()
		defer os.Remove(resultsFileName)
		resultsFile.Close()

		publisher := wmap.NewPublishNode(publisherPluginName, snap.PluginAnyVersion)
		publisher.AddConfigItem("file", resultsFileName)

		// Run Kubernetes
		exec := executor.NewLocal()
		config := kubernetes.UniqueConfig()
		config.RetryCount = 10
		kubernetesLauncher := kubernetes.New(exec, exec, config)
		kubernetesHandle, err := kubernetesLauncher.Launch()
		So(err, ShouldBeNil)
		So(kubernetesHandle, ShouldNotBeNil)
		defer kubernetesHandle.EraseOutput()
		defer kubernetesHandle.Stop()

		// Waiting for Kubernetes Executor.
		kubernetesConfig := executor.DefaultKubernetesConfig()
		kubernetesConfig.Address = fmt.Sprintf("127.0.0.1:%d", config.KubeAPIPort)
		kubeExecutor, err := executor.NewKubernetes(kubernetesConfig)
		So(err, ShouldBeNil)

		podHandle, err := kubeExecutor.Execute("stress-ng -c 1")
		So(err, ShouldBeNil)
		defer podHandle.EraseOutput()
		defer podHandle.Stop()

		Convey("Launching Docker Session", func() {
			dockerConfig := docker.DefaultConfig()
			dockerConfig.SnapteldAddress = snapteldAddr
			dockerConfig.Publisher = publisher
			dockerLauncher, err := docker.NewSessionLauncher(dockerConfig)
			So(err, ShouldBeNil)

			tags := make(map[string]interface{})
			tags["foo"] = "bar"
			dockerHandle, err := dockerLauncher.LaunchSession(
				nil,
				tags,
			)
			So(err, ShouldBeNil)
			defer dockerHandle.Stop()

			So(dockerHandle.Status(), ShouldEqual, executor.RUNNING)
			time.Sleep(10 * time.Second)

			err = dockerHandle.Stop()
			So(err, ShouldBeNil)

			// one measurement should contains more then one metric.
			oneMeasurement, err := testhelpers.GetOneMeasurementFromFile(resultsFileName)
			So(err, ShouldBeNil)
			So(len(oneMeasurement), ShouldBeGreaterThan, 0)
			So(oneMeasurement[0].Tags["foo"], ShouldEqual, "bar")
		})
	})
}
