package common

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
)

// PrepareSpecjbbLoadGenerator creates new LoadGenerator based on specjbb.
func PrepareSpecjbbLoadGenerator(ip string) (executor.LoadGenerator, error) {
	var loadGeneratorExecutor executor.Executor
	var transactionInjectors []executor.Executor
	txICount := specjbb.TxICountFlag.Value()
	if ip != "127.0.0.1" {
		var err error
		loadGeneratorExecutor, err = executor.NewRemoteFromIP(ip)
		if err != nil {
			return nil, err
		}
		for i := 1; i <= txICount; i++ {
			transactionInjector, err := executor.NewRemoteFromIP(ip)
			if err != nil {
				return nil, err
			}
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	} else {
		loadGeneratorExecutor = executor.NewLocal()
		for i := 1; i <= txICount; i++ {
			transactionInjector := executor.NewLocal()
			transactionInjectors = append(transactionInjectors, transactionInjector)
		}
	}

	specjbbLoadGeneratorConfig := specjbb.NewDefaultConfig()
	specjbbLoadGeneratorConfig.ControllerIP = ip
	specjbbLoadGeneratorConfig.TxICount = txICount

	loadGeneratorLauncher := specjbb.NewLoadGenerator(loadGeneratorExecutor,
		transactionInjectors, specjbbLoadGeneratorConfig)

	return loadGeneratorLauncher, nil
}
