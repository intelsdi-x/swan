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

// CheckWithContext checks the error and exit if it is not nil. Logs additional context information.
func CheckWithContext(err error, context string) {
	if err != nil {
		logrus.Debugf("%s: %+v", context, err)
		logrus.Fatalf("%s: %v", context, err)
	}
}
