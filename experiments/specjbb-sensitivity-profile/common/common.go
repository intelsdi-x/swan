// Copyright (c) 2017 Intel Corporation
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package common

import (
	"github.com/intelsdi-x/swan/pkg/executor"
	"github.com/intelsdi-x/swan/pkg/workloads/specjbb"
)

// PrepareSpecjbbLoadGenerator creates new LoadGenerator for SPECjbb workload.
func PrepareSpecjbbLoadGenerator(controllerAddress string, transactionInjectorsCount int) (executor.LoadGenerator, error) {
	transactionInjectors := make([]executor.Executor, 0)
	controller, err := executor.NewShell(controllerAddress)
	if err != nil {
		return nil, err
	}
	for i := 0; i < transactionInjectorsCount; i++ {
		transactionInjector, err := executor.NewShell(controllerAddress)
		if err != nil {
			return nil, err
		}
		transactionInjectors = append(transactionInjectors, transactionInjector)
	}

	loadGeneratorLauncher := specjbb.NewLoadGenerator(
		controller,
		transactionInjectors,
		specjbb.DefaultLoadGeneratorConfig())

	return loadGeneratorLauncher, nil
}
