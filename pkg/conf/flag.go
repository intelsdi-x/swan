package conf

import (
	"fmt"
	"os"
	"strconv"
	"strings"
)

// flag represents option's definition from CLI and Environment variable.
type flag struct {
	name         string
	description  string
	defaultValue string
}

// envName returns name converted to swan environment variable name.
// In order to create environment variable name from flag we need to make it uppercase
// and add SWAN prefix. For instance: "cassandra_host" will be "SWAN_CASSANDRA_HOST".
func (f flag) envName() string {
	return fmt.Sprintf("%s_%s", "SWAN", strings.ToUpper(f.name))
}

// clear unset the corresponded environment variable.
func (f flag) clear() {
	os.Unsetenv(f.envName())
}

// StringFlag represents flag with string value.
type StringFlag struct {
	flag
	value *string
}

// NewRegisteredStringFlag is a constructor of StringFlag struct.
func NewRegisteredStringFlag(flagName string, description string, defaultValue string) StringFlag {
	strFlag := StringFlag{
		flag: flag{
			name:         flagName,
			description:  description,
			defaultValue: defaultValue,
		},
	}

	strFlag.value = app.Flag(flagName, description).
		Default(defaultValue).OverrideDefaultFromEnvar(strFlag.envName()).String()

	return strFlag
}

// Value returns value of defined flag after parse.
func (s StringFlag) Value() string {
	if *s.value == "" {
		return s.defaultValue
	}

	return *s.value
}

// IntFlag represents flag with int value.
type IntFlag struct {
	flag
	value *int
}

// NewRegisteredIntFlag is a constructor of IntFlag struct.
func NewRegisteredIntFlag(flagName string, description string, defaultValue string) IntFlag {
	intFlag := IntFlag{
		flag: flag{
			name:         flagName,
			description:  description,
			defaultValue: defaultValue,
		},
	}

	intFlag.value = app.Flag(flagName, description).
		Default(defaultValue).OverrideDefaultFromEnvar(intFlag.envName()).Int()

	return intFlag
}

// Value returns value of defined flag after parse.
func (i IntFlag) Value() int {
	if *i.value == 0 {
		ret, err := strconv.Atoi(i.defaultValue)
		if err != nil {
			return 0
		}

		return ret
	}

	return *i.value
}
