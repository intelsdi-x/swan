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

package sysctl

import (
	"io/ioutil"
	"path"
	"strings"

	"github.com/pkg/errors"
)

// Get returns the value of the sysctl key specified by name.
func Get(name string) (string, error) {
	// "net.ipv4.tcp_syncookies" translates into "/proc/sys/net/ipv4/tcp_syncookies"
	const sysctlRoot = "/proc/sys"
	relativeSysctlPath := strings.Replace(name, ".", "/", -1)
	sysctlPath := path.Join(sysctlRoot, relativeSysctlPath)

	byteContent, err := ioutil.ReadFile(sysctlPath)
	if err != nil {
		return "", errors.Wrapf(err, "could not read file %q", sysctlPath)
	}

	// As the sys file system represent single values as files, they are
	// terminated with a newline. We trim trailing newline, if present.
	content := strings.TrimSuffix(string(byteContent), "\n")

	return content, nil
}
