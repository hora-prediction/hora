package main

import (
	"log"
	"strings"
	"time"

	"github.com/hora-prediction/hora/adm"
	"github.com/hora-prediction/hora/cfp"
	"github.com/hora-prediction/hora/eval"
	"github.com/hora-prediction/hora/fpm"
	"github.com/hora-prediction/hora/mondat"
	"github.com/hora-prediction/hora/resultio"

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

	batchMode := viper.GetBool("prediction.batch")

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
		log.Printf("Error creating InfluxDB result writer. %s", err)
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
		Batch:     batchMode,
		Interval:  time.Minute,
		Starttime: viper.GetTime("prediction.starttime"),
		Endtime:   viper.GetTime("prediction.endtime"),
	}

	evaluator := eval.New()

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
			evaluator.UpdateCfpResult(cfpResult)
			resultWriter.WriteCfpResult(cfpResult)
		}
	}()

	go func() {
		for {
			fpmResult := <-fpmResultCh
			evaluator.UpdateFpmResult(fpmResult)
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
		evaluator.UpdateMondat(monDat)
	}
	if batchMode {
		log.Println("Running evaluation")
		evaluator.ComputeROC()
	}
}
