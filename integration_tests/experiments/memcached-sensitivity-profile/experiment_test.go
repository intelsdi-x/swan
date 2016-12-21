package experiment

import (
	"bytes"
	"fmt"
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
	snapLogs = "/tmp/swan-integration-tests"
)

func getUUID(outs []byte) string {
	So(outs, ShouldNotBeNil)
	lines := strings.Split(string(outs), "\n")
	So(len(lines), ShouldBeGreaterThan, 0)
	return string(lines[0])
}

func runExp(command string, args ...string) (string, error) {
	c := exec.Command(command, args...)
	b := &bytes.Buffer{}
	c.Stderr = b
	out, err := c.Output()
	if err != nil {
		Printf("[Warning]==> %s", err)
		return "", err
	}

	Printf("[Out]==> %s", string(out))
	Printf("[Err]==> %s", b.String())

	return getUUID(out), nil
}

func TestExperiment(t *testing.T) {
	memcachedSensitivityProfileBin := path.Join(fs.GetSwanBuildPath(), "experiments", "memcached", "memcached-sensitivity-profile")
	memcacheDockerBin := "memcached"
	l1dDockerBin := "l1d"

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

	Convey("When snapteld is launched", t, func() {
		var logDirPerm os.FileMode = 0755
		err := os.MkdirAll(snapLogs, logDirPerm)
		So(err, ShouldBeNil)

		// Snapteld default ports are 8181(API port) and 8082(RPC port).
		snapteld := testhelpers.NewSnapteldOnPort(8181, 8082)
		err = snapteld.Start()
		So(err, ShouldBeNil)

		time.Sleep(1 * time.Second)

		Reset(func() {
			err := snapteld.Stop()
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
				_, err := runExp(memcachedSensitivityProfileBin)

				Convey("Experiment should return with no errors", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("With proper configuration and with l1d aggressors", func() {
				args := []string{"--aggr", "l1d"}
				Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
					experimentID, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)

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
					experimentID, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)

					var ns string
					var tags map[string]string
					var swanRepetitions []string
					var swanAggressorsNames []string
					iter := session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Iter()
					for iter.Scan(&ns, &tags) {
						So(ns, ShouldNotBeBlank)
						So(tags, ShouldNotBeEmpty)
						swanAggressorsNames = append(swanAggressorsNames, tags["swan_aggressor_name"])
						swanRepetitions = append(swanRepetitions, tags["swan_repetition"])
					}

					So("L1 Data", ShouldBeIn, swanAggressorsNames)
					So("None", ShouldBeIn, swanAggressorsNames)

					So("0", ShouldBeIn, swanRepetitions)
					So("1", ShouldBeIn, swanRepetitions)

					So(swanAggressorsNames, ShouldHaveLength, 40)
					So(swanRepetitions, ShouldHaveLength, 40)

					So(iter.Close(), ShouldBeNil)

				})

				Convey("Experiment should succeed also with 2 load points", func() {
					os.Setenv("SWAN_LOAD_POINTS", "2")
					fmt.Println(args)
					experimentID, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)

					var ns string
					var tags map[string]string
					var swanAggressorsNames []string
					var swanPhases []string
					iter := session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Iter()
					for iter.Scan(&ns, &tags) {
						So(ns, ShouldNotBeBlank)
						So(tags, ShouldNotBeEmpty)
						So(tags, ShouldContainKey, "swan_phase")
						So(tags, ShouldContainKey, "swan_aggressor_name")
						swanAggressorsNames = append(swanAggressorsNames, tags["swan_aggressor_name"])
						swanPhases = append(swanPhases, tags["swan_phase"])
						So(tags["swan_repetition"], ShouldEqual, "0")
					}

					So(swanAggressorsNames, ShouldHaveLength, 40)
					So(swanPhases, ShouldHaveLength, 40)

					So("L1 Data", ShouldBeIn, swanAggressorsNames)
					So("None", ShouldBeIn, swanAggressorsNames)

					So("Aggressor None; load point 0;", ShouldBeIn, swanPhases)

					So(iter.Close(), ShouldBeNil)
				})
			})

			SkipConvey("With proper kubernetes configuration and without phases", func() {
				args := []string{"--run_on_kubernetes", "--kube_allow_privileged", "--memcached_path", memcacheDockerBin}
				_, err := runExp(memcachedSensitivityProfileBin, args...)
				Convey("Experiment should return with no errors", func() {
					So(err, ShouldBeNil)
				})
			})

			SkipConvey("With proper kubernetes configuration and with l1d aggressor", func() {
				args := []string{"--run_on_kubernetes", "--kube_allow_privileged", "--aggr", "l1d", "--memcached_path", memcacheDockerBin, "--l1d_path", l1dDockerBin}
				Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
					experimentID, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)

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
				_, err := runExp(memcachedSensitivityProfileBin)
				So(err, ShouldNotBeNil)
			})

			Convey("While setting zero repetitions to phase", func() {
				args := []string{"--aggr", "l1d"}
				os.Setenv("SWAN_LOAD_POINTS", "1")
				os.Setenv("SWAN_REPS", "0")
				Convey("Experiment should pass with no errors", func() {
					_, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)
				})
			})

			Convey("With wrong aggresor name", func() {
				args := []string{"--aggr", "l1e"}
				_, err := runExp(memcachedSensitivityProfileBin, args...)
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
