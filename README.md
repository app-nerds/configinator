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

