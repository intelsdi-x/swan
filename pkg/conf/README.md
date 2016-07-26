Swan Configuration
===============================

This package introduces a way to specify configuration flags for an application.
These flags can be fetched from the command line as well as from environment variables.
`conf` package allows to define flags explicitly using `New<type>Flag(name, help, default)` or using `struct tags`.
This package uses `kinping.v2` go package underneath.

*Supported types*:
- string
  - just a string
  - file (checks if file exists)
  - ip (checks if it is parsable IP address)
- bool
- int
- []string

To parse only environment variables use:

```go
err := conf.ParseEnv()
```

To parse both CLI and environment use:

```go
 err := conf.ParseFlags()
```

It is recommend to use `struct tags` for defining a flag.

## Usage of struct tags for flag specification

To expose a field as a flag using `conf` package, you need to specify some unique struct tag after the field.

Required tag:
- help: It is used for defining a help message for the flag. If it is not specified then this field
will be omitted and _no flag will be exposed_.

Optional tags:
- default: Default value of the flag when not specified in env or CLI.
- defaultFromField: Name of the field in current struct where the default value for the flag is defined.
It is extremely useful when default value is calculated in run time (e.g based on some environment variable)
- name: Overrides name of the flag. By default it takes name of the field and change camelCase to lower case snake_case.
It is suggested to make the value of this tag uppercase to work properly with the prefix
- required: If specified, the flag will be required. If there is no default value and it is not specified in
env or CLI, parse will return error.
- type: String type can have different attributes. It can be more concrete like flag (checking if file exists) or ip (checking if the value can be parsed to an IP address).

When field `flagPrefix` exists then its value will be used to prefix names of all the flags defined in the struct.

### Example struct

```go
type SomeConfig struct {
	StringArg         string `help:"test string" default:"default_string"`
	StringArg2        string `help:"test string" defaultFromField:"defaultStringArg2"`
	defaultStringArg2 string
	RequiredStringArg string `help:"test required string" required:"true"`
	ExcludedStringArg string `default:"default_string2"`

	IntArg         int           `help:"test int" default:"2"`
	DurationArg    time.Duration `help:"test duration" default:"5s"`
	BoolArg        bool          `help:"test bool" default:"true"`
	StringSliceArg []string      `help:"test slice"`
	FileArg        string        `help:"test file" type:"file" default:"/etc/shadow"`
	IPArg          string        `help:"test IP" type:"ip" default:"255.255.255.255"`

	// Prefix optional field.
	flagPrefix string
}

```

### Suggested way to register the struct

The below snippet should be placed in the beginning of the file where the processed struct (here: `SomeConfig`) is defined.

```go
var defaultConfig = SomeConfig{
	defaultStringArg2: "some value",
	flagPrefix:        "demo",
}

func init() {
	conf.Process(&defaultConfig)
}

func DefaultConfig() SomeConfig {
	conf.Process(&defaultConfig)
	return defaultConfig
}
```

### Output of help

```
Flags:
  --help                         Show context-sensitive help (also try --help-long and --help-man).
  --log="error"                  Log level for Swan: debug, info, warn, error, fatal, panic
  --demo_string_arg="default_string"
                                 test string
  --demo_string_arg_2="some value"
                                 test string
  --demo_required_string_arg=DEMO_REQUIRED_STRING_ARG
                                 test required string
  --demo_int_arg=2               test int
  --demo_duration_arg=5s         test duration
  --demo_bool_arg                test bool
  --demo_string_slice_arg=DEMO_STRING_SLICE_ARG ...
                                 test slice
  --demo_file_arg=/etc/shadow    test file
  --demo_ip_arg=255.255.255.255  test IP
```