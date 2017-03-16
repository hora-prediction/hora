package main

import (
	"log"

	"github.com/teeratpitakrat/hora/io"
	"github.com/teeratpitakrat/hora/model/fpm"
)

func main() {

	// Read and create adm from file
	log.Print("reading adm")
	m, err := io.Import("/tmp/adm.json")
	if err != nil {
		log.Print("Error reading adm", err)
	}

	// Create fpm
	log.Print("creating fpm")
	var f fpm.FPMBNR
	f.LoadADM(m)
	log.Print("loaded fpm")
	err = f.Create()
	if err != nil {
		log.Print("Error creating FPM", err)
	}

	// start reading new data from influxdb every 1 min and push to channel mondatch
	log.Print("reading influxdb")
	reader := io.InfluxMonDatReader{
		Archdepmod: m,
		Addr:       "http://localhost:8086",
		Username:   "root",
		Password:   "root",
		Db:         "kieker",
		Batch:      true,
	}
	ch := reader.Read()
	for {
		d, ok := <-ch
		if ok {
			log.Print(d)
		} else {
			break
		}
	}

	// go routine
	// read new data from channel
	// make prediction for that component
	// update prob
	// update fpm (with delay)
	// push result to influxdb
}
