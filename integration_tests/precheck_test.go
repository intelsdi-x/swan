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
