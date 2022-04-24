package settings

import (
	"fmt"
	"time"
)

// fileLoader loads and processes .conf files
type fileLoader struct {
	watchForChanges bool
	filename        string
	stop            chan bool
	watchInterval   time.Duration
}

// NewFileLoader returns a .conf file parser
func NewFileLoader(filename string, watchForChanges bool) Loader {
	loader := &fileLoader{
		watchForChanges: watchForChanges,
		filename:        filename,
		stop:            make(chan bool),
		watchInterval:   time.Second * 3,
	}
	if watchForChanges {
		go loader.watch()
	}
	return loader
}

// Parse the file and returns a map of key=value settings
func (f *fileLoader) Parse() (map[string]string, error) {
	return f.parseFile(f.filename)
}

func (f *fileLoader) parseFile(filename string) (map[string]string, error) {
	strLoader := &stringLoader{filename: filename}
	return strLoader.parseData()
}

func (f *fileLoader) watch() {
	timer := time.NewTimer(f.watchInterval)
	for {
		select {
		case <-timer.C:
			// not implemented
			fmt.Println("checking")
		case <-f.stop:
			timer.Stop()
			return
		}
	}
}
