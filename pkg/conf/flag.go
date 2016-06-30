package conf

import (
	"fmt"
	"net"
	"os"
	"strings"
	"time"
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

// NewFileFlag is a constructor of StringFlag struct which checks if file exists.
func NewFileFlag(flagName string, description string, defaultValue string) StringFlag {
	strFlag := StringFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
		defaultValue: defaultValue,
	}

	strFlag.value = app.Flag(flagName, description).
		Default(defaultValue).OverrideDefaultFromEnvar(strFlag.envName()).ExistingFile()
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

	sliceFlag.value =
		StringList(app.Flag(flagName, description).OverrideDefaultFromEnvar(sliceFlag.envName()))
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

// BoolFlag represents flag with bool value.
type BoolFlag struct {
	flag
	defaultValue bool
	value        *bool
}

// NewBoolFlag is a constructor of BoolFlag struct.
func NewBoolFlag(flagName string, description string, defaultValue bool) BoolFlag {
	boolFlag := BoolFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
		defaultValue: defaultValue,
	}

	boolFlag.value = app.Flag(flagName, description).Default(fmt.Sprintf("%v", defaultValue)).
		OverrideDefaultFromEnvar(boolFlag.envName()).Bool()
	isEnvParsed = false

	return boolFlag
}

// Value returns value of defined flag after parse.
// NOTE: If conf is not parsed it returns default value (!)
func (b BoolFlag) Value() bool {
	if !isEnvParsed {
		return b.defaultValue
	}

	return *b.value
}

// DurationFlag represents flag with duration value.
type DurationFlag struct {
	flag
	defaultValue time.Duration
	value        *time.Duration
}

// NewDurationFlag is a constructor of DurationFlag struct.
func NewDurationFlag(flagName string, description string, defaultValue time.Duration) DurationFlag {
	durationFlag := DurationFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
		defaultValue: defaultValue,
	}

	durationFlag.value = app.Flag(flagName, description).Default(defaultValue.String()).
		OverrideDefaultFromEnvar(durationFlag.envName()).Duration()
	isEnvParsed = false

	return durationFlag
}

// Value returns value of defined flag after parse.
// NOTE: If conf is not parsed it returns default value (!)
func (d DurationFlag) Value() time.Duration {
	if !isEnvParsed {
		return d.defaultValue
	}

	return *d.value
}

// IPFlag represents flag with IP value.
type IPFlag struct {
	flag
	defaultValue string
	value        *net.IP
}

// NewIPFlag is a constructor of IPFlag struct.
func NewIPFlag(flagName string, description string, defaultValue string) IPFlag {
	ipFlag := IPFlag{
		flag: flag{
			name:        flagName,
			description: description,
		},
		defaultValue: defaultValue,
	}

	ip, err := net.ResolveIPAddr("ip4", defaultValue)
	if err != nil {
		panic(err)
	}

	ipFlag.value = app.Flag(flagName, description).Default(ip.String()).
		OverrideDefaultFromEnvar(ipFlag.envName()).IP()
	isEnvParsed = false

	return ipFlag
}

// Value returns value of defined flag after parse.
// NOTE: If conf is not parsed it returns default value (!)
func (i IPFlag) Value() string {
	if !isEnvParsed {
		return i.defaultValue
	}

	return (*i.value).String()
}
