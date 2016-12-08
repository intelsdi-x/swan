package specjbb

import (
	"bufio"
	"os"
	"strings"
	"testing"
	"time"

	"fmt"
	log "github.com/Sirupsen/logrus"
	"github.com/intelsdi-x/athena/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
	. "github.com/smartystreets/goconvey/convey"
)

const (
	txICount     = 1
	load         = 6000
	loadDuration = 40
)

// TestSPECjbb is an integration test with SPECjbb components.
func TestSPECjbb(t *testing.T) {
	log.SetLevel(log.ErrorLevel)

	Convey("While using default config", t, func() {
		specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()
		specjbbLoadGeneratorConfig.TxICount = txICount

		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration*time.Second)

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

						Convey("And should work for at least as long as given load duration", func() {
							loadIsTerminated := loadGeneratorTaskHandle.Wait(loadDuration * time.Second)
							backendIsTerminated := backendTaskHandle.Wait(loadDuration * time.Second)
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
							substringBackend := "Agent GRP1.Backend.JVM2 has attached to Controller"
							substringTxI := "Agent GRP1.TxInjector.JVM1 has attached to Controller"
							var matchLoad, matchBackend, matchTxI bool
							for scanner.Scan() {
								err := scanner.Err()
								So(err, ShouldBeNil)
								line := scanner.Text()
								if result := strings.Contains(line, substringInitialization); result {
									matchLoad = result
								} else if result := strings.Contains(line, substringBackend); result {
									matchBackend = result
								} else if result := strings.Contains(line, substringTxI); result {
									matchTxI = result
								}
							}
							So(matchLoad, ShouldBeTrue)
							So(matchBackend, ShouldBeTrue)
							So(matchTxI, ShouldBeTrue)
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
	Convey("While using default config", t, func() {
		specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()
		specjbbLoadGeneratorConfig.TxICount = txICount

		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration*time.Second)
			output, err := loadGeneratorTaskHandle.StdoutFile()
			So(err, ShouldBeNil)

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

						Convey("And when I stop backend prematurely", func() {
							// Wait for backend to be registered.
							time.Sleep(20 * time.Second)
							err = backendTaskHandle.Stop()
							So(err, ShouldBeNil)
							So(backendTaskHandle.Status(), ShouldEqual, executor.TERMINATED)
							Convey("Controller should indicate that error occurred", func() {
								// Wait for controller to notice an error.
								time.Sleep(15 * time.Second)
								file, err := os.Open(output.Name())
								So(err, ShouldBeNil)
								defer file.Close()
								scanner := bufio.NewScanner(file)
								stoppedBackend := "Stop the run due to the unexpected error:"
								var match bool
								for scanner.Scan() {
									err := scanner.Err()
									So(err, ShouldBeNil)
									line := scanner.Text()
									fmt.Println(line)
									if result := strings.Contains(line, stoppedBackend); result {
										match = result
										break
									}
								}
								So(match, ShouldBeTrue)
								So(loadGeneratorTaskHandle.Status(), ShouldEqual, executor.TERMINATED)

							})
						})

					})

				})
			})
		})

	})
	Convey("While using default config", t, func() {
		specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()
		specjbbLoadGeneratorConfig.TxICount = txICount

		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration*time.Second)

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
					loadIsTerminated := loadGeneratorTaskHandle.Wait(loadDuration * time.Second)
					So(loadIsTerminated, ShouldBeFalse)
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
		specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()

		Convey("And launching SPECjbb load without transaction injectors", func() {
			var transactionInjectors []executor.Executor
			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration*time.Second)

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

						output, err := loadGeneratorTaskHandle.StdoutFile()
						So(err, ShouldBeNil)
						file, err := os.Open(output.Name())
						defer file.Close()
						Convey("But when the transaction injector is not added, controller should not have information about it in its logs", func() {
							loadIsTerminated := loadGeneratorTaskHandle.Wait(loadDuration * time.Second)
							So(loadIsTerminated, ShouldBeFalse)
							scanner := bufio.NewScanner(file)
							substringWithoutTxI := "Agent GRP1.TxInjector.JVM1 has attached to Controller"
							var matchWithoutTxI bool
							for scanner.Scan() {
								err := scanner.Err()
								So(err, ShouldBeNil)
								line := scanner.Text()
								if result := strings.Contains(line, substringWithoutTxI); result {
									matchWithoutTxI = result
									break
								}
							}
							So(matchWithoutTxI, ShouldBeFalse)

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

		})
	})
	Convey("While using config with no existing path to binary", t, func() {
		specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()
		specjbbLoadGeneratorConfig.PathToBinary = "/no/existing/path"
		specjbbLoadGeneratorConfig.TxICount = txICount

		Convey("And launching SPECjbb load", func() {
			var transactionInjectors []executor.Executor
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)

			loadGeneratorLauncher := specjbb.NewLoadGenerator(executor.NewLocal(),
				transactionInjectors, specjbbLoadGeneratorConfig)
			loadGeneratorTaskHandle, err := loadGeneratorLauncher.Load(load, loadDuration*time.Second)

			Convey("Proper handle should be returned", func() {
				So(err, ShouldBeNil)
				So(loadGeneratorTaskHandle, ShouldNotBeNil)

				Reset(func() {
					loadGeneratorTaskHandle.Stop()
					loadGeneratorTaskHandle.Clean()
					loadGeneratorTaskHandle.EraseOutput()
				})

				Convey("But I should receive error and load generator should be terminated", func() {
					loadIsTerminated := loadGeneratorTaskHandle.Wait(0)
					So(loadIsTerminated, ShouldBeTrue)
				})

			})

		})
	})
}
