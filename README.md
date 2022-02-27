# Configinator

![Doofenshmirtz Inator](inator.jpg)

Behold, the **Configinator**! Simply create a struct, annotate it with tags, and **BOOM**! Configuration!

Installation is easy. 

```bash
go get github.com/app-nerds/configinator
```

### Example

```go
import (
  "github.com/app-nerds/configinator"
)

type Config struct {
  Host string `flag:"host" env:"HOST" default:"localhost:8080" description:"Host and port to bind to"`
}

func GetConfig() *Config {
  result := Config{}
  configinator.Behold(&result)
  return &result
}
```

## How It Works

The Configinator reads tags on your structs to get configuration data. As per the rules of Go, only exported fields will be considered. Furthermore you must pass a pointer to the struct to the Configinator. So what does it do? The Configinator will look for configuration data from the following sources, in this order (the last location being the highest precedence).

1. Default value
2. Environment variable
3. Environment file (.env)
4. Flag

So, for example, if in the above struct you have a default value of `localhost:8080` for *host*, and you provide a flag to your executable, the flag will override the default value. It would even override a value you had set in an environment variable.

### Tags

* **flag** - *Requried*. Defines the flag name to look for on the command line.
* **default** - *Required*. Default value to apply.
* **env** - Defines the name of an environment variable to look for. This applies to both OS environment and *.env* file variables.
* **description** - Flag description. Used when displaying flag options on the command line.

### Supported Data Types

* string
* int
* float64
* bool
* time.Time

### License

Copyright 2022 App Nerds LLC

Permission is hereby granted, free of charge, to any person obtaining a copy of this software and associated documentation files (the "Software"), to deal in the Software without restriction, including without limitation the rights to use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of the Software, and to permit persons to whom the Software is furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.

