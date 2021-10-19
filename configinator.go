package configinator

import (
	"flag"
	"fmt"
	"reflect"
	"strconv"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

/*
New initializes a provided struct with values from defaults,
environment, and flags. It does this by adding tags to your
struct. For example:

  type Config struct {
	  Host `flag:"host" env:"HOST" default:"localhost:8080" description:"Host and port to bind to"`
  }

The above example will accept a command line flag of "host",
or an environment variable named "HOST". If none of the above
are provided then the value from 'default' is used.

If an .env file is found that will be read and used.
*/
func New(config interface{}) {
	var (
		err error
	)

	t := reflect.TypeOf(config)

	for index := 0; index < t.NumField(); index++ {
		field := t.Field(index)
		fieldType := strings.ToLower(field.Type.String())
		tagFlag := field.Tag.Get("flag")
		tagEnv := field.Tag.Get("env")
		tagDefault := field.Tag.Get("default")
		tagDescription := field.Tag.Get("description")

		getValue(fieldType, tagFlag, tagEnv, tagDefault, tagDescription)
	}

	pflag.CommandLine.AddGoFlagSet(flag.CommandLine)
	pflag.Parse()
	viper.BindPFlags(pflag.CommandLine)

	viper.SetConfigFile(".env")
	viper.SetConfigType("env")
	viper.AddConfigPath(".")

	if err = viper.ReadInConfig(); err != nil && !strings.Contains(err.Error(), "no such file") {
		fmt.Printf("error reading configuration file '.env'")
		panic(err)
	}

	if err = viper.Unmarshal(&config); err != nil {
		panic(err)
	}
}

func getValue(fieldType, flagName, envName, defaultValue, description string) {
	var (
		err        error
		boolValue  bool
		floatValue float64
		intValue   int
	)

	switch fieldType {
	case "bool":
		if boolValue, err = strconv.ParseBool(defaultValue); err != nil {
			panic("config field '" + flagName + "' isn't a bool")
		}

		viper.SetDefault(flagName, boolValue)
		_ = flag.Bool(flagName, boolValue, description)

	case "int":
	case "int32":
	case "int64":
		if intValue, err = strconv.Atoi(defaultValue); err != nil {
			panic("config field '" + flagName + "' isn't an int")
		}

		viper.SetDefault(flagName, intValue)
		_ = flag.Int(flagName, intValue, description)

	case "float":
	case "float64":
		if floatValue, err = strconv.ParseFloat(defaultValue, 64); err != nil {
			panic("config field '" + flagName + "' isn't a float64")
		}

		viper.SetDefault(flagName, floatValue)
		_ = flag.Float64(flagName, floatValue, description)

	default:
		viper.SetDefault(flagName, defaultValue)
		_ = flag.String(flagName, defaultValue, description)
	}

	_ = viper.BindEnv(flagName, envName)
}
