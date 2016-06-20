package conf

import (
	"fmt"
	"os"
	"strings"
)

// flag represents option's definition from CLI and Environment variable.
type flag struct {
	name        string
	description string
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
	defaultValue string
	value        *string
}

// NewStringFlag is a constructor of StringFlag struct.
func NewStringFlag(flagName string, description string, defaultValue string) StringFlag {
	strFlag := StringFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
		defaultValue: defaultValue,
	}

	strFlag.value = app.Flag(flagName, description).
		Default(defaultValue).OverrideDefaultFromEnvar(strFlag.envName()).String()
	isEnvParsed = false

	return strFlag
}

// Value returns value of defined flag after parse.
// NOTE: If conf is not parsed it returns default value (!)
func (s StringFlag) Value() string {
	if !isEnvParsed {
		return s.defaultValue
	}

	return *s.value
}

// IntFlag represents flag with int value.
type IntFlag struct {
	flag
	defaultValue int
	value        *int
}

// NewIntFlag is a constructor of IntFlag struct.
func NewIntFlag(flagName string, description string, defaultValue int) IntFlag {
	intFlag := IntFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
		defaultValue: defaultValue,
	}

	intFlag.value = app.Flag(flagName, description).
		Default(fmt.Sprintf("%d", defaultValue)).OverrideDefaultFromEnvar(intFlag.envName()).Int()
	isEnvParsed = false

	return intFlag
}

// Value returns value of defined flag after parse.
// NOTE: If conf is not parsed it returns default value (!)
func (i IntFlag) Value() int {
	if !isEnvParsed {
		return i.defaultValue
	}

	return *i.value
}

// SliceFlag represents flag with slice value.
type SliceFlag struct {
	flag
	defaultValue []string
	value        *[]string
}

// NewSliceFlag is a constructor of SliceFlag struct.
func NewSliceFlag(flagName string, description string) SliceFlag {
	sliceFlag := SliceFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
	}

	sliceFlag.value = app.Flag(flagName, description).OverrideDefaultFromEnvar(sliceFlag.envName()).Strings()
	isEnvParsed = false

	return sliceFlag
}

// Value returns value of defined flag after parse.
func (s SliceFlag) Value() []string {
	if !isEnvParsed {
		return []string{}
	}

	return *s.value
}
