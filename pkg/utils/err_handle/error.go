package err_handle

import "github.com/Sirupsen/logrus"

// Check the supplied error, log and exit if non-nil.
func Check(err error) {
	if err != nil {
		logrus.Debugf("%+v", err)
		logrus.Fatalf("%v", err)
	}
}
