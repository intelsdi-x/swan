// Copyright (c) 2018 Intel Corporation
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

package krico

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/snap"
	"time"
	//"github.com/intelsdi-x/snap/scheduler/wmap"
	//"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/snap/publishers"
	"github.com/intelsdi-x/snap/scheduler/wmap"
	"github.com/intelsdi-x/swan/pkg/conf"
)

func DefaultConfig(cgroup string, domain string) snap.SessionConfig {

	pub := publishers.Publisher{
		PluginName: snap.CassandraPublisher,
		Publisher: wmap.NewPublishNode("cassandra", snap.PluginAnyVersion),
	}
	pub.Publisher.AddConfigItem("server", conf.CassandraAddress.Value())
	pub.Publisher.AddConfigItem("keyspaceName", conf.CassandraKeyspaceName.Value())

	return snap.SessionConfig{
		SnapteldAddress: snap.SnapteldAddress.Value(),
		Interval:        1 * time.Second,
		Publisher:       pub.Publisher,
		Plugins: []string{
			pub.PluginName,
			"snap-plugin-collector-perfevents",
			"snap-plugin-collector-libvirt",
		},
		TaskName: "swan-krico-session",
		Metrics: []string{
			"/intel/libvirt/"+domain+"/cpu/cputime", // cpu:time ( 1.0 == 1 logic CPU )
			"/intel/libvirt/"+domain+"/memory/rss", // ram:used [GB]
			"/intel/linux/perfevents/cgroup/cache-references/"+cgroup, // cpu:cache:references ( L3 memory [1/s] )
			"/intel/linux/perfevents/cgroup/cache-misses/"+cgroup, // cpu:cache:misses ( L3 memory [1/s] )
			"/intel/libvirt/"+domain+"/disk/*/wrbytes", // disk:bandwidth:read [MiB/s]
			"/intel/libvirt/"+domain+"/disk/*/rdbytes", // disk:bandwidth:write [MiB/s]
			"/intel/libvirt/"+domain+"/disk/*/wrreq", // disk:operations:read [1/s]
			"/intel/libvirt/"+domain+"/disk/*/rdreq", // disk:operations:write [1/s]
			"/intel/libvirt/"+domain+"/network/*/txbytes", // network:bandwidth:send [MiB/s]
			"/intel/libvirt/"+domain+"/network/*/rxbytes", // network:bandwidth:receive [MiB/s]
			"/intel/libvirt/"+domain+"/network/*/txpackets", // network:packets:send [1/s]
			"/intel/libvirt/"+domain+"/network/*/rxpackets", // network:packets:receive [1/s]
		},
	}
}

type Session struct {
	session *snap.Session
}

func NewSessionLauncher(config snap.SessionConfig) (*Session, error) {

	session, err := snap.NewSessionLauncher(config)

	if err != nil {
		return nil, err
	}

	return &Session{
		session: session,
	}, nil
}

func (s *Session) Launch() (executor.TaskHandle, error) {
	return s.session.Launch()
}

func (s *Session) String() string  {
	return "Snap KRICO Collection"
}