package executor

import (
  "testing"
  . "github.com/smartystreets/goconvey/convey"
)

func TestCluster(t *testing.T) {
  Convey("Creating a new cluster", t, func(){
    executor, err := NewCluster("test-foobar", "127.0.0.1")
    So(err, ShouldBeNil)
    So(executor, ShouldNotBeNil)
    _, err = executor.Execute("foobar")
    So(err, ShouldBeNil)

    agent1, err := NewAgent("agent1", "127.0.0.1")
    So(err, ShouldBeNil)

    err = agent1.StealJob()
    So(err, ShouldBeNil)
    So(agent1.CurrentJob, ShouldNotBeNil)
  })
}
