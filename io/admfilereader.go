package io

import (
	"encoding/json"
	"io/ioutil"
	"log"
	//"os"

	"github.com/teeratpitakrat/hora/model/adm"
)

func Import(path string) (adm.ADM, error) {
	var m adm.ADM
	b, err := ioutil.ReadFile(path)
	if err != nil {
		log.Print("Error reading json file", err)
		return nil, err
	}
	err = json.Unmarshal(b, &m)
	return m, nil
}

func Export(m adm.ADM, path string) error {
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
