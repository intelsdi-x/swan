package specjbb

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

const (
	testLoad     = 60
	testDuration = 10 * time.Millisecond
	testSlo      = 0 // Tune in specjbb does not accept any slo, we can set it to 0.
)

func TestSPECjbbLoadGenerator(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("When creating load generator", t, func() {
		controller := new(mocks.Executor)
		transactionInjector := new(mocks.Executor)
		config := DefaultLoadGeneratorConfig()
		config.PathToBinary = "test"

		loadGenerator := NewLoadGenerator(controller, []executor.Executor{
			transactionInjector,
		}, config)

		Convey("And generating load", func() {
			controller.On("Execute", mock.AnythingOfType("string")).Return(new(mocks.TaskHandle), nil)

			transactionInjector.On("Execute", mock.AnythingOfType("string")).Return(new(mocks.TaskHandle), nil)

			loadGeneratorTaskHandle, err := loadGenerator.Load(testLoad, testDuration)

			Convey("On success, error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("On success, task handle should not be nil", func() {
				So(loadGeneratorTaskHandle, ShouldNotBeNil)
			})
		})
	})

}
