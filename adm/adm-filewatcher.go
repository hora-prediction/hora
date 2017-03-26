package adm

import (
	"encoding/json"
	"io/ioutil"
	"log"

	"github.com/spf13/viper"
)

type FileWatcher struct {
	m     ADM
	admCh chan ADM
}

func NewFileWatcher() FileWatcher {
	viper.SetDefault("adm.filewatcher.path", "/tmp/adm.json")
	fileWatcher := FileWatcher{
		m:     ADM{},
		admCh: make(chan ADM, 2),
	}

	//filepath := viper.GetString("adm.filewatcher.path")
	return fileWatcher
}

func (r *FileWatcher) UpdateADM(m ADM) {
	r.m = m
}

func ReadFile(path string) (ADM, error) {
	var m ADM
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print("Error reading json file", err)
		return nil, err
	}
	err = json.Unmarshal(b, &m)
	return m, nil
}

func (m *ADM) WriteFile(path string) error {
	b, err := json.Marshal(m)
	if err != nil {
		log.Println(err)
	}

	err = ioutil.WriteFile(path, b, 0644)
	if err != nil {
		log.Println(err)
	}
	return nil
}
