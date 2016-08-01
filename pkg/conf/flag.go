package conf

import (
	"fmt"
	"gopkg.in/alecthomas/kingpin.v2"
	"net"
	"os"
	"strings"
	"time"
)

// flagType is an internal interface for all flags.
// Every flag should have method for creating `envName` from its name and `clear` method
// for clearing corresponding environment variable from env.
type flagType interface {
	envName() string
	clear()
}

// definedFlags is a package variable which stores all the defined flags. It helps to find
// duplicates when defining flag with the same name.
var definedFlags = map[string]flagType{}

// flag represents option's definition from CLI and Environment variable.
// It stores generic data for each defined flag.
// It implements swan flagType interface.
type cliAndEnvFlag struct {
	*kingpin.FlagClause
}

func newCliAndEnvFlag(flagName string, description string, defaultValues ...string) *cliAndEnvFlag {
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		panic("This flag was already defined. Flag definition is lack of duplicate check.")
	}

	c := &cliAndEnvFlag{FlagClause: app.Flag(flagName, description)}
	c.OverrideDefaultFromEnvar(c.envName())

	for _, defaultValue := range defaultValues {
		if defaultValue == "" {
			continue
		}
		c.Default(defaultValue)
	}

	return c
}

// envName returns name converted to swan environment variable name.
// In order to create environment variable name from flag we need to make it uppercase
// and add SWAN prefix. For instance: "cassandra_host" will be "SWAN_CASSANDRA_HOST".
func (f *cliAndEnvFlag) envName() string {
	return fmt.Sprintf("%s_%s", "SWAN", strings.ToUpper(f.Model().Name))
}

// clear unset the corresponded environment variable.
func (f *cliAndEnvFlag) clear() {
	os.Unsetenv(f.envName())
}

// StringFlag represents flag with string value.
type StringFlag struct {
	*cliAndEnvFlag
	defaultValue string
	value        *string
}

// NewStringFlag is a constructor of StringFlag struct.
func NewStringFlag(flagName string, description string, defaultValue string) *StringFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*StringFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		if flagDef.defaultValue != defaultValue {
			panic("Flag was redefined but with different default value. Unify the default.")
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &StringFlag{
		cliAndEnvFlag: newCliAndEnvFlag(flagName, description, defaultValue),
		defaultValue:  defaultValue,
	}

	// Define type of the flag and register in internal map.
	flagDef.value = flagDef.String()
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
}

// FileFlag represents flag with string value.
type FileFlag struct {
	*StringFlag
}

// NewFileFlag is a constructor of StringFlag struct which checks if file exists.
func NewFileFlag(flagName string, description string, defaultValue string) *FileFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*FileFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		if flagDef.defaultValue != defaultValue {
			panic("Flag was redefined but with different default value. Unify the default.")
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &FileFlag{
		StringFlag: &StringFlag{
			cliAndEnvFlag: newCliAndEnvFlag(flagName, description, defaultValue),
			defaultValue:  defaultValue,
		},
	}

	// Define type of the flag and register in internal map.
	flagDef.value = flagDef.ExistingFile()
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
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
	*cliAndEnvFlag
	defaultValue int
	value        *int
}

// NewIntFlag is a constructor of IntFlag struct.
func NewIntFlag(flagName string, description string, defaultValue int) *IntFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*IntFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		if flagDef.defaultValue != defaultValue {
			panic("Flag was redefined but with different default value. Unify the default.")
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &IntFlag{
		cliAndEnvFlag: newCliAndEnvFlag(flagName, description, fmt.Sprintf("%d", defaultValue)),
		defaultValue:  defaultValue,
	}

	// Define type of the flag and register in internal map.
	flagDef.value = flagDef.Int()
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
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
	*cliAndEnvFlag
	defaultValue []string
	value        *[]string
}

// NewSliceFlag is a constructor of SliceFlag struct.
func NewSliceFlag(flagName string, description string, elemsInDefaultSlice ...string) *SliceFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*SliceFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		for i, elem := range elemsInDefaultSlice {
			if flagDef.defaultValue[i] != elem {
				panic("Flag was redefined but with different default value. Unify the default.")
			}
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &SliceFlag{
		cliAndEnvFlag: newCliAndEnvFlag(flagName, description, strings.Join(elemsInDefaultSlice, ",")),
		defaultValue:  elemsInDefaultSlice,
	}

	// Define type of the flag and register in internal map.
	flagDef.value = StringList(flagDef)
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
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
	*cliAndEnvFlag
	defaultValue bool
	value        *bool
}

// NewBoolFlag is a constructor of BoolFlag struct.
func NewBoolFlag(flagName string, description string, defaultValue bool) *BoolFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*BoolFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		if flagDef.defaultValue != defaultValue {
			panic("Flag was redefined but with different default value. Unify the default.")
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &BoolFlag{
		cliAndEnvFlag: newCliAndEnvFlag(flagName, description, fmt.Sprintf("%v", defaultValue)),
		defaultValue:  defaultValue,
	}

	// Define type of the flag and register in internal map.
	flagDef.value = flagDef.Bool()
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
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
	*cliAndEnvFlag
	defaultValue time.Duration
	value        *time.Duration
}

// NewDurationFlag is a constructor of DurationFlag struct.
func NewDurationFlag(flagName string, description string, defaultValue time.Duration) *DurationFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*DurationFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		if flagDef.defaultValue != defaultValue {
			panic("Flag was redefined but with different default value. Unify the default.")
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &DurationFlag{
		cliAndEnvFlag: newCliAndEnvFlag(flagName, description, defaultValue.String()),
		defaultValue:  defaultValue,
	}

	// Define type of the flag and register in internal map.
	flagDef.value = flagDef.Duration()
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
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
	*cliAndEnvFlag
	defaultValue string
	value        *net.IP
}

// NewIPFlag is a constructor of IPFlag struct.
func NewIPFlag(flagName string, description string, defaultValue string) *IPFlag {
	// Check for duplicates and use it if it defines the same type of flag.
	duplicatedFlag := definedFlags[flagName]
	if duplicatedFlag != nil {
		// Check if the type is the same.
		flagDef, ok := duplicatedFlag.(*IPFlag)
		if !ok {
			panic("Flag was redefined but with different type. Unify the type.")
		}

		if flagDef.defaultValue != defaultValue {
			panic("Flag was redefined but with different default value. Unify the default.")
		}

		return flagDef
	}

	// Flag is not yet defined, so create one.
	flagDef := &IPFlag{
		cliAndEnvFlag: newCliAndEnvFlag(flagName, description, net.ParseIP(defaultValue).String()),
		defaultValue:  defaultValue,
	}

	// Define type of the flag and register in internal map.
	flagDef.value = flagDef.IP()
	definedFlags[flagName] = flagDef
	isEnvParsed = false
	return flagDef
}

// Value returns value of defined flag after parse.
// NOTE: If conf is not parsed it returns default value (!)
func (i IPFlag) Value() string {
	if !isEnvParsed {
		return i.defaultValue
	}

	return (*i.value).String()
}
