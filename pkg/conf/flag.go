package conf

import (
	"fmt"
	"os"
	"strings"
)

// Flag represents option's definition from CLI and Environment variable.
type Flag struct {
	name         string
	description  string
	defaultValue string
}

// NewFlag is a constructor of Flag struct.
func NewFlag(flagName string, description string, defaultValue string) Flag {
	return Flag{
		name:         flagName,
		description:  description,
		defaultValue: defaultValue,
	}
}

// EnvName returns name converted to swan environment variable name.
// In order to create environment variable name from flag we need to make it uppercase
// and add SWAN prefix. For instance: "cassandra_host" will be "SWAN_CASSANDRA_HOST".
func (f Flag) EnvName() string {
	return fmt.Sprintf("%s_%s", "SWAN", strings.ToUpper(f.name))
}

// Clear unset the corresponded environment variable.
func (f Flag) Clear() {
	os.Unsetenv(f.EnvName())
}
