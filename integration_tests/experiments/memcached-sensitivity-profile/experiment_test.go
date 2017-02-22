package experiment

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/utils/err_collection"
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
		Printf("[Out]==> %s", string(out))
		Printf("[Err]==> %s", b.String())
		Printf("[Warning]==> %s", err)
		return "", err
	}
	return getUUID(out), nil
}

func TestExperiment(t *testing.T) {

	memcachedSensitivityProfileBin := testhelpers.AssertFileExists("memcached-sensitivity-profile")

	envs := map[string]string{
		"SWAN_LOG":                  "debug",
		"SWAN_BE_SETS":              "0:0",
		"SWAN_HP_SETS":              "0:0",
		"SWAN_REPS":                 "1",
		"SWAN_LOAD_POINTS":          "1",
		"SWAN_PEAK_LOAD":            "5000",
		"SWAN_LOAD_DURATION":        "1s",
		"SWAN_MUTILATE_WARMUP_TIME": "1s",
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
			var errCollection errcollection.ErrorCollection
			errCollection.Add(snapteld.Stop())
			errCollection.Add(snapteld.CleanAndEraseOutput())
			So(errCollection.GetErrIfAny(), ShouldBeNil)
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
				args := []string{"-aggr", "l1d"}
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

					// Check metadata was saved.
					var (
						metadata     = make(map[string]string)
						iterMetadata map[string]string
					)

					iter := session.Query(`SELECT metadata FROM swan.metadata WHERE experiment_id = ? ALLOW FILTERING`, experimentID).Iter()
					for iter.Scan(&iterMetadata) {
						So(iterMetadata, ShouldNotBeEmpty)
						for k, v := range iterMetadata {
							metadata[k] = v
						}
					}
					So(err, ShouldBeNil)
					So(metadata, ShouldNotBeEmpty)
					So(metadata["SWAN_PEAK_LOAD"], ShouldEqual, "5000")
					So(metadata["load_points"], ShouldEqual, "1")
					So(metadata["load_duration"], ShouldEqual, "1s")
					So(metadata[experiment.CPUModelNameKey], ShouldNotEqual, "")

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

					So(swanAggressorsNames, ShouldHaveLength, 36)
					So(swanRepetitions, ShouldHaveLength, 36)

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

					So(swanAggressorsNames, ShouldHaveLength, 36)
					So(swanPhases, ShouldHaveLength, 36)

					So("L1 Data", ShouldBeIn, swanAggressorsNames)
					So("None", ShouldBeIn, swanAggressorsNames)

					So("Aggressor None; load point 0;", ShouldBeIn, swanPhases)

					So(iter.Close(), ShouldBeNil)
				})
			})

			Convey("With proper kubernetes configuration and without phases", func() {
				args := []string{"-kubernetes", "-kube_allow_privileged"}
				_, err := runExp(memcachedSensitivityProfileBin, args...)
				Convey("Experiment should return with no errors", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("With proper kubernetes configuration and with l1d aggressor", func() {
				args := []string{"-kubernetes", "-kube_allow_privileged", "-aggr", "l1d"}
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

			Convey("With proper kubernetes and caffe", func() {
				args := []string{"-kubernetes", "-aggr", "caffe"}
				Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
					experimentID, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)

					var ns string
					var tags map[string]string
					err = session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Scan(&ns, &tags)
					So(err, ShouldBeNil)
					So(ns, ShouldNotBeBlank)
					So(tags, ShouldNotBeEmpty)
					So(tags["swan_aggressor_name"], ShouldEqual, "Caffe")
				})
			})

			Convey("With invalid configuration stop experiment if error", func() {
				os.Setenv("SWAN_LOAD_POINTS", "abc")
				_, err := runExp(memcachedSensitivityProfileBin)
				So(err, ShouldNotBeNil)
			})

			Convey("While setting zero repetitions to phase", func() {
				args := []string{"-aggr", "l1d"}
				os.Setenv("SWAN_LOAD_POINTS", "1")
				os.Setenv("SWAN_REPS", "0")
				Convey("Experiment should pass with no errors", func() {
					_, err := runExp(memcachedSensitivityProfileBin, args...)
					So(err, ShouldBeNil)
				})
			})

			Convey("With wrong aggresor name", func() {
				args := []string{"-aggr", "l1e"}
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
