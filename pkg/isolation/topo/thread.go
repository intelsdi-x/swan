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

package topo

import (
	"github.com/Sirupsen/logrus"
	"github.com/pkg/errors"
)

// Thread represents a hyperthread, typically presented by the operating
// system as a logical CPU.
type Thread interface {
	ID() int
	Core() int
	Socket() int
	Equals(Thread) bool
}

// NewThread returns a new thread with the supplied thread, core, and
// socket IDs.
func NewThread(id int, core int, socket int) Thread {
	return thread{id, core, socket}
}

// NewThreadFromID returns new Thread from ThreadID
func NewThreadFromID(id int) (thread Thread, err error) {
	allThreads, err := Discover()
	if err != nil {
		logrus.Errorf("NewThreadFromID: Could not discover CPUs")
		return thread, errors.Wrapf(err, "could not discover CPUs")
	}

	foundThreads := allThreads.Filter(
		func(t Thread) bool {
			return t.ID() == id
		})

	if len(foundThreads) == 0 {
		logrus.Errorf("NewThreadFromID: Could not find thread with id %d on platform\n", id)
		return thread, errors.Errorf("could not find thread with id %d on platform", id)
	}

	return foundThreads[0], err
}

type thread struct {
	id     int
	core   int
	socket int
}

func (t thread) ID() int {
	return t.id
}

func (t thread) Core() int {
	return t.core
}

func (t thread) Socket() int {
	return t.socket
}

func (t thread) Equals(that Thread) bool {
	return t.ID() == that.ID() &&
		t.Core() == that.Core() &&
		t.Socket() == that.Socket()
}
