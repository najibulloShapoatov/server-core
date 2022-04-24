# Configuration library

Allows the application to load it's configuration from `.config` files or environment variables

## Install

To install the platform

```
$ go get github.com/najibulloShapoatov/server-core/settings
```

## Usage example

```go
import "github.com/najibulloShapoatov/server-core/settings"

type MySettings struct {
    Host    string        `config:"app.host" default:"localhost"`
    Port    int           `config:"app.port" default:"80"`
    Timeout time.Duration `config:"app.timeout" default:"3s"`
    Debug   bool          `config:"debug" default:"on"`
    
}

func main() {

   s := settings.GetSettings()
   err := s.Load(
         settings.NewFileLoader("file.conf", false),  // load config from this file
         settings.NewEnvLoader(false, "app")          // and also from environment variables
   )
   if err != nil {
       fmt.Prinln("failed to parse settings", err)
   }

   // unmarshall directly into a struct
   var mySettings MySettings
   if err := settings.Unmarshal(&mySettings); err != nil {
       fmt.Println("some error", err)
   }

   // retrieve a value by name
   isDebug, exists := settings.GetBool("debug")
}

```

## Configuration files

Configuration files are mostly `key=value` files but with few additions. For example, numbers are evaluated by the parser and booleans can be all truthy values besides true or false. Other files can be included using the `include` directive

```bash
# String values
key.name = "value" # this is inline comment
key.multiline = "multi \
line \
string"

# Number values
test.int.value = 5
test.float.value = 3.14
test.negative.value = -1.2
test.hex.number = 0x1234 # will parse to 4660
test.octal.number = 0o123 # will parse to 83
test.binary.number = 0b1010101 # will parse to 85
test.exponential.number = 1e3 # will parse to 1000
test.negative.exponential.number = 2e-2 # will parse to 0.02

# Boolean values
test.bool.value1 = yes       # or no
test.bool.value2 = on        # or off
test.bool.value3 = set       # or unset
test.bool.value4 = active    # or inactive
test.bool.value5 = enabled   # or disabled
test.bool.value6 = true      # or false
test.bool.value6 = 1         # or 0

# Duration
test.duration.value1 = "1h5m"
test.duration.value2 = "3s"

# Include other file
include "sub-config-file.conf"
```