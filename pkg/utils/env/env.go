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

package env

import "os"

// GetOrDefault returns the environment variable or default string.
func GetOrDefault(env string, defaultStr string) string {
	if env == "" {
		return defaultStr
	}

	fetchedEnv := os.Getenv(env)

	if fetchedEnv == "" {
		return defaultStr
	}

	return fetchedEnv
}
