package cli

import "github.com/codegangsta/cli"

const (
	cassandraAddressKey = "SWAN_CASSANDRA_ADDRESS"
	snapDAddress = "SWAN_SNAPD_ADDRESS"

	cassandraAddressDefault = "127.0.0.1"
	snapDAddressDefault = "10.4.1.1"
)

type SomeLauncher struct {
	c Config
	e executor.Executor
}

type Config struct {
	CassandraAddress string
	SnapDAddress 	 string
	Threads      int
}

func DefaultConfig() Config {
	return Config{
		CassandraAddress: osutil.GetEnvOrDefault(cassandraAddressKey, cassandraAddressDefault),
		SnapDAddress: osutil.GetEnvOrDefault(snapDAddress, snapDAddressDefault),
		Threads: 4,
	}
}

func (s SomeLauncher) Launch() (executor.TaskHandle, error) {
	task, err := s.e.Execute("command")
	if err != nil {
		return nil, err
	}

	return task, nil
}

// Helper function informs user of enviroment variables that are used by this Launcher
// It should be part of Launcher Interface
func (s SomeLauncher) Helper() []cli.Helper {
	// Helper returs list of used enviroment variables in format
	// <KEY, DEFAULT_VALUE, Description for experiment user, Optional/required>
	// Optional/required - if required key is non-existent, CLI will throw error
	return
	{
		cli.Helper{cassandraAddressKey, cassandraAddressDefault, "Cassandra hostname", cli.OPTIONAL},
		cli.Helper{snapDAddress, snapDAddressDefault, "Cassandra hostname", cli.REQUIRED},
	}
}

// Name returns human readable name for job.
func (m SomeLauncher) Name() string {
	return "SomeLauncher"
}
