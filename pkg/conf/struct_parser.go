package conf

import (
	"github.com/fatih/camelcase"
	"github.com/pkg/errors"
	"reflect"
	"strconv"
	"strings"
	"time"
)

const (
	// Tag for specifying the help description of the field. [Required]
	helpTag = "help"
	// Tag for specifying default value for field. [Optional]
	defaultTag = "default"
	// Tag for specifying name of the field which contain default value. [Optional]
	defaultFromFieldTag = "defaultFromField"
	// Tag for overriding the name of the field. [Optional]
	nameTag = "name"
	// Tag for specifying that the flag is required. [Optional]
	requiredTag = "required"
	// Tag for specifying that the flag string type means something more concrete. [Optional]
	stringTypeTag = "type"
	// Supported values of above are:
	stringTypeFile = "file"
	stringTypeIP   = "ip"
	// Special field name indicating prefix for all flags in struct.
	prefixFieldName = "flagPrefix"
)

// Process parses given struct and expose flags when config struct tags are present.
// It also gets the values from these flags when CLI or Env is parsed. If not it will
// set the default values.
func Process(data interface{}) error {
	s := &structProcessor{
		data: reflect.ValueOf(data),
	}
	return s.process()
}

func getStringFromField(field reflect.Value) string {
	switch k := field.Kind(); k {
	case reflect.String:
		return field.String()
	}

	return ""
}

type structProcessor struct {
	data reflect.Value

	// Internal variables.
	typeOfData reflect.Type
}

func (s *structProcessor) validate() error {
	if s.data.Kind() != reflect.Ptr {
		return errors.Errorf("Argument needs to be a pointer to struct. Got %s",
			s.data.Kind().String())
	}

	dataValue := s.data.Elem()

	if dataValue.Kind() != reflect.Struct {
		return errors.Errorf("Argument needs to be a pointer to struct. Got %s",
			dataValue.Kind().String())
	}

	return nil
}

// Inspired by github.com/kelseyhightower/envconfig/blob/master/envconfig.go.
func (s *structProcessor) process() error {
	err := s.validate()
	if err != nil {
		return nil
	}

	dataValue := s.data.Elem()
	s.typeOfData = dataValue.Type()

	prefix := getStringFromField(dataValue.FieldByName(prefixFieldName))

	// Process each field in struct.
	for i := 0; i < dataValue.NumField(); i++ {
		field := dataValue.Field(i)
		if !field.CanSet() {
			continue
		}

		// Embedded field are not supported yet (nested processing)
		if s.typeOfData.Field(i).Anonymous && field.Kind() == reflect.Struct {
			continue
		}

		// Process a field.
		f := &fieldProcessor{
			prefix:      prefix,
			data:        dataValue,
			field:       field,
			fieldStruct: s.typeOfData.Field(i),
		}
		err = f.process()
		if err != nil {
			return err
		}
	}
	return nil
}

func nameFromFieldName(name string) string {
	// Parse the name e.g SomeSome to some_some.
	words := camelcase.Split(name)
	wordsToUse := []string{}
	for _, word := range words {
		if word == "_" {
			continue
		}
		wordsToUse = append(wordsToUse, strings.ToLower(word))
	}

	return strings.Join(wordsToUse, "_")
}

type fieldProcessor struct {
	prefix      string
	data        reflect.Value
	field       reflect.Value
	fieldStruct reflect.StructField
}

func (f *fieldProcessor) isAnyTagSpecified() bool {
	tags := []string{
		nameTag,
		defaultTag,
		defaultFromFieldTag,
		requiredTag,
		stringTypeTag,
		helpTag,
	}
	for _, tag := range tags {
		if f.fieldStruct.Tag.Get(tag) != "" {
			return true
		}
	}

	return false
}

func (f *fieldProcessor) getHelpMessage() (string, error) {
	help := f.fieldStruct.Tag.Get(helpTag)
	if help == "" {
		if f.isAnyTagSpecified() {
			return "", errors.New("Required help tag is missing. Cannot process the struct for flags.")
		}

		// If help is not specified and not tag was used than this field is just excluded from processing.
		return "", nil
	}

	return help, nil
}

func (f *fieldProcessor) getFlagName() string {
	name := f.fieldStruct.Tag.Get(nameTag)
	if name == "" {
		name = f.fieldStruct.Name
	}
	return nameFromFieldName(f.prefix + name)
}

func (f *fieldProcessor) getDefaultValue() string {
	defaultValue := f.fieldStruct.Tag.Get(defaultTag)
	if defaultValue == "" {
		// If default value is not present, check defaultFromField tag.
		fieldWithDefaultValue := f.fieldStruct.Tag.Get(defaultFromFieldTag)
		// Fetch default value from specified field.
		defaultValue = getStringFromField(f.data.FieldByName(fieldWithDefaultValue))
	}

	return defaultValue
}

func (f *fieldProcessor) isDurationType() bool {
	return f.field.Kind() == reflect.Int64 &&
		f.field.Type().PkgPath() == "time" &&
		f.field.Type().Name() == "Duration"
}

func (f *fieldProcessor) process() error {
	// Parse Help.
	help, err := f.getHelpMessage()
	if err != nil {
		return err
	}

	if help == "" {
		// Exclude this field from processing.
		return nil
	}

	// Parsing optional Name override.
	// Prefix and name from tag are being parsed to lowercase and camelCase being split with "_".
	name := f.getFlagName()

	// Parse default or defaultFromField tag if present.
	defaultValue := f.getDefaultValue()

	// Fetch the type of the field.
	typeOfField := f.field.Type()

	// Get a proper type in case of a pointer.
	// NOTE: In case of not initialized pointer we need to allocate a new one.
	if typeOfField.Kind() == reflect.Ptr {
		typeOfField = typeOfField.Elem()
		if f.field.IsNil() {
			f.field.Set(reflect.New(typeOfField))
		}
		f.field = f.field.Elem()
	}

	var flagClause *cliAndEnvFlag

	// Define flag and insert default values for each type.
	// TODO(bp): It might be worth to introduce here a way to support custom types (strategy patterns).
	switch typeOfField.Kind() {
	case reflect.String:
		// Switch between different string types.
		switch f.fieldStruct.Tag.Get(stringTypeTag) {
		case stringTypeFile:
			flag := NewFileFlag(name, help, defaultValue)
			f.field.SetString(flag.Value())
			flagClause = flag.cliAndEnvFlag
		case stringTypeIP:
			flag := NewIPFlag(name, help, defaultValue)
			f.field.SetString(flag.Value())
			flagClause = flag.cliAndEnvFlag
		default:
			flag := NewStringFlag(name, help, defaultValue)
			f.field.SetString(flag.Value())
			flagClause = flag.cliAndEnvFlag
		}
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		if f.isDurationType() {
			var defaultDurationValue time.Duration
			var err error
			if defaultValue != "" {
				defaultDurationValue, err = time.ParseDuration(defaultValue)
				if err != nil {
					return errors.Wrap(err, "Wrong default value for Duration type flag")
				}
			}

			flag := NewDurationFlag(name, help, defaultDurationValue)
			f.field.SetInt(int64(flag.Value()))
			flagClause = flag.cliAndEnvFlag
		} else {
			var defaultIntValue int
			var err error
			if defaultValue != "" {
				defaultIntValue, err = strconv.Atoi(defaultValue)
				if err != nil {
					return errors.Wrap(err, "Wrong default value for Int type flag")
				}
			}

			flag := NewIntFlag(name, help, defaultIntValue)
			f.field.SetInt(int64(flag.Value()))
			flagClause = flag.cliAndEnvFlag
		}
	case reflect.Bool:
		var defaultBoolValue bool
		var err error
		if defaultValue != "" {
			defaultBoolValue, err = strconv.ParseBool(defaultValue)
			if err != nil {
				return errors.Wrap(err, "Wrong default value for Bool type flag")
			}
		}

		flag := NewBoolFlag(name, help, defaultBoolValue)
		f.field.SetBool(flag.Value())
		flagClause = flag.cliAndEnvFlag
	case reflect.Slice:
		if typeOfField != reflect.TypeOf([]string(nil)) {
			return errors.Errorf("%s type not supported for a slice flag", typeOfField.String())
		}

		var defaultSliceValue StringListValue
		var err error
		if defaultValue != "" {
			err = (&defaultSliceValue).Set(defaultValue)
			if err != nil {
				return errors.Wrap(err, "Wrong default value for String Slice type flag")
			}
		}

		flag := NewSliceFlag(name, help, defaultSliceValue...)
		slice := reflect.MakeSlice(typeOfField, len(flag.Value()), len(flag.Value()))
		for i, value := range flag.Value() {
			slice.Index(i).SetString(value)
		}

		f.field.Set(slice)
		flagClause = flag.cliAndEnvFlag
	default:
		return errors.Errorf("%s type not supported for a flag", typeOfField.String())
	}

	// Get required tag if present.
	requiredTag := f.fieldStruct.Tag.Get(requiredTag)
	if requiredTag == "true" {
		// Mark flag as required.
		flagClause.Required()
	}

	return nil
}
