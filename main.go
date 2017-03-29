package main

import (
	"log"
	"strings"
	"time"

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
		viper.GetString("influxdb.hora.addr"),
		viper.GetString("influxdb.hora.username"),
		viper.GetString("influxdb.hora.password"),
		viper.GetString("influxdb.hora.db"),
	)
	if err != nil {
		log.Println(err)
	}

	influxKiekerReader := mondat.InfluxKiekerReader{
		Archdepmod: m,
		KiekerDb: mondat.InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: viper.GetString("influxdb.kieker.username"),
			Password: viper.GetString("influxdb.kieker.password"),
			DbName:   viper.GetString("influxdb.kieker.db"),
		},
		K8sDb: mondat.InfluxDBConfig{
			Addr:     viper.GetString("influxdb.k8s.addr"),
			Username: viper.GetString("influxdb.k8s.username"),
			Password: viper.GetString("influxdb.k8s.password"),
			DbName:   viper.GetString("influxdb.k8s.db"),
		},
		Batch:    false,
		Interval: time.Minute,
	}

	go func() {
		for {
			m := <-admController.AdmCh
			log.Println("Updating ADM")
			influxKiekerReader.UpdateADM(m)
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
	monDatCh := influxKiekerReader.Read()
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
