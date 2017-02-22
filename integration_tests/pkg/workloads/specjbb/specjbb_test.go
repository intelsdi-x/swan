package specjbb

import (
	"bufio"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	load         = 1000
	loadDuration = 20 * time.Second
)

// TestSPECjbb is an integration test with SPECjbb components.
func TestSPECjbb(t *testing.T) {
	log.SetLevel(log.ErrorLevel)
	specjbbLoadGeneratorConfig := specjbb.DefaultLoadGeneratorConfig()
	specjbbLoadGeneratorConfig.JVMHeapMemoryGBs = 1
	if _, err := exec.LookPath(specjbbLoadGeneratorConfig.PathToBinary); err != nil {
		t.Logf("Skipping test due to an error %s", err)
		t.Skip("SPECjbb binary is not distributed with Swan. It requires license and should be purchased " +
			"separately (see README for details).")
	}

	Convey("While using default config", t, func() {
		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration)

			Convey("Proper handle should be returned", func() {
				So(err, ShouldBeNil)
				So(loadGeneratorTaskHandle, ShouldNotBeNil)

				Reset(func() {
					loadGeneratorTaskHandle.Stop()
					loadGeneratorTaskHandle.Clean()
					loadGeneratorTaskHandle.EraseOutput()
				})

				Convey("And after adding the SPECjbb backend", func() {
					backendConfig := specjbb.DefaultSPECjbbBackendConfig()
					backendConfig.JVMHeapMemoryGBs = 1
					backendLauncher := specjbb.NewBackend(executor.NewLocal(), backendConfig)
					backendTaskHandle, err := backendLauncher.Launch()

					Reset(func() {
						backendTaskHandle.Stop()
					})

					Convey("Proper handle should be returned", func() {
						So(err, ShouldBeNil)
						So(backendTaskHandle, ShouldNotBeNil)

						Convey("And should work for at least as long as given load duration", func() {
							loadIsTerminated := loadGeneratorTaskHandle.Wait(loadDuration)
							So(loadIsTerminated, ShouldBeFalse)
							backendIsTerminated := backendTaskHandle.Wait(loadDuration)
							So(backendIsTerminated, ShouldBeFalse)

							// Now wait for backend and transaction injectors to finish.
							loadGeneratorTaskHandle.Wait(0)
							backendTaskHandle.Wait(0)

							output, err := loadGeneratorTaskHandle.StdoutFile()
							So(err, ShouldBeNil)
							file, err := os.Open(output.Name())
							defer file.Close()
							scanner := bufio.NewScanner(file)

							// When SPECjbb composite mode is successfully started, the output is:
							//1s:  Agent GRP1.Backend.JVM2 has attached to Controller
							//     1s:  Agent GRP1.TxInjector.JVM1 has attached to Controller
							//     1s:
							//     1s: All agents have connected.
							//     1s:
							//     1s: Attached agents info:
							// Group "GRP1"
							// TxInjectors:
							// JVM1, includes { Driver } @ [127.0.0.1:40910, 127.0.0.1:41757, 127.0.0.1:41462]
							// Backends:
							// JVM2, includes { SM(2),SP(2) } @ [127.0.0.1:38571, 127.0.0.1:45981, 127.0.0.1:35478]
							//
							//1s: Initializing... (init) OK
							// We should look for the proper lines to be sure that our configuration works.
							substringInitialization := "Initializing... (init) OK"
							substringBackend := "Agent GRP1.Backend.specjbbbackend1 has attached to Controller"
							substringTxI := "Agent GRP1.TxInjector.JVM1 has attached to Controller"
							var initializationSuccessful, backendAttachedToController, transactionInjectorAttachedToController bool
							for scanner.Scan() {
								line := scanner.Text()
								if result := strings.Contains(line, substringInitialization); result {
									initializationSuccessful = result
								} else if result := strings.Contains(line, substringBackend); result {
									backendAttachedToController = result
								} else if result := strings.Contains(line, substringTxI); result {
									transactionInjectorAttachedToController = result
								}
							}
							err = scanner.Err()
							So(err, ShouldBeNil)
							So(initializationSuccessful, ShouldBeTrue)
							So(backendAttachedToController, ShouldBeTrue)
							So(transactionInjectorAttachedToController, ShouldBeTrue)

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
	Convey("While using default config", t, func() {
		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration)
			Convey("Proper handle should be returned", func() {
				So(err, ShouldBeNil)
				So(loadGeneratorTaskHandle, ShouldNotBeNil)

				Reset(func() {
					loadGeneratorTaskHandle.Stop()
					loadGeneratorTaskHandle.Clean()
					loadGeneratorTaskHandle.EraseOutput()
				})

				Convey("And after adding the SPECjbb backend", func() {
					backendConfig := specjbb.DefaultSPECjbbBackendConfig()
					backendConfig.JVMHeapMemoryGBs = 1
					backendLauncher := specjbb.NewBackend(executor.NewLocal(), backendConfig)
					backendTaskHandle, err := backendLauncher.Launch()

					Reset(func() {
						backendTaskHandle.Stop()
						backendTaskHandle.Clean()
						backendTaskHandle.EraseOutput()
					})

					Convey("Proper handle should be returned", func() {
						So(err, ShouldBeNil)
						So(backendTaskHandle, ShouldNotBeNil)

						Convey("And when I stop backend prematurely, "+
							"both backend and load generator should be terminated", func() {
							// Wait for backend to be registered.
							time.Sleep(20 * time.Second)
							err = backendTaskHandle.Stop()
							So(err, ShouldBeNil)
							So(backendTaskHandle.Status(), ShouldEqual, executor.TERMINATED)
							loadGeneratorTaskHandle.Wait(0)
							So(loadGeneratorTaskHandle.Status(), ShouldEqual, executor.TERMINATED)
						})

					})

				})
			})
		})

	})
	//TODO(skonefal): Consider deleting this case.
	Convey("While using default config", t, func() {
		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration)

			Convey("Proper handle should be returned", func() {
				So(err, ShouldBeNil)
				So(loadGeneratorTaskHandle, ShouldNotBeNil)

				Reset(func() {
					loadGeneratorTaskHandle.Stop()
					loadGeneratorTaskHandle.Clean()
					loadGeneratorTaskHandle.EraseOutput()
				})

				output, err := loadGeneratorTaskHandle.StdoutFile()
				So(err, ShouldBeNil)
				file, err := os.Open(output.Name())
				defer file.Close()
				Convey("But when the SPECjbb backend is not added, controller should not have information about it in its logs", func() {
					loadGeneratorTaskHandle.Wait(loadDuration)
					scanner := bufio.NewScanner(file)
					substringWithoutBackend := "Agent GRP1.Backend.JVM2 has attached to Controller"
					var matchWithoutBackend bool
					for scanner.Scan() {
						err := scanner.Err()
						So(err, ShouldBeNil)
						line := scanner.Text()
						if result := strings.Contains(line, substringWithoutBackend); result {
							matchWithoutBackend = result
							break
						}
					}
					So(matchWithoutBackend, ShouldBeFalse)
					Convey("And I should be able to stop with no problem and be terminated", func() {
						err = loadGeneratorTaskHandle.Stop()
						So(err, ShouldBeNil)

						state := loadGeneratorTaskHandle.Status()
						So(state, ShouldEqual, executor.TERMINATED)
					})
				})

			})

		})
	})
	Convey("While using default config", t, func() {
		specjbbLoadGeneratorConfig := specjbb.DefaultLoadGeneratorConfig()

		Convey("And launching SPECjbb load without transaction injectors", func() {
			var transactionInjectors []executor.Executor
			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration)
			Convey("Should retrun error.", func() {
				So(loadGeneratorTaskHandle, ShouldBeNil)
				So(err, ShouldNotBeNil)
			})
		})
	})
}
