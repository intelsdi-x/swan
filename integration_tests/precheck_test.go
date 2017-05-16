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

package precheck

import (
	"testing"

	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/snap"
	. "github.com/smartystreets/goconvey/convey"
)

const ()

func TestExecutables(t *testing.T) {

	requiredExecutables := []string{

		// aggressors
		"caffe.sh",
		"stress-ng",

		// experiments
		"memcached-sensitivity-profile",
		"specjbb-sensitivity-profile",

		// snap
		"snaptel",
		"snapteld",

		// snap plugins
		snap.CaffeInferenceCollector,
		snap.DockerCollector,
		snap.MutilateCollector,
		snap.SPECjbbCollector,
		snap.CassandraPublisher,
		snap.FilePublisher,
		snap.SessionPublisher,

		// snap.RDTCollector - not yet available

		// workloads
		"memcached",
		"mutilate",

		// kubernetes
		"hyperkube",
	}

	Convey("Make sure all depedencies are there", t, func() {
		for _, executable := range requiredExecutables {
			path := testhelpers.AssertFileExists(executable)
			Println()
			Printf(" %s found in: %s ", executable, path)
		}
	})

}
