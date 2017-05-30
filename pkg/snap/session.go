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

package snap

import (
	"fmt"
	"os"
	"time"

	"github.com/intelsdi-x/snap/core"
	"github.com/intelsdi-x/snap/mgmt/rest/client"
	"github.com/intelsdi-x/snap/mgmt/rest/v1/rbody"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/pkg/errors"
)

type taskInfo struct {
	Version  int
	Schedule *client.Schedule
	Workflow *wmap.WorkflowMap
	Name     string
	Deadline string
	ID       string
	State    string
}

// CollectNodeConfigItem represents ConfigItem in CollectWorkflowMapNode.
type CollectNodeConfigItem struct {
	Ns    string
	Key   string
	Value interface{}
}

// Session provides construct for tagging metrics for a specified time span
// defined by Start() and Stop().
type Session struct {
	// TaskName is name of task in Snap.
	TaskName string

	// Metrics to tag in session.
	Metrics []string

	// CollectNodeConfigItems represent ConfigItems for CollectNode.
	CollectNodeConfigItems []CollectNodeConfigItem

	// Publisher for tagged metrics.
	Publisher *wmap.PublishWorkflowMapNode

	// Client to Snapteld.
	pClient *client.Client

	schedule *client.Schedule
}

// NewSession generates a session with a name and a list of metrics to tag.
// The interval cannot be less than second granularity.
func NewSession(
	taskName string,
	metrics []string,
	interval time.Duration,
	pClient *client.Client,
	publisher *wmap.PublishWorkflowMapNode) *Session {

	return &Session{
		TaskName: taskName,
		schedule: &client.Schedule{
			Type:     "simple",
			Interval: interval.String(),
		},
		Metrics:                metrics,
		pClient:                pClient,
		Publisher:              publisher,
		CollectNodeConfigItems: []CollectNodeConfigItem{},
	}
}

// Launch starts Snap task.
func (s *Session) Launch(tags map[string]interface{}) (executor.TaskHandle, error) {
	task := taskInfo{
		Name:     s.TaskName,
		Version:  1,
		Schedule: s.schedule,
	}

	wf := wmap.NewWorkflowMap()

	formattedTags := make(map[string]string)
	for key, value := range tags {
		formattedTags[key] = fmt.Sprintf("%v", value)
	}
	wf.CollectNode.Tags = map[string]map[string]string{"": formattedTags}

	for _, metric := range s.Metrics {
		wf.CollectNode.AddMetric(metric, PluginAnyVersion)
	}

	for _, configItem := range s.CollectNodeConfigItems {
		wf.CollectNode.AddConfigItem(configItem.Ns, configItem.Key, configItem.Value)
	}

	loaderConfig := DefaultPluginLoaderConfig()
	loaderConfig.SnapteldAddress = s.pClient.URL

	// Add specified publisher to workflow as well.
	wf.CollectNode.Add(s.Publisher)

	task.Workflow = wf

	r := s.pClient.CreateTask(
		task.Schedule,
		task.Workflow,
		task.Name,
		task.Deadline,
		true,
		10,
	)
	if r.Err != nil {
		return nil, errors.Wrapf(r.Err, "could not create snap task %q", task.Name)
	}

	task.ID = r.ID
	task.State = r.State

	return &Handle{
		task:    task,
		pClient: s.pClient,
	}, nil
}

// Handle is handle for Snap task.
type Handle struct {
	// Active taskInfo.
	task taskInfo

	// Client to Snapteld.
	pClient *client.Client

	// lastFailureMessage must only be set by getSnapTask().
	lastFailureMessage string
}

// String returns name of snap task.
func (s *Handle) String() string {
	return fmt.Sprintf("Snap Task %q running on node %q",
		s.task.Name, s.pClient.URL)
}

// Address returns snapteld address.
func (s *Handle) Address() string {
	return s.pClient.URL
}

// ExitCode returns -1 when snap task is disabled because of errors or not terminated.
// Returns '0' when task is ended.
func (s *Handle) ExitCode() (int, error) {
	task, err := s.getSnapTask()
	if err != nil {
		return -1, err
	}

	if s.Status() != executor.TERMINATED {
		return -1, errors.Errorf("snap task %q is not finished yet", s.task.Name)
	}

	if task.State == "Disabled" {
		return -1, nil
	}

	// Stopped or Ended statuses are OK.
	return 0, nil
}

// Status checks if Snap task is running.
func (s *Handle) Status() executor.TaskState {
	taskState, _, _ := s.getStatus()
	return taskState
}

func (s *Handle) getStatus() (executor.TaskState, core.TaskState, error) {
	task, err := s.getSnapTask()
	if err != nil {
		return executor.TERMINATED, core.TaskDisabled, err
	}

	if task.State == core.TaskDisabled.String() {
		s.lastFailureMessage = task.LastFailureMessage
		return executor.TERMINATED, core.TaskDisabled, nil
	}

	if task.State == core.TaskSpinning.String() ||
		task.State == core.TaskFiring.String() ||
		task.State == core.TaskStopping.String() {
		return executor.RUNNING, core.TaskSpinning, nil
	}

	// Task Ended or Stopped.
	return executor.TERMINATED, core.TaskStopped, nil
}

// getSnapTaskStatus connects to snap to obtain current state of the task.
func (s *Handle) getSnapTask() (*rbody.ScheduledTaskReturned, error) {
	task := s.pClient.GetTask(s.task.ID)
	if task.Err != nil {
		return nil, errors.Wrapf(task.Err, "could not get information about snap task %q with id %q: "+
			"it could be removed or never existed",
			s.task.Name, s.task.ID)
	}

	return task.ScheduledTaskReturned, nil
}

// Stop blocks and stops Snap task.
// When task is already stopped or ended, then it will immediately return.
func (s *Handle) Stop() error {
	executorStatus, coreStatus, err := s.getStatus()
	if err != nil {
		return err
	}

	if coreStatus == core.TaskDisabled {
		return errors.Errorf("snap task %q has been disabled because of errors: %q", s.task.Name, s.lastFailureMessage)
	}

	if executorStatus == executor.TERMINATED {
		return nil
	}

	rs := s.pClient.StopTask(s.task.ID)
	if rs.Err != nil {
		return errors.Wrapf(rs.Err, "could not stop snap task %q: %v", s.task.ID)
	}

	err = s.waitForStop()
	if err != nil {
		return errors.Wrapf(err, "could not stop snap task %q", s.task.ID)
	}

	return nil
}

// Wait blocks until the Snap task is executed at least once
// (including hits that happened in the past).
func (s *Handle) Wait(timeout time.Duration) (bool, error) {
	executorStatus, coreStatus, err := s.getStatus()
	if err != nil {
		return true, err
	}

	if coreStatus == core.TaskDisabled {
		return true, errors.Errorf("snap task %q has been disabled because of errors: %q", s.task.Name, s.lastFailureMessage)
	}

	if executorStatus == executor.TERMINATED {
		return true, nil
	}

	// Task has not finished yet.

	stopper := make(chan struct{})
	taskCompletionInfo := s.waitForTaskCompletion(stopper)

	var timeoutChannel <-chan time.Time
	if timeout != 0 {
		// In case of wait with timeout set the timeout channel.
		timeoutChannel = time.After(timeout)
	}

	select {
	case err := <-taskCompletionInfo:
		// If waitEndChannel is closed then task is terminated.
		return true, err
	case <-timeoutChannel:
		// If timeout time exceeded return then task did not terminate yet.
		stopper <- struct{}{}
		return false, nil
	}
}

// StdoutFile returns error for snap handles.
func (s *Handle) StdoutFile() (*os.File, error) {
	return nil, errors.New("snap tasks don't support stdout file")
}

// StderrFile returns error for snap handles.
func (s *Handle) StderrFile() (*os.File, error) {
	return nil, errors.New("snap tasks don't support stderr file")
}

// EraseOutput does nothing for snap tasks.
func (s *Handle) EraseOutput() error {
	return nil
}

func (s *Handle) waitForTaskCompletion(stopWaiting <-chan struct{}) (result <-chan error) {
	errorChan := make(chan error)
	go func() {
		for {
			task, err := s.getSnapTask()
			if err != nil {
				err := errors.Wrapf(err, "cannot get snap task %q information", s.task.ID)
				errorChan <- err
				return
			}

			if task.State == core.TaskStopped.String() {
				errorChan <- nil
				return
			}

			if task.State == core.TaskDisabled.String() {
				err := errors.Errorf("snap task %q has been disabled because of errors: %q",
					s.task.Name, task.LastFailureMessage)
				errorChan <- err
				return

			}

			// TODO(skonefal): Refactor this when Snap 1.3 with 'count' support is released.
			if (task.HitCount - (task.FailedCount + task.MissCount)) > 0 {
				stopErr := s.Stop()
				errorChan <- stopErr
				return
			}

			timeoutChannel := time.After(500 * time.Millisecond)
			select {
			case <-stopWaiting:
				return
			case <-timeoutChannel:
				break
			}

		}

	}()

	return errorChan
}

func (s *Handle) waitForStop() error {
	for {
		t := s.pClient.GetTask(s.task.ID)
		if t.Err != nil {
			return errors.Wrapf(t.Err, "could not get task %q", s.task.ID)
		}

		if t.State == "Stopped" {
			return nil
		}

		if t.State == "Disabled" {
			return errors.Errorf("snap task %q has been disabled because of errors: %q", t.Name, t.LastFailureMessage)
		}

		time.Sleep(100 * time.Millisecond)
	}
}
