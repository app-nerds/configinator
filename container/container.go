package container

import (
	"flag"
	"fmt"
	"os"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// Supported struct tags
const (
	TagFlagName     string = "flag"
	TagEnvName      string = "env"
	TagDefaultValue string = "default"
	TagDescription  string = "description"
)

// Custom errors
var (
	ErrNoFlagName = fmt.Errorf("no flag name")
	ErrCantSet    = fmt.Errorf("can't set private fields")

	timeFormats = []string{
		"2006-01-02",
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02T15:04:05Z",
		"2006-01-02T15:04:05 MST",
		"2006-01-02T15:04:05-0700",
	}
)

/*
Container is a host to a given struct field and it's tag configuration. It is
here where the logic to get values and determine if values are set as flags,
env, etc.. is done.
*/
type Container struct {
	boolValue    *bool
	config       interface{}
	configValue  reflect.Value
	defaultValue string
	description  string
	envFile      map[string]string
	envName      string
	field        reflect.StructField
	fieldName    string
	fieldType    string
	fieldValue   reflect.Value
	flagName     string
	floatValue   *float64
	intValue     *int
	stringValue  *string
	timeValue    *string
}

/*
New creates a new Container. This will verify that the struct
field can be set and has the required tags.
*/
func New(config interface{}, index int, envFile map[string]string) (*Container, error) {
	var (
		hasFlag bool
	)

	t := reflect.TypeOf(config).Elem()

	result := &Container{
		config:    config,
		envFile:   envFile,
		fieldType: strings.ToLower(t.Field(index).Type.String()),

		configValue: reflect.ValueOf(config).Elem(),
		field:       t.Field(index),
		fieldName:   t.Field(index).Name,
	}

	/*
	 * If this field doesn't have a flag name, or is private and
	 * cannot be set, return an error
	 */
	canSet := result.configValue.Field(index).CanSet()

	if !canSet {
		return result, ErrCantSet
	}

	result.flagName, hasFlag = result.field.Tag.Lookup(TagFlagName)

	if !hasFlag {
		return result, ErrNoFlagName
	}

	result.fieldValue = result.configValue.Field(index)
	result.envName = result.field.Tag.Get(TagEnvName)
	result.defaultValue = result.field.Tag.Get(TagDefaultValue)
	result.description = result.field.Tag.Get(TagDescription)

	if !flag.Parsed() {
		result.addFlag()
	}
	result.SetDefaultValueOnConfig()
	return result, nil
}

func (c *Container) EnvBool() (bool, bool) {
	value := os.Getenv(c.envName)

	if value != "" {
		if result, err := strconv.ParseBool(value); err == nil {
			return result, true
		}
	}

	return false, false
}

func (c *Container) EnvFloat() (float64, bool) {
	value := os.Getenv(c.envName)

	if value != "" {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result, true
		}
	}

	return 0.0, false
}

func (c *Container) EnvInt() (int, bool) {
	value := os.Getenv(c.envName)

	if value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result, true
		}
	}

	return 0, false
}

func (c *Container) EnvString() (string, bool) {
	value := os.Getenv(c.envName)

	if value != "" {
		return value, true
	}

	return value, false
}

func (c *Container) EnvTime() (time.Time, bool) {
	value := os.Getenv(c.envName)

	if value != "" {
		return c.parseTime(value), true
	}

	return time.Time{}, false
}

func (c *Container) EnvFileBool() (bool, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := strconv.ParseBool(value); err == nil {
			return result, true
		}
	}

	return false, false
}

func (c *Container) EnvFileFloat() (float64, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result, true
		}
	}

	return 0.0, false
}

func (c *Container) EnvFileInt() (int, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		if result, err := strconv.Atoi(value); err == nil {
			return result, true
		}
	}

	return 0, false
}

func (c *Container) EnvFileString() (string, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		return value, true
	}

	return "", false
}

func (c *Container) EnvFileTime() (time.Time, bool) {
	if value, ok := c.envFile[c.envName]; ok {
		return c.parseTime(value), true
	}

	return time.Time{}, false
}

func (c *Container) FlagBool() (bool, bool) {
	if c.boolValue != nil && *c.boolValue != c.defaultValueToBool() {
		return *c.boolValue, true
	}

	return false, false
}

func (c *Container) FlagFloat() (float64, bool) {
	if c.floatValue != nil && *c.floatValue != c.defaultValueToFloat() {
		return *c.floatValue, true
	}

	return 0.0, false
}

func (c *Container) FlagInt() (int, bool) {
	if c.intValue != nil && *c.intValue != c.defaultValueToInt() {
		return *c.intValue, true
	}

	return 0, false
}

func (c *Container) FlagString() (string, bool) {
	if c.stringValue != nil && *c.stringValue != c.defaultValueToString() {
		return *c.stringValue, true
	}

	return "", false
}

func (c *Container) FlagTime() (time.Time, bool) {
	if c.timeValue != nil && *c.timeValue != c.defaultValue {
		return c.parseTime(*c.timeValue), true
	}

	return time.Time{}, false
}

func (c *Container) IsBool() bool {
	return c.fieldType == "bool"
}

func (c *Container) IsFloat() bool {
	return c.fieldType == "float64"
}

func (c *Container) IsInt() bool {
	return c.fieldType == "int"
}

func (c *Container) IsString() bool {
	return c.fieldType == "string"
}

func (c *Container) IsTime() bool {
	return c.fieldType == "time.time"
}

func (c *Container) SetConfigBool(value bool) {
	c.fieldValue.SetBool(value)
}

func (c *Container) SetConfigFloat(value float64) {
	c.fieldValue.SetFloat(value)
}

func (c *Container) SetConfigInt(value int) {
	c.fieldValue.SetInt(int64(value))
}

func (c *Container) SetConfigString(value string) {
	c.fieldValue.SetString(value)
}

func (c *Container) SetConfigTime(value time.Time) {
	c.fieldValue.Set(reflect.ValueOf(value))
}

func (c *Container) SetDefaultValueOnConfig() {
	if c.IsBool() {
		c.SetConfigBool(c.defaultValueToBool())
	}

	if c.IsFloat() {
		c.SetConfigFloat(c.defaultValueToFloat())
	}

	if c.IsInt() {
		c.SetConfigInt(c.defaultValueToInt())
	}

	if c.IsString() {
		c.SetConfigString(c.defaultValueToString())
	}

	if c.IsTime() {
		c.SetConfigTime(c.defaultValueToTime())
	}
}

func (c *Container) addFlag() {
	if c.IsBool() {
		c.boolValue = flag.Bool(c.flagName, c.defaultValueToBool(), c.description)
	}

	if c.IsFloat() {
		c.floatValue = flag.Float64(c.flagName, c.defaultValueToFloat(), c.description)
	}

	if c.IsInt() {
		c.intValue = flag.Int(c.flagName, c.defaultValueToInt(), c.description)
	}

	if c.IsString() {
		c.stringValue = flag.String(c.flagName, c.defaultValueToString(), c.description)
	}

	if c.IsTime() {
		c.timeValue = flag.String(c.flagName, c.defaultValueToString(), c.description)
	}
}

func (c *Container) defaultValueToBool() bool {
	var (
		err    error
		result bool
	)

	if result, err = strconv.ParseBool(c.defaultValue); err != nil {
		return false
	}

	return result
}

func (c *Container) defaultValueToFloat() float64 {
	var (
		err    error
		result float64
	)

	if result, err = strconv.ParseFloat(c.defaultValue, 64); err != nil {
		return 0.0
	}

	return result
}

func (c *Container) defaultValueToInt() int {
	var (
		err    error
		result int
	)

	if result, err = strconv.Atoi(c.defaultValue); err != nil {
		return 0
	}

	return result
}

func (c *Container) defaultValueToString() string {
	return c.defaultValue
}

func (c *Container) defaultValueToTime() time.Time {
	if !c.isTime(c.defaultValue) {
		return time.Time{}
	}

	result := c.parseTime(c.defaultValue)
	return result
}

func (c *Container) isTime(value string) bool {
	for _, f := range timeFormats {
		if _, err := time.Parse(f, value); err == nil {
			return true
		}
	}

	return false
}

func (c *Container) parseTime(value string) time.Time {
	for _, f := range timeFormats {
		if t, err := time.Parse(f, value); err == nil {
			return t
		}
	}

	return time.Time{}
}
