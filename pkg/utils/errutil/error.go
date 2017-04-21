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

package errutil

import (
	"github.com/Sirupsen/logrus"
)

// Check the supplied error, log and exit if non-nil.
func Check(err error) {
	if err != nil {
		logrus.Debugf("%+v", err)
		logrus.Fatalf("%v", err)
	}
}

// CheckWithContext checks error provided and if it is not nil then logs some information and terminates the program.
func CheckWithContext(err error, context string) {
	if err != nil {
		logrus.Debugf("%s: %+v", context, err)
		logrus.Fatalf("%s: %v", context, err)
	}
}

// PanicWithContext checks error provided and if it is not nil then logs some information and emits panic.
func PanicWithContext(err error, context string) {
	if err != nil {
		logrus.Debugf("%s: %+v", context, err)
		logrus.Panicf("%s: %q", context, err)
	}
}
