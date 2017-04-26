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

package random

import (
	"math/rand"
	"time"
)

var source int64

// PortsFromRange returns 'count' random ports between 'start' and 'end'.
func PortsFromRange(start int, end int, count int) []int {
	if source == 0 {
		source = time.Now().UnixNano()
	} else {
		source = source + 1
	}
	r := rand.New(rand.NewSource(source))
	ports := map[int]struct{}{}
	for len(ports) < count {
		port := r.Intn(end-start) + start
		ports[port] = struct{}{}
	}

	out := []int{}
	for port := range ports {
		out = append(out, port)
	}

	return out
}

// Ports return 'count' random ports in range between 22768 to 32768.
func Ports(count int) []int {
	const lowEnd = 22768
	const highEnd = 32768
	return PortsFromRange(lowEnd, highEnd, count)
}
