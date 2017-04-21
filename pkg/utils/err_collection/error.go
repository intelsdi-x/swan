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

package errcollection

import (
	"fmt"

	"github.com/pkg/errors"
)

const delimiter = ";\n "

// ErrorCollection gives ability to return multiple errors instead of one.
// It gathers errors and return error with message combined with messages
// from all given errors delimited by defined string.
type ErrorCollection struct {
	errorList []error
}

// Add inserts new error to collection.
func (e *ErrorCollection) Add(err error) {
	if err != nil {
		e.errorList = append(e.errorList, err)
	}
}

// GetErrIfAny returns error with combined message from all given errors.
// In case of no error it returns nil.
func (e *ErrorCollection) GetErrIfAny() error {
	if len(e.errorList) == 0 {
		// No error passed so nothing to report.
		return nil
	}

	errMsg := ""
	for i, err := range e.errorList {
		errMsg += fmt.Sprintf("%s", err.Error())

		if i != (len(e.errorList) - 1) {
			// Except the last one, place the delimiter between error messages.
			errMsg += delimiter
		}
	}
	return errors.Errorf(errMsg)
}
