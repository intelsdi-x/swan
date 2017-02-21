package precheck

import (
	"testing"

	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	. "github.com/smartystreets/goconvey/convey"
)

func TestFunction(t *testing.T) {

	requiredExecutables := []string{

		// aggressors
		"caffe",
		"l1d",
		"l1i",
		"l3",
		"memBw",
		"stream.100M",

		// experiments
		"memcached-sensitivity-profile",
		"specjbb-sensitivity-profile",

		// snap
		"snaptel",
		"snapteld",

		// snap plugins
		"snap-plugin-collector-caffe-inference",
		"snap-plugin-collector-docker",
		"snap-plugin-collector-mutilate",
		"snap-plugin-collector-specjbb",
		"snap-plugin-processor-tag",
		"snap-plugin-publisher-cassandra",
		"snap-plugin-publisher-file",
		"snap-plugin-publisher-session-test",

		// workloads
		"memcached",
		"mutilate",

		// kubernetes
		"apiserver",
		"controller-manager",
		"federation-apiserver",
		"federation-controller-manager",
		"hyperkube",
		"kubectl",
		"kubelet",
		"proxy",
		"scheduler",
	}

	Convey("Make sure all depedencies are there", t, func() {
		for _, executable := range requiredExecutables {
			path := testhelpers.AssertFileExists(executable)
			Println()
			Printf(" %s found in: %s ", executable, path)
		}
	})

}
