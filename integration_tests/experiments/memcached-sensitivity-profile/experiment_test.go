/*
go test -c -i ./integration_tests/experiments/memcached-sensitivity-profile/
go install ./experiments/memcached-sensitivity-profile/...
docker inspect -f '{{.State.Status}}' cassandra-swan
find -name 'local_*' | sudo xargs rm -rf
sudo pkill -9 memcached

(
sudo docker rm -f `sudo docker ps -q -a -f "name=k8s_"` || true
sudo pkill -9 -e kube
sudo systemctl stop snapd
etcdctl rm --recursive --dir registry
etcdctl rm --recursive --dir swan
make build
./scripts/isolate-pid.sh go test -v ./integration_tests/experiments/memcached-sensitivity-profile
)
*/
package experiment

import (
	"io"
	"os"
	"os/exec"
	"path"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/gocql/gocql"
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

func kubectlWaitFor(args, expected string, timeoutSec int, t *testing.T) {
	kubectlPath, err := exec.LookPath("kubectl")
	if err != nil {
		t.Fatal(err)
	}

	timeout := time.After(time.Duration(timeoutSec) * time.Second)
	for {
		time.Sleep(1 * time.Second)
		kubectl := exec.Command(kubectlPath, strings.Fields(args)...)
		output, err := kubectl.Output()
		if err == nil {
			log.Infof("%q output (expecting: %q): %s", args, expected, string(output))
			if !strings.Contains(string(output), expected) {
				goto stillNotFound
			}
			log.Infof("%q found in output of %q", expected, args)
			break
		} else {
			log.Errorf("error from %q: %s", args, err)
		}
	stillNotFound:
		select {
		case <-timeout:
			t.Fatalf("k8s is no ready! cannot find %q in %q", expected, string(output))
		default:
		}
	}
}

func TestExperiment(t *testing.T) {
	log.SetLevel(log.DebugLevel)
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

		Reset(func() {
			err := snapd.Process.Kill()
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

			Convey("with k8s with aggressor with correct QoS classes", func() {

				args := []string{"--run_on_kubernetes",
					"--aggr", "l1d", // at least one BE
					"--kube_loglevel=4", // debuggin k8s - check local_*/stderr files
					"--load_duration=5s",
					"--memcached_path=/opt/gopath/src/github.com/intelsdi-x/swan/workloads/data_caching/memcached/memcached-1.4.25/build/memcached", // docker path are different
					"--l1d_path=/opt/gopath/src/github.com/intelsdi-x/swan/workloads/low-level-aggressors/l1d",
				}
				log.Debug("args:", args)

				exp := exec.Command(memcachedSensitivityProfileBin, args...)

				// enable experiment outout proxing when debug is on
				if log.GetLevel() == log.DebugLevel {
					errPipe, err := exp.StderrPipe()
					So(err, ShouldBeNil)
					go func() {
						io.Copy(log.StandardLogger().Out, errPipe)
					}()
				}

				err = exp.Start()
				So(err, ShouldBeNil)

				hostname, _ := os.Hostname()
				kubectlWaitFor("get cs", "Healthy", 30, t)
				kubectlWaitFor("get nodes", hostname, 30, t)
				kubectlWaitFor("get pods", "swan-hp", 30, t)
				kubectlWaitFor("get pods", "swan-aggr", 60, t)
				kubectlWaitFor("get pods", "Running", 30, t)
				kubectlWaitFor("describe pod swan-hp", "Guaranteed", 30, t)
				kubectlWaitFor("describe pod swan-aggr", "BestEffort", 30, t)

				err = exp.Wait()
				So(err, ShouldBeNil)
			})

			return

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
