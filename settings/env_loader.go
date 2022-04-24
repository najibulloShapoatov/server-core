package settings

import (
	"fmt"
	"os"
	"strings"
	"time"
)

// envLoader loads and processes
type envLoader struct {
	watchForChanges bool
	stop            chan bool
	watchInterval   time.Duration
	prefix          string
}

// NewEnvLoader returns a env vars parser
func NewEnvLoader(watchForChanges bool, prefix string) Loader {
	loader := &envLoader{
		watchForChanges: watchForChanges,
		prefix:          prefix,
		stop:            make(chan bool),
		watchInterval:   time.Second * 3,
	}
	if watchForChanges {
		go loader.watch()
	}
	return loader
}

func (f envLoader) Parse() (map[string]string, error) {
	// Collect the environment variable
	// If prefix is provided returning only variables with prefix
	var res = map[string]string{}
	if f.prefix == "" {
		for _, v := range os.Environ() {
			kv := strings.SplitN(v, "=", 2)
			res[kv[0]] = kv[1]
		}
	} else {
		for _, v := range os.Environ() {
			if strings.HasPrefix(strings.ToLower(v), strings.ToLower(f.prefix)) {
				kv := strings.SplitN(v, "=", 2)
				res[kv[0]] = kv[1]
			}
		}
	}

	return res, nil
}

func (f *envLoader) watch() {
	timer := time.NewTimer(f.watchInterval)
	for {
		select {
		case <-timer.C:
			// parse again , notify is diffs (use diff)
			fmt.Println("checking")
		case <-f.stop:
			timer.Stop()
			return
		}
	}
}
