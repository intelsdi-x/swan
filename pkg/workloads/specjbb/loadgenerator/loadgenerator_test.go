package loadgenerator

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/athena/pkg/executor/mocks"
	. "github.com/smartystreets/goconvey/convey"
	"github.com/stretchr/testify/mock"
)

func TestSPECjbbLoadGenerator(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	load := 60
	duration := 10 * time.Millisecond

	Convey("When creating load generator", t, func() {
		controller := new(mocks.Executor)
		transactionInjector := new(mocks.Executor)
		config := NewDefaultConfig()
		config.PathToBinary = "test"

		loadGenerator := NewLoadGenerator(controller, []executor.Executor{
			transactionInjector,
		}, config)

		Convey("And generating load", func() {
			controller.On("Execute", mock.AnythingOfType("string")).Return(new(mocks.TaskHandle), nil)

			transactionInjector.On("Execute", mock.AnythingOfType("string")).Return(new(mocks.TaskHandle), nil)

			loadGeneratorTaskHandle, err := loadGenerator.Load(load, duration)

			Convey("On success, error should be nil", func() {
				So(err, ShouldBeNil)
			})
			Convey("On success, task handle should not be nil", func() {
				So(loadGeneratorTaskHandle, ShouldNotBeNil)
			})
		})
	})

}
