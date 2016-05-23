package visualization

import (
	log "github.com/Sirupsen/logrus"
	. "github.com/smartystreets/goconvey/convey"
	"testing"
)

func TestSensitivityProfile(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While connecting to casandra with proper parameters", t, func() {
		cluster := configureCluster("127.0.0.1", "snap")
		drawTable(cluster)
	})
}
