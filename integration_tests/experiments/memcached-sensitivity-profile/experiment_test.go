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

package experiment

import (
	"bytes"
	"fmt"
	"io/ioutil"
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
	env := ""
	for _, e := range os.Environ() {
		if strings.Contains(e, "SWAN_") {
			env += e + " "
		}
	}
	fullCommand := "sudo -E env PATH=$PATH " + env + " " + command + " " + strings.Join(args, " ")
	// Extra logs vs asked explictly.
	log.Debugf("[FullCommand]==> %q", fullCommand)

	c := exec.Command(command, args...)
	b := &bytes.Buffer{}
	c.Stderr = b
	out, err := c.Output()

	log.Debugf("[Out]==> %s", string(out))
	log.Debugf("[Err]==> %s", b.String())
	log.Debugf("[Warning]==> %s", err)

	if err != nil {
		if dumpOutputOnError {
			Printf("[FullCommand]==> %q", fullCommand)
			Printf("[Out]==> %s", string(out))
			Printf("[Err]==> %s", b.String())
			Printf("[Warning]==> %s", err)
		}
		return "", err
	}

	return getUUID(out), nil
}

func loadDataFromCassandra(session *gocql.Session, experimentID string) (tags map[string]string, swanRepetitions, swanAggressorsNames, swanPhases []string, metricsCount int) {

	var ns string
	iter := session.Query(`SELECT ns, tags FROM snap.metrics WHERE tags['swan_experiment'] = ? ALLOW FILTERING`, experimentID).Iter()
	for iter.Scan(&ns, &tags) {
		metricsCount++
		swanAggressorsNames = append(swanAggressorsNames, tags["swan_aggressor_name"])
		swanPhases = append(swanPhases, tags["swan_phase"])
		swanRepetitions = append(swanRepetitions, tags["swan_repetition"])
	}
	err := iter.Close()
	if err != nil {
		panic(err)
	}
	log.Debugf("experimentID=%s metrics=%d ns=%s tags=%#v", experimentID, metricsCount, ns, tags)

	return
}

// Use experiment binaries from build directory to simplify development flow (doesn't required make bist install).
var memcachedSensitivityProfileBin = path.Join(testhelpers.SwanPath, "build/experiments/memcached/memcached-sensitivity-profile")

func TestExperimentConfiguration(t *testing.T) {
	log.SetLevel(log.DebugLevel)
	const confFilename = "temp_new_config"

	Convey("generated config should contain some flags", t, func() {
		config, err := exec.Command(memcachedSensitivityProfileBin, "-config-dump").Output()
		So(err, ShouldBeNil)

		So(string(config), ShouldContainSubstring, "KUBERNETES=false")

		Convey("and after replace new value is dumped", func() {
			newConfig := strings.Replace(string(config), "KUBERNETES=false", "KUBERNETES=true", -1)

			err = ioutil.WriteFile(confFilename, []byte(newConfig), os.ModePerm)
			So(err, ShouldBeNil)

			reloadedConfig, err := exec.Command(memcachedSensitivityProfileBin, "-config", confFilename, "-config-dump").CombinedOutput()
			So(err, ShouldBeNil)
			So(string(reloadedConfig), ShouldContainSubstring, "KUBERNETES=true")

			Reset(func() { os.Remove(confFilename) })
		})

	})

}

func TestExperiment(t *testing.T) {

	log.SetLevel(log.ErrorLevel)

	envs := map[string]string{
		"SWAN_LOG_LEVEL":                           "debug",
		"SWAN_EXPERIMENT_HP_WORKLOAD_CPU_RANGE":    "0",
		"SWAN_EXPERIMENT_BE_WORKLOAD_L1_CPU_RANGE": "0",
		"SWAN_EXPERIMENT_BE_WORKLOAD_L3_CPU_RANGE": "0",
		"SWAN_EXPERIMENT_REPETITIONS":              "1",
		"SWAN_EXPERIMENT_LOAD_POINTS":              "1",
		"SWAN_EXPERIMENT_PEAK_LOAD":                "5000",
		"SWAN_EXPERIMENT_LOAD_DURATION":            "1s",
		"SWAN_MUTILATE_WARMUP_TIME":                "1s",
	}

	Convey("With environment prepared for experiment", t, func() {
		for k, v := range envs {
			os.Setenv(k, v)
		}

		session, err := getCassandraSession()
		So(err, ShouldBeNil)
		defer session.Close()

		Convey("With proper configuration and without aggressor phases", func() {
			_, err := runExp(memcachedSensitivityProfileBin, true)

			Convey("Experiment should return with no errors", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("With caffe aggressor and baseline", func() {
			args := []string{"-experiment_be_workloads", "caffe"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				_, _, swanAggressorsNames, _, _ := loadDataFromCassandra(session, experimentID)
				So("None", ShouldBeIn, swanAggressorsNames)
				So("Caffe", ShouldBeIn, swanAggressorsNames)
			})
		})

		Convey("With proper configuration and with l1d aggressors", func() {
			args := []string{"-experiment_be_workloads", "l1d"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				_, _, swanAggressorsNames, _, metricsCount := loadDataFromCassandra(session, experimentID)
				So(metricsCount, ShouldBeGreaterThan, 0)
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
				os.Setenv("SWAN_EXPERIMENT_REPETITIONS", "2")
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				_, swanRepetitions, swanAggressorsNames, _, metricsCount := loadDataFromCassandra(session, experimentID)
				So(metricsCount, ShouldBeGreaterThan, 0)

				So("L1 Data", ShouldBeIn, swanAggressorsNames)
				So("None", ShouldBeIn, swanAggressorsNames)

				So("0", ShouldBeIn, swanRepetitions)
				So("1", ShouldBeIn, swanRepetitions)

				So(swanAggressorsNames, ShouldHaveLength, 36)
				So(swanRepetitions, ShouldHaveLength, 36)

			})

			Convey("Experiment should succeed also with 2 load points", func() {
				os.Setenv("SWAN_EXPERIMENT_LOAD_POINTS", "2")
				fmt.Println(args)
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, swanAggressorsNames, swanPhases, metricsCount := loadDataFromCassandra(session, experimentID)
				So(metricsCount, ShouldBeGreaterThan, 0)

				So(tags["swan_repetition"], ShouldEqual, "0")

				So(swanAggressorsNames, ShouldHaveLength, 36)
				So(swanPhases, ShouldHaveLength, 36)

				So("L1 Data", ShouldBeIn, swanAggressorsNames)
				So("None", ShouldBeIn, swanAggressorsNames)

				So("Aggressor None; load point 0;", ShouldBeIn, swanPhases)

			})
		})

		Convey("With proper kubernetes configuration and without phases", func() {
			args := []string{"-kubernetes"}
			_, err := runExp(memcachedSensitivityProfileBin, true, args...)
			Convey("Experiment should return with no errors", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("With proper kubernetes configuration and with l1d aggressor", func() {
			args := []string{"-kubernetes", "-experiment_be_workloads", "l1d", "-experiment_baseline=false"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, swanAggressorsNames, _, metricsCount := loadDataFromCassandra(session, experimentID)
				So(metricsCount, ShouldBeGreaterThan, 0)
				So(tags["swan_aggressor_name"], ShouldEqual, "L1 Data")
				So("None", ShouldNotBeIn, swanAggressorsNames)
			})
		})

		Convey("With proper kubernetes configuration and with stress-ng-stream aggressor", func() {
			args := []string{"-kubernetes", "-experiment_be_workloads", "stress-ng-stream", "-experiment_baseline=false"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, swanAggressorsNames, _, metricsCount := loadDataFromCassandra(session, experimentID)
				So(metricsCount, ShouldBeGreaterThan, 0)
				So(tags["swan_aggressor_name"], ShouldEqual, "stress-ng-stream")
				So("None", ShouldNotBeIn, swanAggressorsNames)
			})
		})

		Convey("With proper kubernetes and caffe", func() {
			args := []string{"-kubernetes", "-experiment_be_workloads", "caffe", "-experiment_baseline=false"}
			Convey("Experiment should run with no errors and results should be stored in a Cassandra DB", func() {
				experimentID, err := runExp(memcachedSensitivityProfileBin, true, args...)
				So(err, ShouldBeNil)

				tags, _, _, _, metricsCount := loadDataFromCassandra(session, experimentID)
				So(metricsCount, ShouldBeGreaterThan, 0)
				So(tags["swan_aggressor_name"], ShouldEqual, "Caffe")
			})
		})

		Convey("With invalid configuration stop experiment if error", func() {
			os.Setenv("SWAN_EXPERIMENT_LOAD_POINTS", "abc")
			_, err := runExp(memcachedSensitivityProfileBin, false)
			So(err, ShouldNotBeNil)
		})

		Convey("While setting zero repetitions to phase", func() {
			args := []string{"-experiment_be_workloads", "l1d"}
			os.Setenv("SWAN_EXPERIMENT_LOAD_POINTS", "1")
			os.Setenv("SWAN_EXPERIMENT_REPETITIONS", "0")
			Convey("Experiment should pass with no errors", func() {
				_, err := runExp(memcachedSensitivityProfileBin, false, args...)
				So(err, ShouldBeNil)
			})
		})

		Convey("With wrong aggresor name", func() {
			args := []string{"-experiment_be_workloads", "not-existing-aggressor"}
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
