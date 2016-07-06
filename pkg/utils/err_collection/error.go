package errcollection

import (
	"fmt"

	"github.com/pkg/errors"
)

const delimiter = "; "

// ErrorCollection gives ability to return multiple errors instead of one.
// It gathers errors and return error with message combined with messages
// from all given errors delimited by defined string.
type ErrorCollection struct {
	errorList []error
}

// Add inserts new error to collection.
func (e *ErrorCollection) Add(err error) {
	e.errorList = append(e.errorList, err)
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
