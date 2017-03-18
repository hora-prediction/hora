package main

import (
	"log"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"
	"github.com/teeratpitakrat/hora/mondat"
)

func main() {

	// Read and create adm from file
	log.Print("reading adm")
	m, err := adm.Import("/tmp/adm.json")
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
	reader := mondat.InfluxMonDatReader{
		Archdepmod: m,
		Addr:       "http://localhost:8086",
		Username:   "root",
		Password:   "root",
		Db:         "kieker",
		Batch:      true,
	}
	monDatCh := reader.Read()
	log.Print("starting cfp")
	cfpCh := cfp.Predict(monDatCh)
	//for {
	//_, ok := <-monDatCh
	//if ok {
	////log.Print(d)
	//} else {
	//break
	//}
	//}

	for cfpres := range cfpCh {
		f.Update(cfpres.Component, cfpres.FailProb)
		time.Sleep(100 * time.Millisecond)
		res, err := f.Predict()
		if err != nil {
			log.Print("Error making prediction", err)
		}
		log.Print("fpmres=", res)
	}

	// update prob
	// update fpm (with delay)
	//f.Update(adm.Component{"public void com.netflix.recipes.rss.manager.RSSManager.deleteSubscription(java.lang.String, java.lang.String)", "middletier-6d65k"}, 0.9)
	//time.Sleep(time.Second)
	//res, err = f.Predict()
	//if err != nil {
	//log.Print("Error making prediction", err)
	//}
	//log.Print(res)

	// go routine
	// read new data from channel
	// make prediction for that component
	// push result to influxdb
}
