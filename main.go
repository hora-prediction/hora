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
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	// Read and create adm from file
	var m adm.ADM
	if viper.GetBool("adm.fileio.enabled") {
		admPath := viper.GetString("adm.fileio.path")
		log.Print("Reading adm from ", admPath)
		var err error
		m, err = adm.ReadFile(admPath)
		if err != nil {
			log.Print("Error reading adm", err)
		}
	}

	// Create fpm
	log.Print("Creating fpm")
	f, fpmResultCh, err := fpm.NewBayesNetR(m)
	if err != nil {
		log.Print("Error creating FPM", err)
	}

	resultWriter, err := resultio.New(viper.GetString("influxdb.addr"), viper.GetString("influxdb.username"), viper.GetString("influxdb.password"))
	if err != nil {
		log.Print(err)
	}

	// Read monitoring data
	log.Print("Reading from influxdb")
	reader := mondat.InfluxKiekerReader{
		Archdepmod: m,
		Addr:       viper.GetString("influxdb.addr"),
		Username:   viper.GetString("influxdb.username"),
		Password:   viper.GetString("influxdb.password"),
		Db:         viper.GetString("influxdb.db.kieker"),
		Batch:      viper.GetBool("influxdb.batch"),
		Interval:   viper.GetDuration("prediction.interval"),
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
