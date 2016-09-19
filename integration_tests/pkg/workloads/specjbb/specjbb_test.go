package specjbb

import (
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/backend"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb/loadgenerator"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	txICount     = 1
	load         = 6000
	loadDuration = 1000
)

// TestSPECjbb is an integration test with SPECjbb components.
func TestSPECjbb(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using default config", t, func() {
		specjbbLoadGeneratorConfig := loadgenerator.NewDefaultConfig()
		specjbbLoadGeneratorConfig.TxICount = txICount

		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			for i := 1; i <= txICount; i++ {
				transactionInjector := executor.NewLocal()
				transactionInjectors = append(transactionInjectors, transactionInjector)
			}
			loadGeneratorLauncher := loadgenerator.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration)

			Convey("Proper handle should be returned", func() {
				So(err, ShouldBeNil)
				So(loadGeneratorTaskHandle, ShouldNotBeNil)

				defer loadGeneratorTaskHandle.EraseOutput()
				defer loadGeneratorTaskHandle.Clean()

				Convey("And after adding SPECjbb backend", func() {
					backendConfig := backend.DefaultSPECjbbBackendConfig()
					backendLauncher := backend.NewBackend(executor.NewLocal(), backendConfig)
					backendTaskHandle, err := backendLauncher.Launch()

					defer backendTaskHandle.EraseOutput()
					defer backendTaskHandle.Clean()

					Convey("Proper handle should be returned", func() {
						So(err, ShouldBeNil)
						So(backendTaskHandle, ShouldNotBeNil)

						Convey("And should work for at least as long as given load duration", func() {
							loadIsTerminated := loadGeneratorTaskHandle.Wait(loadDuration * time.Millisecond)
							backendIsTerminated := backendTaskHandle.Wait(loadDuration * time.Millisecond)
							So(loadIsTerminated, ShouldBeFalse)
							So(backendIsTerminated, ShouldBeFalse)

							Convey("And I should be able to stop with no problem and be terminated", func() {
								err = loadGeneratorTaskHandle.Stop()
								So(err, ShouldBeNil)
								err = backendTaskHandle.Stop()
								So(err, ShouldBeNil)

								state := loadGeneratorTaskHandle.Status()
								So(state, ShouldEqual, executor.TERMINATED)
								state = backendTaskHandle.Status()
								So(state, ShouldEqual, executor.TERMINATED)
							})
						})

					})

				})
			})
		})

	})

}
