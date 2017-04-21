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

package netutil

import (
	"net"
	"time"
)

const retries = 30

// IsListeningFunction is a function type for checking if http endpoint is responding.
type IsListeningFunction func(address string, timeout time.Duration) bool

// IsListening tries to establish TCP connection to given address in a form of `ip:port`.
// It returns true when it was able to connect to given endpoint within timeout time.
func IsListening(address string, timeout time.Duration) bool {
	sleepTime := time.Duration(
		timeout.Nanoseconds() / int64(retries))
	connected := false
	for i := 0; i < retries; i++ {
		conn, err := net.Dial("tcp", address)
		if err != nil {
			time.Sleep(sleepTime)
			continue
		}
		defer conn.Close()
		connected = true
	}

	return connected
}
