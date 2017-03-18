package cfp

import (
	"log"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"
)

var cfps map[string]*ARIMAR
var step time.Duration = 5 * time.Minute

type CFPResult struct {
	Component adm.Component
	Timestamp time.Time
	Predtime  time.Time
	FailProb  float64
}

func Predict(monDatCh <-chan mondat.MonDatPoint) <-chan CFPResult {
	var resCh = make(chan CFPResult)
	cfps = make(map[string]*ARIMAR)
	go func() {
		for monDatPoint := range monDatCh {
			comp := monDatPoint.Component
			cfp, ok := cfps[comp.UniqName()]
			if !ok {
				var err error
				cfp, err = New(comp, time.Minute, 5*time.Minute, 1e8)
				if err != nil {
					log.Print(err)
				}
				cfps[comp.UniqName()] = cfp
			}
			//go func(cfp *ARIMAR, monDatPoint mondat.MonDatPoint) {
			cfp.Insert(monDatPoint)
			res, err := cfp.Predict()
			if err != nil {
				log.Print(err)
			}
			resCh <- res
			//}(cfp, monDatPoint)
		}
	}()
	return resCh
}
