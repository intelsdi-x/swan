package experiment

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/athena/integration_tests/test_helpers"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	snapBuildPath = "src/github.com/intelsdi-x/snap/build"
	snapLogs      = "/tmp/swan-integration-tests"
)

func getUUID(outs []byte) string {
	So(outs, ShouldNotBeNil)
	lines := strings.Split(string(outs), "\n")
	So(len(lines), ShouldBeGreaterThan, 0)
	return string(lines[0])
}

func TestExperiment(t *testing.T) {
	memcachedSensitivityProfileBin := path.Join(fs.GetSwanBuildPath(), "experiments", "memcached", "memcached-sensitivity-profile")

	envs := map[string]string{
		"SWAN_LOG":                  "debug",
		"SWAN_BE_SETS":              "0:0",
		"SWAN_HP_SETS":              "0:0",
		"SWAN_REPS":                 "1",
		"SWAN_LOAD_POINTS":          "1",
		"SWAN_PEAK_LOAD":            "5000",
		"SWAN_LOAD_DURATION":        "1s",
		"SWAN_MUTILATE_WARMUP_TIME": "1s",
		"SWAN_KUBE_APISERVER_PATH":  path.Join(fs.GetAthenaBinPath(), "kube-apiserver"),
		"SWAN_KUBE_CONTROLLER_PATH": path.Join(fs.GetAthenaBinPath(), "kube-controller-manager"),
		"SWAN_KUBELET_PATH":         path.Join(fs.GetAthenaBinPath(), "kubelet"),
		"SWAN_KUBE_PROXY_PATH":      path.Join(fs.GetAthenaBinPath(), "kube-proxy"),
		"SWAN_KUBE_SCHEDULER_PATH":  path.Join(fs.GetAthenaBinPath(), "kube-scheduler"),
	}

	Convey("When snapd is launched", t, func() {
		var logDirPerm os.FileMode = 0755
		err := os.MkdirAll(snapLogs, logDirPerm)
		So(err, ShouldBeNil)

		snapd := testhelpers.NewSnapdOnPort(8181)
		err = snapd.Start()
		So(err, ShouldBeNil)

		time.Sleep(1 * time.Second)

		Reset(func() {
			err := snapd.Stop()
			So(err, ShouldBeNil)
			if err == nil {
				os.RemoveAll(snapLogs)
			}
		})

		Convey("While doing experiment", func() {
			for k, v := range envs {
				os.Setenv(k, v)
			}

			session, err := getCassandraSession()
			So(err, ShouldBeNil)
			defer session.Close()

			Convey("With proper configuration and without aggressor phases", func() {
				exp := exec.Command(memcachedSensitivityProfileBin)
				err := exp.Run()

				Convey("Experiment should return with no errors", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("With proper configuration and with l1d aggressors", func() {
				args := []string{"--aggr", "l1d"}
				Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
					exp := exec.Command(memcachedSensitivityProfileBin, args...)
					output, err := exp.Output()
					So(err, ShouldBeNil)
					experimentID := getUUID(output)

					var ns string
					var tags map[string]string
					err = session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Scan(&ns, &tags)
					So(err, ShouldBeNil)
					So(ns, ShouldNotBeBlank)
					So(tags, ShouldNotBeEmpty)
					So(tags["swan_aggressor_name"], ShouldEqual, "L1 Data")
				})

				Convey("While having two repetitions to phase", func() {
					os.Setenv("SWAN_REPS", "2")
					exp := exec.Command(memcachedSensitivityProfileBin, args...)
					output, err := exp.Output()
					So(err, ShouldBeNil)
					experimentID := getUUID(output)

					var ns string
					var tags map[string]string
					err = session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Scan(&ns, &tags)
					So(err, ShouldBeNil)
					So(ns, ShouldNotBeBlank)
					So(tags, ShouldNotBeEmpty)
					So(tags["swan_aggressor_name"], ShouldEqual, "L1 Data")
					So(tags["swan_repetition"], ShouldEqual, "1")
				})

				Convey("Experiment should succeed also with 2 load points", func() {
					os.Setenv("SWAN_LOAD_POINTS", "2")
					exp := exec.Command(memcachedSensitivityProfileBin, args...)
					output, err := exp.Output()
					So(err, ShouldBeNil)
					experimentID := getUUID(output)

					var ns string
					var tags map[string]string
					err = session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Scan(&ns, &tags)
					So(err, ShouldBeNil)
					So(ns, ShouldNotBeBlank)
					So(tags, ShouldNotBeEmpty)
					So(tags["swan_aggressor_name"], ShouldEqual, "L1 Data")
					So(tags["swan_phase"], ShouldEqual, "aggressor_nr_0_measurement_for_loadpoint_id_2")
				})
			})

			Convey("With proper kubernetes configuration and without phases", func() {
				args := []string{"--run_on_kubernetes", "--kube_allow_privileged"}
				exp := exec.Command(memcachedSensitivityProfileBin, args...)
				err := exp.Run()
				remoteDebugTest(exp)
				Convey("Experiment should return with no errors", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("With proper kubernetes configuration and with l1d aggressor", func() {
				args := []string{"--run_on_kubernetes", "--kube_allow_privileged", "--aggr", "l1d"}
				Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
					exp := exec.Command(memcachedSensitivityProfileBin, args...)
					output, err := exp.Output()
					So(err, ShouldBeNil)
					experimentID := getUUID(output)

					var ns string
					var tags map[string]string
					err = session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Scan(&ns, &tags)
					So(err, ShouldBeNil)
					So(ns, ShouldNotBeBlank)
					So(tags, ShouldNotBeEmpty)
					So(tags["swan_aggressor_name"], ShouldEqual, "L1 Data")
				})

			})

			Convey("With invalid configuration stop experiment if error", func() {
				os.Setenv("SWAN_LOAD_POINTS", "abc")
				exp := exec.Command(memcachedSensitivityProfileBin)
				err := exp.Run()

				So(err, ShouldNotBeNil)
			})

			Convey("While setting zero repetitions to phase", func() {
				args := []string{"--aggr", "l1d"}
				os.Setenv("SWAN_LOAD_POINTS", "1")
				os.Setenv("SWAN_REPS", "0")
				Convey("Experiment should pass with no errors", func() {
					exp := exec.Command(memcachedSensitivityProfileBin, args...)
					err := exp.Run()
					So(err, ShouldBeNil)
				})
			})

			Convey("With wrong aggresor name", func() {
				args := []string{"--aggr", "l1e"}
				exp := exec.Command(memcachedSensitivityProfileBin, args...)
				err := exp.Run()
				So(err, ShouldNotBeNil)
			})
		})
	})
}

func getCassandraSession() (*gocql.Session, error) {
	cluster := gocql.NewCluster("127.0.0.1")
	cluster.Keyspace = "snap"
	cluster.ProtoVersion = 4
	cluster.Timeout = 100 * time.Second
	session, err := cluster.CreateSession()
	return session, err
}

func remoteDebugTest(exp *exec.Cmd) {
	stdoutPipe, _ := exp.StdoutPipe()
	stderrPipe, _ := exp.StderrPipe()
	if stdoutPipe == nil || stderrPipe == nil {
		return
	}
	stdoutBuffor, _ := ioutil.ReadAll(stdoutPipe)
	stderrBuffor, _ := ioutil.ReadAll(stderrPipe)
	fmt.Println(stdoutBuffor)
	fmt.Println(stderrBuffor)
}
