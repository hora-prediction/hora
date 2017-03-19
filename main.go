package main

import (
	"log"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"
	"github.com/teeratpitakrat/hora/mondat"
	"github.com/teeratpitakrat/hora/resultio"

	"github.com/spf13/viper"
)

func main() {

	// Read configurations
	log.Print("Reading configuration")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	// Read adm before continue
	admCh := adm.NewReader()
	log.Print("Reading ADM")
	m := <-admCh
	log.Print("Reading ADM done")

	// Creating CFPs
	cfpController, cfpResultCh := cfp.NewController(m)

	// Creating FPM
	f, fpmResultCh, err := fpm.NewBayesNetR(m)
	if err != nil {
		log.Print("Error creating FPM", err)
	}

	resultWriter, err := resultio.New(
		viper.GetString("influxdb.addr"),
		viper.GetString("influxdb.username"),
		viper.GetString("influxdb.password"),
	)
	if err != nil {
		log.Print(err)
	}

	go func() {
		for {
			m := <-admCh
			log.Print("Updating ADM")
			cfpController.UpdateADM(m)
			f.UpdateAdm(m)
			log.Print("Updating ADM done")
		}
	}()

	go func() {
		for {
			cfpResult := <-cfpResultCh
			f.UpdateCfpResult(cfpResult)
			resultWriter.WriteCfpResult(cfpResult)
		}
	}()

	go func() {
		for {
			fpmResult := <-fpmResultCh
			resultWriter.WriteFpmResult(fpmResult)
		}
	}()

	// Start reading monitoring data
	influxReader := mondat.InfluxKiekerReader{
		Archdepmod: m,
		Addr:       viper.GetString("influxdb.addr"),
		Username:   viper.GetString("influxdb.username"),
		Password:   viper.GetString("influxdb.password"),
		Db:         viper.GetString("influxdb.db.kieker"),
		Batch:      viper.GetBool("influxdb.batch"),
		Interval:   viper.GetDuration("prediction.interval"),
	}
	monDatCh := influxReader.Read()

	for {
		monDat, ok := <-monDatCh
		if !ok {
			log.Print("Monitoring data channel closed. Terminating")
			break
		}
		cfpController.AddMonDat(monDat)
		// TODO: send mondat to eval
	}

}
