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
	config       interface{}
	defaultValue string
	envFile      map[string]string
	FieldType    string
	fieldValue   reflect.Value

	ConfigValue reflect.Value
	Field       reflect.StructField
	FieldName   string
	FlagName    string
	EnvName     string
	description string

	BoolValue   *bool
	IntValue    *int
	FloatValue  *float64
	StringValue *string
	TimeValue   *string
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
		FieldType: strings.ToLower(t.Field(index).Type.String()),

		ConfigValue: reflect.ValueOf(config).Elem(),
		Field:       t.Field(index),
		FieldName:   t.Field(index).Name,
	}

	/*
	 * If this field doesn't have a flag name, or is private and
	 * cannot be set, return an error
	 */
	canSet := result.ConfigValue.Field(index).CanSet()

	if !canSet {
		return result, ErrCantSet
	}

	result.FlagName, hasFlag = result.Field.Tag.Lookup(TagFlagName)

	if !hasFlag {
		return result, ErrNoFlagName
	}

	result.fieldValue = result.ConfigValue.Field(index)
	result.EnvName = result.Field.Tag.Get(TagEnvName)
	result.defaultValue = result.Field.Tag.Get(TagDefaultValue)
	result.description = result.Field.Tag.Get(TagDescription)

	result.addFlag()
	result.SetDefaultValueOnConfig()
	return result, nil
}

func (c *Container) DefaultValueToBool() bool {
	var (
		err    error
		result bool
	)

	if result, err = strconv.ParseBool(c.defaultValue); err != nil {
		return false
	}

	return result
}

func (c *Container) DefaultValueToFloat() float64 {
	var (
		err    error
		result float64
	)

	if result, err = strconv.ParseFloat(c.defaultValue, 64); err != nil {
		return 0.0
	}

	return result
}

func (c *Container) DefaultValueToInt() int {
	var (
		err    error
		result int
	)

	if result, err = strconv.Atoi(c.defaultValue); err != nil {
		return 0
	}

	return result
}

func (c *Container) DefaultValueToString() string {
	return c.defaultValue
}

func (c *Container) DefaultValueToTime() time.Time {
	if !c.isTime(c.defaultValue) {
		return time.Time{}
	}

	result := c.parseTime(c.defaultValue)
	return result
}

func (c *Container) EnvBool() (bool, bool) {
	value := os.Getenv(c.EnvName)

	if value != "" {
		if result, err := strconv.ParseBool(value); err == nil {
			return result, true
		}
	}

	return false, false
}

func (c *Container) EnvFloat() (float64, bool) {
	value := os.Getenv(c.EnvName)

	if value != "" {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result, true
		}
	}

	return 0.0, false
}

func (c *Container) EnvInt() (int, bool) {
	value := os.Getenv(c.EnvName)

	if value != "" {
		if result, err := strconv.Atoi(value); err == nil {
			return result, true
		}
	}

	return 0, false
}

func (c *Container) EnvString() (string, bool) {
	value := os.Getenv(c.EnvName)

	if value != "" {
		return value, true
	}

	return value, false
}

func (c *Container) EnvTime() (time.Time, bool) {
	value := os.Getenv(c.EnvName)

	if value != "" {
		return c.parseTime(value), true
	}

	return time.Time{}, false
}

func (c *Container) EnvFileBool() (bool, bool) {
	if value, ok := c.envFile[c.EnvName]; ok {
		if result, err := strconv.ParseBool(value); err == nil {
			return result, true
		}
	}

	return false, false
}

func (c *Container) EnvFileFloat() (float64, bool) {
	if value, ok := c.envFile[c.EnvName]; ok {
		if result, err := strconv.ParseFloat(value, 64); err == nil {
			return result, true
		}
	}

	return 0.0, false
}

func (c *Container) EnvFileInt() (int, bool) {
	if value, ok := c.envFile[c.EnvName]; ok {
		if result, err := strconv.Atoi(value); err == nil {
			return result, true
		}
	}

	return 0, false
}

func (c *Container) EnvFileString() (string, bool) {
	if value, ok := c.envFile[c.EnvName]; ok {
		return value, true
	}

	return "", false
}

func (c *Container) EnvFileTime() (time.Time, bool) {
	if value, ok := c.envFile[c.EnvName]; ok {
		return c.parseTime(value), true
	}

	return time.Time{}, false
}

func (c *Container) FlagBool() (bool, bool) {
	if c.BoolValue != nil && *c.BoolValue != c.DefaultValueToBool() {
		return *c.BoolValue, true
	}

	return false, false
}

func (c *Container) FlagFloat() (float64, bool) {
	if c.FloatValue != nil && *c.FloatValue != c.DefaultValueToFloat() {
		return *c.FloatValue, true
	}

	return 0.0, false
}

func (c *Container) FlagInt() (int, bool) {
	if c.IntValue != nil && *c.IntValue != c.DefaultValueToInt() {
		return *c.IntValue, true
	}

	return 0, false
}

func (c *Container) FlagString() (string, bool) {
	if c.StringValue != nil && *c.StringValue != c.DefaultValueToString() {
		return *c.StringValue, true
	}

	return "", false
}

func (c *Container) FlagTime() (time.Time, bool) {
	if c.TimeValue != nil && *c.TimeValue != c.defaultValue {
		return c.parseTime(*c.TimeValue), true
	}

	return time.Time{}, false
}

func (c *Container) IsBool() bool {
	return c.FieldType == "bool"
}

func (c *Container) IsFloat() bool {
	return c.FieldType == "float64"
}

func (c *Container) IsInt() bool {
	return c.FieldType == "int"
}

func (c *Container) IsString() bool {
	return c.FieldType == "string"
}

func (c *Container) IsTime() bool {
	return c.FieldType == "time.time"
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
		c.SetConfigBool(c.DefaultValueToBool())
	}

	if c.IsFloat() {
		c.SetConfigFloat(c.DefaultValueToFloat())
	}

	if c.IsInt() {
		c.SetConfigInt(c.DefaultValueToInt())
	}

	if c.IsString() {
		c.SetConfigString(c.DefaultValueToString())
	}

	if c.IsTime() {
		c.SetConfigTime(c.DefaultValueToTime())
	}
}

func (c *Container) addFlag() {
	if c.IsBool() {
		c.BoolValue = flag.Bool(c.FlagName, c.DefaultValueToBool(), c.description)
	}

	if c.IsFloat() {
		c.FloatValue = flag.Float64(c.FlagName, c.DefaultValueToFloat(), c.description)
	}

	if c.IsInt() {
		c.IntValue = flag.Int(c.FlagName, c.DefaultValueToInt(), c.description)
	}

	if c.IsString() {
		c.StringValue = flag.String(c.FlagName, c.DefaultValueToString(), c.description)
	}

	if c.IsTime() {
		c.TimeValue = flag.String(c.FlagName, c.DefaultValueToString(), c.description)
	}
}

func (c *Container) parseTime(value string) time.Time {
	for _, f := range timeFormats {
		if t, err := time.Parse(f, value); err == nil {
			return t
		}
	}

	return time.Time{}
}

func (c *Container) isTime(value string) bool {
	for _, f := range timeFormats {
		if _, err := time.Parse(f, value); err == nil {
			return true
		}
	}

	return false
}
