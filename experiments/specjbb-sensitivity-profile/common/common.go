package common

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
)

// PrepareSpecjbbLoadGenerator creates new LoadGenerator based on specjbb.
func PrepareSpecjbbLoadGenerator(controllerAddress string, transactionInjectorsCount int) (executor.LoadGenerator, error) {
	var loadGeneratorExecutor executor.Executor
	var transactionInjectors []executor.Executor
	if controllerAddress != "127.0.0.1" {
		var err error
		loadGeneratorExecutor, err = executor.NewRemoteFromIP(controllerAddress)
		if err != nil {
			return nil, err
		}
		for i := 0; i < transactionInjectorsCount; i++ {
			transactionInjector, err := executor.NewRemoteFromIP(controllerAddress)
			if err != nil {
				return nil, err
			}
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	} else {
		loadGeneratorExecutor = executor.NewLocal()
		for i := 0; i < transactionInjectorsCount; i++ {
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	}

	specjbbLoadGeneratorConfig := specjbb.DefaultLoadGeneratorConfig()
	specjbbLoadGeneratorConfig.ControllerAddress = controllerAddress

	loadGeneratorLauncher := specjbb.NewLoadGenerator(loadGeneratorExecutor,
		transactionInjectors, specjbbLoadGeneratorConfig)

	return loadGeneratorLauncher, nil
}
