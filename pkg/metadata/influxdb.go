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

package metadata

import (
	"fmt"
	"strings"
	"time"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/intelsdi-x/swan/pkg/conf"
	"github.com/pkg/errors"
)

const (
	influxMetadata = "metadata"
)

// InfluxDBConfig holds configuration for InfluxDB
type InfluxDBConfig struct {
	httpConfig client.HTTPConfig
	dbName     string
}

// InfluxDB is a helper struct which keeps the InfluxDB session alive,
// holds the active configuration and the experiment id to tag the metadata with.
type InfluxDB struct {
	experimentID string
	session      client.Client
	config       InfluxDBConfig
}

// DefaultInfluxDBConfig applies the InfluxDB settings from the command line flags and
// environment variables.
func DefaultInfluxDBConfig() InfluxDBConfig {
	return InfluxDBConfig{
		dbName: conf.InfluxDBName.Value(),
		httpConfig: client.HTTPConfig{
			Addr:               fmt.Sprintf("http://%s:%d", conf.InfluxDBAddress.Value(), conf.InfluxDBPort.Value()),
			Password:           conf.InfluxDBPassword.Value(),
			Username:           conf.InfluxDBUsername.Value(),
			InsecureSkipVerify: conf.InfluxDBInsecureSkipVerify.Value(),
		},
	}
}

// NewInfluxDB returns the Metadata helper from an experiment id and configuration.
func NewInfluxDB(experimentID string, config InfluxDBConfig) (Metadata, error) {
	var err error

	metadata := &InfluxDB{
		experimentID: experimentID,
		config:       config,
	}

	metadata.session, err = client.NewHTTPClient(metadata.config.httpConfig)

	if err != nil {
		return nil, errors.Wrapf(err, "cannot create influx client for experiment %s", experimentID)
	}

	if conf.InfluxDBCreateDatabase.Value() {
		response, err := metadata.session.Query(client.Query{
			Command:  fmt.Sprintf("CREATE DATABASE %s", config.dbName),
			Database: ""})
		if err != nil {
			return nil, errors.Wrapf(err, "cannot create influx database for experiment %s", experimentID)
		}
		if response.Error() != nil {
			return nil, errors.Wrapf(response.Error(), "response contains error for experiment %s", experimentID)
		}

	}

	return metadata, nil
}

// influxDBStoreMap writes metadata to the database with tags attached to it.
// It writes values (metadata) one by one/row by row. No aggregation is being done.
func influxDBStoreMap(m *InfluxDB, metadata map[string]string, kind string) error {

	//err := m.session.Query(`INSERT INTO metadata (experiment_id, kind, time, timeuuid, metadata) VALUES (?, ?, ?, ?, ?)`, m.experimentID, kind, time.Now(), gocql.TimeUUID(), metadata).Exec()

	batchPoints, err := client.NewBatchPoints(client.BatchPointsConfig{Database: m.config.dbName})
	if err != nil {
		return errors.Wrapf(err, "creation of batch points for InfluxDB failed for metadata kind %q", kind)
	}

	tags := map[string]string{"kind": kind, "experiment_id": m.experimentID}

	now := time.Now()
	fields := make(map[string]interface{})
	// Copy metadata into proper structure
	for key := range metadata {
		fields[key] = metadata[key]
	}
	point, err := client.NewPoint(influxMetadata, tags, fields, now)
	if err != nil {
		return errors.Wrapf(err, "cannot create new point, kind %q", kind)
	}

	batchPoints.AddPoint(point)

	err = m.session.Write(batchPoints)
	if err != nil {
		return errors.Wrapf(err, "cannot publish metadata of kind %q", kind)
	}
	return nil
}

// Record stores a key and value and associates with the experiment id.
func (m *InfluxDB) Record(key, value, kind string) error {
	metadata := map[string]string{}
	metadata[key] = value
	return influxDBStoreMap(m, metadata, kind)
}

// RecordMap stores a key and value map and associates with the experiment id.
func (m *InfluxDB) RecordMap(metadata map[string]string, kind string) error {
	return influxDBStoreMap(m, metadata, kind)
}

// GetByKind retrive single kind from the database. If duplicates are found then
// the last one is returned.
// Returns error if no kind or too many groups found.
func (m *InfluxDB) GetByKind(kind string) (map[string]string, error) {
	var metadata = make(map[string]string)
	// There are two tags currently and query gets rid of them by groupping.
	cmd := fmt.Sprintf("SELECT last(*) FROM %s WHERE experiment_id='%s' AND kind='%s' GROUP BY experiment_id,kind", influxMetadata, m.experimentID, kind)

	query := client.Query{
		Command:  cmd,
		Database: m.config.dbName,
	}

	response, err := m.session.Query(query)

	if err != nil {
		return nil, errors.Wrapf(err, "Failed to contsturct query for influxdb for experiment %s", m.experimentID)
	}

	if response.Error() != nil {
		return nil, errors.Wrapf(response.Error(), "Response from influxdb contained error for experiment %s", m.experimentID)
	}

	for _, result := range response.Results {
		for _, row := range result.Series {
			for _, value := range row.Values {
				for idx, cell := range value {
					// InfluxDB at index 0 returns timestamp and timestamp is not needed in the metadata. Skip it.
					// Also the results may be sparse thus skip empty cells.
					if cell != nil && idx != 0 {
						column := strings.Replace(row.Columns[idx], "last_", "", 1)
						metadata[column] = cell.(string)
					}
				}
			}
		}
	}

	return metadata, nil
}

// Clear deletes all metadata entries associated with the current experiment id.
func (m *InfluxDB) Clear() error {

	cmd := fmt.Sprintf("DROP SERIES FROM %s WHERE experiment_id ='%s'", influxMetadata, m.experimentID)

	query := client.Query{
		Command:  cmd,
		Database: m.config.dbName,
	}

	response, err := m.session.Query(query)

	if err != nil {
		return errors.Wrapf(err, "Failed to contsturct query for influxdb for experiment %s", m.experimentID)
	}

	if response.Error() != nil {
		return errors.Wrapf(response.Error(), "Response from influxdb contained error for experiment %s", m.experimentID)
	}
	return nil
}
