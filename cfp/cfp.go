package cfp

import (
	"log"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"

	"github.com/spf13/viper"
)

var cfps map[string]Cfp

type Cfp interface {
	Insert(mondat.TSPoint)
	TSPoints() mondat.TSPoints
	Predict() (Result, error)
}

type Result struct {
	Component adm.Component
	Timestamp time.Time
	Predtime  time.Time
	FailProb  float64
}

func Predict(monCh <-chan mondat.TSPoint) <-chan Result {
	var cfpResultCh = make(chan Result)
	cfps = make(map[string]Cfp)
	go func() {
		for tsPoint := range monCh {
			comp := tsPoint.Component
			cfp, ok := cfps[comp.UniqName()]
			if !ok {
				var err error
				// TODO: choose predictor based on component type
				interval := viper.GetDuration("prediction.interval")
				leadtime := viper.GetDuration("prediction.leadtime")
				history := viper.GetDuration("cfp.responsetime.history")
				threshold := float64(viper.GetDuration("cfp.responsetime.threshold") / viper.GetDuration("cfp.responsetime.unit"))
				log.Print("threshold=", threshold)
				cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
				if err != nil {
					log.Print(err)
				}
				cfps[comp.UniqName()] = cfp
			}
			cfp.Insert(tsPoint)
			res, err := cfp.Predict()
			if err != nil {
				log.Print(err)
			}
			cfpResultCh <- res
		}
		close(cfpResultCh)
	}()
	return cfpResultCh
}
