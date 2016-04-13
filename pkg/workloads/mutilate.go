package workloads

import (
	"fmt"
	"github.com/intelsdi-x/swan/pkg/executor"
)

type Mutilate struct {
	executor executor.Executor

	mutilate_threads     int
	mutilate_connections int

	memcached_uri string
}

/**
Real deployments of memcached often handle the requests of dozens, hundreds, or thousands of front-end clients simultaneously. However, by default, mutilate establishes one connection per server and meters requests one at a time (it waits for a reply before sending the next request). This artificially limits throughput (i.e. queries per second), as the round-trip network latency is almost certainly far longer than the time it takes for the memcached server to process one request.

In order to get reasonable benchmark results with mutilate, it needs to be configured to more accurately portray a realistic client workload. In general, this means ensuring that (1) there are a large number of client connections, (2) there is the potential for a large number of outstanding requests, and (3) the memcached server saturates and experiences queuing delay far before mutilate does. I suggest the following guidelines:

    Establish on the order of 100 connections per memcached server thread.
    Don't exceed more than about 16 connections per mutilate thread.
    Use multiple mutilate agents in order to achieve (1) and (2).
    Do not use more mutilate threads than hardware cores/threads.
    Use -Q to configure the "master" agent to take latency samples at slow, a constant rate.

https://github.com/leverich/mutilate

*/
func NewMutilate(
	executor executor.Executor,
	memached_uri string,
	mutilate_srv_connections int,
	mutilate_srv_threads int,
	mutilate_agent_threads int) Mutilate {
	return Mutilate{
	//exec 		     : executor,
	//mutilate_threads     : mutilate_threads,
	//mutilate_connections : mutilate_connections,
	//memached_uri 	     : memached_uri,
	}
}

func (m *Mutilate) Populate() error {
	pop_cmd := fmt.Sprintf("mutilate -s %s --loadonly", m.memcached_uri)
	populateTask, err := m.executor.Execute(pop_cmd)
	if err != nil {
		return err
	}
	populateTask.Wait(0)

	_, status := populateTask.Status()

	_ = status

	//status->code()

	return nil
}

func (m Mutilate) Tune(slo int, timeoutMs int) (targetQPS int, err error) {

	return -1, nil
}

func (m Mutilate) Load(qps int, durationMs int) (sli int, err error) {
	return -1, nil
}
