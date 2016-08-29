package experiment

import (
	"os"
	"os/exec"
	"path"
	"testing"
	"time"
	"strings"

	"github.com/gocql/gocql"
	"github.com/intelsdi-x/athena/pkg/utils/fs"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	snapBuildPath = "src/github.com/intelsdi-x/snap/build"
	snapLogs      = "/tmp/swan-integration-tests"
)

func getUUID(outs []byte) string {
	lines := strings.Split(string(outs), "\n")
	return string(lines[0])
}

func TestExperiment(t *testing.T) {
	memcachedSensitivityProfileBin := path.Join(fs.GetSwanBuildPath(), "experiments", "memcached", "memcached-sensitivity-profile")
	snapdBin := path.Join(os.Getenv("GOPATH"), snapBuildPath, "bin", "snapd")

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

	Convey("When snapd is launched", t, func() {
		var logDirPerm os.FileMode = 0755
		err := os.MkdirAll(snapLogs, logDirPerm)
		So(err, ShouldBeNil)
		snapd := exec.Command(snapdBin, "--plugin-trust=0", "--log-level=1", "--log-path", snapLogs)

		err = snapd.Start()
		So(err, ShouldBeNil)
		time.Sleep(1 * time.Second)

		Convey("While doing experiment", func() {
			for k, v := range envs {
				os.Setenv(k, v)
			}

			Convey("With proper configuration and empty phases", func() {
				exp := exec.Command(memcachedSensitivityProfileBin)

				err := exp.Run()

				Convey("Experiment should return with no errors", func() {
					So(err, ShouldBeNil)
				})
			})

			Convey("With proper configuration and not empty phases", func() {
				args := []string{"--aggr", "l1d"}
				exp := exec.Command(memcachedSensitivityProfileBin, args...)

				Convey("Experiment should be stored in a Cassandra DB", func() {
					var ns string
					output, err := exp.Output()
					So(err, ShouldBeNil)
					experimentID := getUUID(output)
					cluster := gocql.NewCluster("127.0.0.1")
					cluster.Keyspace = "snap"
					cluster.ProtoVersion = 4
					cluster.Timeout = 100 * time.Second
					session, err := cluster.CreateSession()

					defer session.Close()

					session.Query(`SELECT ns FROM snap.metrics WHERE tags['swan_experiment'] = '?' ALLOW FILTERING`, experimentID).Scan(&ns)
					So(ns, ShouldNotBeNil)
				})

				Convey("While having two repetitions to phase", func() {
					os.Setenv("SWAN_REPS", "2")
					exp := exec.Command(memcachedSensitivityProfileBin, args...)
					err := exp.Run()
					os.Setenv("SWAN_REPS", "1")
					So(err, ShouldBeNil)

					Convey("Experiment should succeed also with 2 load points", func() {
						os.Setenv("SWAN_LOAD_POINTS", "2")
						exp := exec.Command(memcachedSensitivityProfileBin, args...)
						err := exp.Run()
						So(err, ShouldBeNil)
					})
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

		Reset(func() {
			err := snapd.Process.Kill()
			So(err, ShouldBeNil)
			if err == nil {
				os.RemoveAll(snapLogs)
			}
		})
	})
}
