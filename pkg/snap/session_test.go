// +build integration

package snap

import (
	"errors"
	"fmt"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"testing"
	"time"

	. "github.com/smartystreets/goconvey/convey"
)

type Snapd struct {
	taskHandle executor.TaskHandle
}

func NewSnapd() *Snapd {
	return &Snapd{}
}

func (s *Snapd) Execute() error {
	l := executor.NewLocal()
	gopath := os.Getenv("GOPATH")
	if gopath == "" {
		return errors.New("Cannot find GOPATH")
	}

	snapRoot := path.Join(gopath, "src", "github.com", "intelsdi-x", "snap", "build", "bin", "snapd")
	snapCommand := fmt.Sprintf("%s -t 0", snapRoot)

	taskHandle, err := l.Execute(snapCommand)
	if err != nil {
		return err
	}

	s.task = taskHandle

	return nil
}

func (s *Snapd) Stop() error {
	if s.task == nil {
		return errors.New("Snapd not started: cannot find task")
	}

	return s.task.Stop()
}

func (s *Snapd) CleanAndEraseOutput() error {
	if s.task == nil {
		return errors.New("Snapd not started: cannot find task")
	}

	s.task.Clean()
	return s.task.EraseOutput()
}

func (s *Snapd) Connected() bool {
	retries := 5
	connected := false
	for i := 0; i < retries; i++ {
		conn, err := net.Dial("tcp", "127.0.0.1:8181")
		if err != nil {
			time.Sleep(100 * time.Millisecond)
			continue
		}
		defer conn.Close()
		connected = true
	}

	return connected
}

func TestSnap(t *testing.T) {
	var snapd *Snapd
	var c *client.Client
	var s *Session
	var publisher *wmap.PublishWorkflowMapNode
	var metricsFile string

	Convey("Testing snap session", t, func() {
		Convey("Starting snapd", func() {
			snapd = NewSnapd()
			snapd.Execute()

			// Wait until snap is up.
			So(snapd.Connected(), ShouldBeTrue)
		})

		Convey("Connecting to snapd", func() {
			ct, err := client.New("http://127.0.0.1:8181", "v1", true)

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})

			c = ct
		})

		Convey("Loading collectors", func() {
			plugins := NewPlugins(c)
			So(plugins, ShouldNotBeNil)
			err := plugins.Load("snap-plugin-collector-session-test")
			So(err, ShouldBeNil)

			// Wait until metric is available in namespace.
			retries := 10
			found := false
			for i := 0; i < retries && !found; i++ {
				m := c.GetMetricCatalog()
				So(m.Err, ShouldBeNil)
				for _, metric := range m.Catalog {
					if metric.Namespace == "/intel/swan/session/metric1" {
						found = true
						break
					}
				}
				time.Sleep(500 * time.Millisecond)
			}
			So(found, ShouldBeTrue)
		})

		Convey("Loading publishers", func() {
			plugins := NewPlugins(c)
			So(plugins, ShouldNotBeNil)

			plugins.Load("snap-plugin-publisher-session-test")

			publisher = wmap.NewPublishNode("session-test", 1)

			So(publisher, ShouldNotBeNil)

			tmpfile, err := ioutil.TempFile("", "session_test")
			tmpfile.Close()
			So(err, ShouldBeNil)

			metricsFile = tmpfile.Name()

			publisher.AddConfigItem("file", metricsFile)
		})

		Convey("Creating a Snap experiment session", func() {
			s = NewSession([]string{"/intel/swan/session/metric1"}, 1*time.Second, c, publisher)
			So(s, ShouldNotBeNil)
		})

		Convey("Starting a session", func() {
			So(s, ShouldNotBeNil)
			err := s.Start("foobar", "barbaz")

			Convey("Shouldn't return any errors", func() {
				So(err, ShouldBeNil)
			})
		})

		Convey("Contacting snap to get the task status", func() {
			status, err := s.Status()

			So(err, ShouldBeNil)

			Convey("And the task should be running", func() {
				So(status, ShouldEqual, "Running")
			})
		})

		Convey("Reading samples from file", func() {
			retries := 5
			found := false
			for i := 0; i < retries; i++ {
				time.Sleep(500 * time.Millisecond)

				dat, err := ioutil.ReadFile(metricsFile)
				if err != nil {
					continue
				}

				if len(dat) > 0 {
					// Look for tag on metric line.
					lines := strings.Split(string(dat), "\n")
					if len(lines) < 1 {
						continue
					}

					columns := strings.Split(lines[0], "\t")
					if len(columns) < 2 {
						continue
					}

					tags := strings.Split(columns[1], ",")
					if len(tags) < 2 {
						continue
					}

					So(columns[0], ShouldEqual, "/intel/swan/session/metric1")
					So(tags[0], ShouldBeIn, "swan_experiment=foobar", "swan_phase=barbaz")
					So(tags[1], ShouldBeIn, "swan_experiment=foobar", "swan_phase=barbaz")

					found = true

					break
				}
			}

			So(found, ShouldBeTrue)
		})

		Convey("Stopping a session", func() {
			So(s, ShouldNotBeNil)
			err := s.Stop()

			So(err, ShouldBeNil)

			_, err = s.Status()
			So(err, ShouldNotBeNil)
		})

		Convey("Stopping snapd", func() {
			So(snapd, ShouldNotBeNil)

			if snapd != nil {
				snapd.Stop()
				snapd.CleanAndEraseOutput()
			}
		})
	})
}
