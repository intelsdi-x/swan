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
	"github.com/intelsdi-x/swan/pkg/executor"
)

// SessionLauncher starts Snap Collection session and returns handle to that session.
type SessionLauncher interface {
	LaunchSession(monitoredTask executor.TaskInfo, tags map[string]interface{}) (executor.TaskHandle, error)
}
