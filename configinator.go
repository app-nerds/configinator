package configinator

import (
	"flag"
	"os"
	"reflect"

	"github.com/app-nerds/configinator/container"
	"github.com/app-nerds/configinator/env"
)

var (
	envFile map[string]string
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
func Behold(config interface{}) {
	var (
		err        error
		index      int
		containers []*container.Container
	)

	envFile = make(map[string]string)

	/*
	 * If we have an environment file, load it
	 */
	if env.FileExists(".env") {
		if envFile, err = env.ReadFile(".env"); err != nil {
			panic(err)
		}
	}

	/*
	 * Read the type info for this struct
	 */
	t := reflect.TypeOf(config).Elem()
	containers = make([]*container.Container, t.NumField())

	/*
	 * First setup each field of the config struct. These are stored in "containers".
	 * Each container know the field type, value, env name, flag name, and adds
	 * to the provided flag set.
	 */
	for index = 0; index < t.NumField(); index++ {
		containers[index], _ = container.New(config, index, envFile)
	}

	/*
	 * Parse flags
	 */
	if len(os.Args) > 1 {
		flag.Parse()
	}

	/*
	 * Set the values in the config struct. They already have default value set.
	 * So first we check to see if there is an environment variable. Then we
	 * check to see if there is an environment file value. Finally we check for a
	 * flag value.
	 */
	for index = 0; index < t.NumField(); index++ {
		c := containers[index]

		if c.IsBool() {
			if value, ok := c.EnvBool(); ok {
				c.SetConfigBool(value)
			}

			if value, ok := c.EnvFileBool(); ok {
				c.SetConfigBool(value)
			}

			if value, ok := c.FlagBool(); ok {
				c.SetConfigBool(value)
			}
		}

		if c.IsFloat() {
			if value, ok := c.EnvFloat(); ok {
				c.SetConfigFloat(value)
			}

			if value, ok := c.EnvFileFloat(); ok {
				c.SetConfigFloat(value)
			}

			if value, ok := c.FlagFloat(); ok {
				c.SetConfigFloat(value)
			}
		}

		if c.IsInt() {
			if value, ok := c.EnvInt(); ok {
				c.SetConfigInt(value)
			}

			if value, ok := c.EnvFileInt(); ok {
				c.SetConfigInt(value)
			}

			if value, ok := c.FlagInt(); ok {
				c.SetConfigInt(value)
			}
		}

		if c.IsString() {
			if value, ok := c.EnvString(); ok {
				c.SetConfigString(value)
			}

			if value, ok := c.EnvFileString(); ok {
				c.SetConfigString(value)
			}

			if value, ok := c.FlagString(); ok {
				c.SetConfigString(value)
			}
		}
	}
}
