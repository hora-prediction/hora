package cfp

import (
	"log"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"
)

var cfps map[string]Cfp
var step time.Duration = 5 * time.Minute
var threshold float64 = 1e12

type Cfp interface {
	Insert(mondat.TSPoint)
	TSPoints() mondat.TSPoints
	Predict() (CfpResult, error)
}

type CfpResult struct {
	Component adm.Component
	Timestamp time.Time
	Predtime  time.Time
	FailProb  float64
}

func Predict(monCh <-chan mondat.TSPoint) <-chan CfpResult {
	var cfpResultCh = make(chan CfpResult)
	cfps = make(map[string]Cfp)
	go func() {
		for tsPoint := range monCh {
			comp := tsPoint.Component
			cfp, ok := cfps[comp.UniqName()]
			if !ok {
				var err error
				cfp, err = NewArimaR(comp, time.Minute, 5*time.Minute, threshold)
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
	}()
	return cfpResultCh
}
