package main

import (
	"log"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"
	"github.com/teeratpitakrat/hora/mondat"
	"github.com/teeratpitakrat/hora/resultio"
)

func main() {

	// Read and create adm from file
	log.Print("reading adm")
	m, err := adm.ReadFile("/tmp/adm.json")
	if err != nil {
		log.Print("Error reading adm", err)
	}

	// Create fpm
	log.Print("creating fpm")
	f, fpmResultCh, err := fpm.NewBayesNetR(m)
	if err != nil {
		log.Print("Error creating FPM", err)
	}

	resultWriter, err := resultio.New("http://localhost:8086", "root", "root")
	if err != nil {
		log.Print(err)
	}

	// start reading new data from influxdb every 1 min and push to channel mondatch
	log.Print("reading influxdb")
	reader := mondat.InfluxKiekerReader{
		Archdepmod: m,
		Addr:       "http://localhost:8086",
		Username:   "root",
		Password:   "root",
		Db:         "kieker",
		Batch:      true,
		Interval:   time.Minute,
	}
	monDatCh := reader.Read()

	log.Print("starting cfp")
	cfpResultCh := cfp.Predict(monDatCh)

	for cfpResult := range cfpResultCh {
		resultWriter.WriteCfpResult(cfpResult)
		f.UpdateCfpResult(cfpResult)
		fpmResult := <-fpmResultCh
		resultWriter.WriteFpmResult(fpmResult)
	}
}
