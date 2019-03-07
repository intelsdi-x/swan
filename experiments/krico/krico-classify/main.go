// Copyright (c) 2018 Intel Corporation
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
	"strings"
	"time"
)

var (
	appName         = os.Args[0]
	kricoApiAddress = conf.NewStringFlag("krico_api_address", "Ip address of KRICO API service.", "localhost:5000")
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

	// Initialize workloads.
	workload.Initialize(experimentID)

	// Run workloads.
	instances := workload.RunWorkloadsClassification()

	// Prepare metadata.
	records := map[string]string{
		experiment.ExperimentKey: experimentID,
		"commands_arguments":     strings.Join(os.Args, ","),
		"experiment_name":        appName,
	}

	// Save metadata.
	err = metaData.RecordMap(records, metadata.TypeEmpty)
	errutil.CheckWithContext(err, "Cannot save metadata in Cassandra Metadata Database")

	// Connect to KRICO via gRPC to update metadata.
	conn, err := grpc.Dial(kricoApiAddress.Value(), grpc.WithInsecure())
	errutil.CheckWithContext(err, "Cannot connect to KRICO!")

	krico := api.NewApiClient(conn)

	_, err = krico.ImportMonitorSamplesFromSwanExperiment(context.Background(), &api.ImportMonitorSamplesFromSwanExperimentRequest{ExperimentId: experimentID})
	errutil.CheckWithContext(err, "Cannot send request to KRICO for importing monitor samples!")
	log.Infof("Monitor samples imported!")

	for _, instance := range instances {
		predictedCategory, err := krico.Classify(context.Background(), &api.ClassifyRequest{InstanceId: instance})
		errutil.CheckWithContext(err, fmt.Sprintf("Cannot send request to KRICO for classify %v instance.", instance))
		log.Infof("Predicted category for %v instance: %q", instance, predictedCategory.ClassifiedAs)
	}

	defer conn.Close()
}
