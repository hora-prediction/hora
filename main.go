package main

import (
	"log"
	"strings"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"
	"github.com/teeratpitakrat/hora/mondat"
	"github.com/teeratpitakrat/hora/resultio"

	"github.com/spf13/viper"
)

func main() {
	// Read configurations
	log.Println("Reading configuration")
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath(".")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Println("Fatal error config file: %s \n", err)
	}

	viper.SetEnvPrefix("hora") // will be uppercased automatically
	viper.AutomaticEnv()
	viper.SetEnvKeyReplacer(strings.NewReplacer(".", "_"))

	// Read first ADM before continue
	admController := adm.NewController()
	log.Println("Waiting for ADM")
	m := <-admController.AdmCh
	log.Println("Waiting for ADM done")

	// Creating CFPs
	cfpController, cfpResultCh := cfp.NewController(m)

	// Creating FPM
	f, fpmResultCh, err := fpm.NewBayesNetR(m)
	if err != nil {
		log.Println("Error creating FPM", err)
	}

	resultWriter, err := resultio.New(
		viper.GetString("influxdb.addr"),
		viper.GetString("influxdb.username"),
		viper.GetString("influxdb.password"),
	)
	if err != nil {
		log.Println(err)
	}

	go func() {
		for {
			m := <-admController.AdmCh
			log.Println("Updating ADM")
			cfpController.UpdateADM(m)
			f.UpdateAdm(m)
			log.Println("Updating ADM done")
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
		KiekerDb:   viper.GetString("influxdb.db.kieker"),
		K8sDb:      viper.GetString("influxdb.db.k8s"),
		Batch:      viper.GetBool("influxdb.batch"),
		Interval:   viper.GetDuration("prediction.interval"),
	}
	monDatCh := influxReader.Read()

	for {
		monDat, ok := <-monDatCh
		if !ok {
			log.Println("Monitoring data channel closed. Terminating")
			break
		}
		cfpController.AddMonDat(monDat)
		// TODO: send mondat to eval
	}

}
