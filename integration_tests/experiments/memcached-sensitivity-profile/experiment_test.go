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

	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
	"github.com/intelsdi-x/swan/integration_tests/test_helpers"
	"github.com/intelsdi-x/swan/pkg/experiment"
	. "github.com/smartystreets/goconvey/convey"
)

func getUUID(outs []byte) string {
	So(outs, ShouldNotBeNil)
	lines := strings.Split(string(outs), "\n")
	So(len(lines), ShouldBeGreaterThan, 0)
	return string(lines[0])
}

func runExp(command string, dumpOutputOnError bool, args ...string) (string, error) {
	c := exec.Command(command, args...)
	b := &bytes.Buffer{}
	c.Stderr = b
	out, err := c.Output()

	// Extra logs vs asked explictly.
	log.Debugf("[Out]==> %s", string(out))
	log.Debugf("[Err]==> %s", b.String())
	log.Debugf("[Warning]==> %s", err)

	if err != nil {
		if dumpOutputOnError {
			Printf("[Out]==> %s", string(out))
			Printf("[Err]==> %s", b.String())
			Printf("[Warning]==> %s", err)
		}
		return "", err
	}

	return getUUID(out), nil
}

func TestExperiment(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	// Use experiment binaries from build directory to simplify development flow (doesn't required make bist install).
	memcachedSensitivityProfileBin := path.Join(testhelpers.SwanPath, "build/experiments/memcached/memcached-sensitivity-profile")

	envs := map[string]string{
		"SWAN_LOG":                  "debug",
		"SWAN_BE_RANGE":             "0",
		"SWAN_HP_RANGE":             "0",
		"SWAN_REPS":                 "1",
		"SWAN_LOAD_POINTS":          "1",
		"SWAN_PEAK_LOAD":            "5000",
		"SWAN_LOAD_DURATION":        "1s",
		"SWAN_MUTILATE_WARMUP_TIME": "1s",
	}

	Convey("With environment prepared for experiment", t, func() {
		for k, v := range envs {
			os.Setenv(k, v)
		}

		session, err := getCassandraSession()
		So(err, ShouldBeNil)
		defer session.Close()

		loadDataFromCassandra := func(experimentID string) (tags map[string]string, swanRepetitions, swanAggressorsNames, swanPhases []string) {

			metricsCount := 0
			var ns string
			iter := session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Iter()
			for iter.Scan(&ns, &tags) {
				metricsCount++
				log.Debugf("experimentID=%s ns=%s tags=%#v", experimentID, ns, tags)
				So(ns, ShouldNotBeBlank)
				So(tags, ShouldNotBeEmpty)
				So(tags, ShouldContainKey, "swan_phase")
				So(tags, ShouldContainKey, "swan_aggressor_name")
				swanAggressorsNames = append(swanAggressorsNames, tags["swan_aggressor_name"])
				swanPhases = append(swanPhases, tags["swan_phase"])
				swanRepetitions = append(swanRepetitions, tags["swan_repetition"])
			}
			log.Debugf("%d metrics found in cassandra for experiment id %q", metricsCount, experimentID)
			So(metricsCount, ShouldBeGreaterThan, 0)

			So(iter.Close(), ShouldBeNil)
			return
		}

		Convey("With proper configuration and without aggressor phases", func() {
			_, err := runExp(memcachedSensitivityProfileBin, true)

			Convey("Experiment should return with no errors", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("With proper configuration and with l1d aggressors", func() {
			args := []string{"-aggr", "l1d"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				_, _, swanAggressorsNames, _ := loadDataFromCassandra(experimentID)
				So("None", ShouldBeIn, swanAggressorsNames)
				So("L1 Data", ShouldBeIn, swanAggressorsNames)

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
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				_, swanRepetitions, swanAggressorsNames, _ := loadDataFromCassandra(experimentID)

				So("L1 Data", ShouldBeIn, swanAggressorsNames)
				So("None", ShouldBeIn, swanAggressorsNames)

				So("0", ShouldBeIn, swanRepetitions)
				So("1", ShouldBeIn, swanRepetitions)

				So(swanAggressorsNames, ShouldHaveLength, 36)
				So(swanRepetitions, ShouldHaveLength, 36)

			})

			Convey("Experiment should succeed also with 2 load points", func() {
				os.Setenv("SWAN_LOAD_POINTS", "2")
				fmt.Println(args)
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, swanAggressorsNames, swanPhases := loadDataFromCassandra(experimentID)

				So(tags["swan_repetition"], ShouldEqual, "0")

				So(swanAggressorsNames, ShouldHaveLength, 36)
				So(swanPhases, ShouldHaveLength, 36)

				So("L1 Data", ShouldBeIn, swanAggressorsNames)
				So("None", ShouldBeIn, swanAggressorsNames)

				So("Aggressor None; load point 0;", ShouldBeIn, swanPhases)

			})
		})

		Convey("With proper kubernetes configuration and without phases", func() {
			args := []string{"-kubernetes", "-kube_allow_privileged"}
			_, err := runExp(memcachedSensitivityProfileBin, true, args...)
			Convey("Experiment should return with no errors", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("With proper kubernetes configuration and with l1d aggressor", func() {
			args := []string{"-kubernetes", "-aggr", "l1d", "-baseline=false", "-kube_allow_privileged"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, swanAggressorsNames, _ := loadDataFromCassandra(experimentID)
				So(tags["swan_aggressor_name"], ShouldEqual, "L1 Data")
				So("None", ShouldNotBeIn, swanAggressorsNames)
			})
		})

		Convey("With proper kubernetes and caffe", func() {
			args := []string{"-kubernetes", "-aggr", "caffe", "-baseline=false", "-kube_allow_privileged"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, _, _ := loadDataFromCassandra(experimentID)
				So(tags["swan_aggressor_name"], ShouldEqual, "Caffe")
			})
		})

		Convey("With invalid configuration stop experiment if error", func() {
			os.Setenv("SWAN_LOAD_POINTS", "abc")
			_, err := runExp(memcachedSensitivityProfileBin, false)
			So(err, ShouldNotBeNil)
		})

		Convey("While setting zero repetitions to phase", func() {
			args := []string{"-aggr", "l1d"}
			os.Setenv("SWAN_LOAD_POINTS", "1")
			os.Setenv("SWAN_REPS", "0")
			Convey("Experiment should pass with no errors", func() {
				_, err := runExp(memcachedSensitivityProfileBin, false, args...)
				So(err, ShouldBeNil)
			})
		})

		Convey("With wrong aggresor name", func() {
			args := []string{"-aggr", "not-existing-aggressor"}
			_, err := runExp(memcachedSensitivityProfileBin, false, args...)
			So(err, ShouldNotBeNil)
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
