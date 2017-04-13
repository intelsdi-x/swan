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
