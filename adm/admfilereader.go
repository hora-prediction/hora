package adm

import (
	"encoding/json"
	"io/ioutil"
	"log"
)

func Import(path string) (ADM, error) {
	var m ADM
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print("Error reading json file", err)
		return nil, err
	}
	err = json.Unmarshal(b, &m)
	return m, nil
}

func (m *ADM) Export(path string) error {
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
