// Copyright (c) 2019 Intel Corporation
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

package main

import (
	"context"
	"fmt"
	"github.com/intelsdi-x/swan/experiments/krico/api"
	"github.com/intelsdi-x/swan/experiments/krico/workloads"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/intelsdi-x/swan/pkg/experiment"
	"github.com/intelsdi-x/swan/pkg/experiment/logger"
	"github.com/intelsdi-x/swan/pkg/experiment/sensitivity/validate"
	"github.com/intelsdi-x/swan/pkg/metadata"
	"github.com/intelsdi-x/swan/pkg/utils/errutil"
	"github.com/intelsdi-x/swan/pkg/utils/uuid"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"os"
	"strconv"
	"strings"
	"time"
)

var (
	appName          = os.Args[0]
	kricoAPIAddress  = conf.NewStringFlag("krico_api_address", "Ip address of KRICO API service.", "localhost:5000")
	workloadCategory = conf.NewStringFlag("krico_prediction_category", "Workload category", "")
	workloadImage    = conf.NewStringFlag("krico_prediction_image", "Workload image", "default")
	parameterDisk    = conf.NewStringFlag("krico_prediction_disk", "Disk", "0.0")
	parameterMemory  = conf.NewStringFlag("krico_prediction_memory", "Memory ", "0.0")
	parameterRatio   = conf.NewStringFlag("krico_prediction_ratio", "Ratio", "0.0")
	parameterClients = conf.NewStringFlag("krico_prediction_clients", "Clients", "0.0")
)

func main() {

	// Preparing application - setting name, help, parsing flags etc.
	experimentStart := time.Now()
	experiment.Configure()

	// Generate an experiment ID and start the metadata session.
	experimentID := uuid.New()

	// Initialize logger.
	logger.Initialize(appName, experimentID)

	// Connect to metadata database.
	metaData, err := metadata.NewDefault(experimentID)
	errutil.CheckWithContext(err, "Cannot connect to Cassandra Metadata Database")

	// Save experiment runtime environment (configuration, environmental variables, etc).
	err = metadata.RecordRuntimeEnv(metaData, experimentStart)
	errutil.CheckWithContext(err, "Cannot save runtime environment in Cassandra Metadata Database")

	// Validate preconditions.
	validate.OS()

	// Prepare metadata.
	records := map[string]string{
		experiment.ExperimentKey: experimentID,
		"commands_arguments":     strings.Join(os.Args, ","),
		"experiment_name":        appName,
	}

	// Save metadata.
	err = metaData.RecordMap(records, metadata.TypeEmpty)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")

	// Connect to KRICO.
	conn, err := grpc.Dial(kricoAPIAddress.Value(), grpc.WithInsecure())
	errutil.CheckWithContext(err, "Cannot connect to KRICO!")

	krico := api.NewApiClient(conn)

	// Gain parameters for specific workload.
	parameters, err := getWorkloadParameters(workloadCategory.Value())
	errutil.CheckWithContext(err, fmt.Sprintf("Cannot gain parameters for %q workload!", workloadCategory.Value()))

	// Do prediction.
	prediction, err := krico.Predict(context.Background(), &api.PredictRequest{
		Category:   workloadCategory.Value(),
		Image:      workloadImage.Value(),
		Parameters: parameters,
	})
	errutil.CheckWithContext(err, fmt.Sprintf("Cannot make prediction for %q workload!", workloadCategory.Value()))

	log.Infof("Prediction for %v workload: \nRequirements: %q \nFlavor: %q \nHost aggregate: %q", workloadCategory.Value(), prediction.Requirements, prediction.Flavors, prediction.HostAggregates)

	err = conn.Close()
	errutil.CheckWithContext(err, "Cannot close connection to KRICO!")
}

func getWorkloadParameters(category string) (map[string]float64, error) {

	parameters := map[string]float64{}

	param, err := strconv.ParseFloat(parameterDisk.Value(), 64)
	if err != nil {
		return nil, err
	}

	parameters["disk"] = param

	switch category {
	case workload.TypeCaching:

		param, err := strconv.ParseFloat(parameterMemory.Value(), 64)
		if err != nil {
			return nil, err
		}

		parameters["memory"] = param

		param, err = strconv.ParseFloat(parameterRatio.Value(), 64)
		if err != nil {
			return nil, err
		}

		parameters["ratio"] = param

		param, err = strconv.ParseFloat(parameterClients.Value(), 64)
		if err != nil {
			return nil, err
		}

		parameters["clients"] = param

		break
	}

	return parameters, nil
}
