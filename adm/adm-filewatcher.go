package adm

import (
	"encoding/json"
	"io/ioutil"
	"log"
	"os"

	"github.com/spf13/viper"
)

type FileWatcher struct {
	filepath string
	m        ADM
	admCh    chan ADM
}

func NewFileWatcher() FileWatcher {
	viper.SetDefault("adm.filewatcher.path", os.TempDir()+"/adm.json") // TODO: use OS-specific path separator
	filepath := viper.GetString("adm.filewatcher.path")
	fileWatcher := FileWatcher{
		filepath: filepath,
		m:        ADM{},
		admCh:    make(chan ADM, 2),
	}

	return fileWatcher
}

func (w *FileWatcher) UpdateADM(m ADM) {
	w.m = m
	w.Write()
}

func (w *FileWatcher) Start() {
	// TODO: use fsnotify
	w.Read()
	w.admCh <- w.m
}

func (w *FileWatcher) Read() {
	b, err := ioutil.ReadFile(w.filepath)
	if err != nil {
		log.Printf("Error reading json file: %s", err)
	}
	err = json.Unmarshal(b, &w.m)
	if err != nil {
		log.Printf("Error parsing json: %s", err)
	}
}

func (w *FileWatcher) Write() {
	b, err := json.MarshalIndent(w.m, "", "  ")
	if err != nil {
		log.Printf("Error marshalling ADM: %s", err)
	}

	err = ioutil.WriteFile(w.filepath, b, 0644)
	if err != nil {
		log.Printf("Error writing ADM to file: %s", err)
	}
}
