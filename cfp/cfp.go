package cfp

import (
	"log"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"

	"github.com/spf13/viper"
)

type CfpController struct {
	cfps        map[string]Cfp
	m           adm.ADM
	monCh       chan mondat.TSPoint
	admCh       chan adm.ADM
	cfpResultCh chan Result
}

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

func NewController(model adm.ADM) (*CfpController, <-chan Result) {
	c := CfpController{
		cfps:        make(map[string]Cfp),
		m:           model,
		monCh:       make(chan mondat.TSPoint, 10),
		admCh:       make(chan adm.ADM, 1),
		cfpResultCh: make(chan Result, 10),
	}
	c.start()
	return &c, c.cfpResultCh
}

func (c *CfpController) AddMonDat(d mondat.TSPoint) {
	c.monCh <- d
}

func (c *CfpController) UpdateADM(m adm.ADM) {
	c.admCh <- m
}

func (c *CfpController) start() {
	log.Print("Starting CfpController")
	go func() {
		for {
			select {
			case tsPoint, ok := <-c.monCh:
				comp := tsPoint.Component
				cfp, ok := c.cfps[comp.UniqName()]
				if !ok {
					var err error
					// TODO: choose predictor based on component type
					interval := viper.GetDuration("prediction.interval")
					leadtime := viper.GetDuration("prediction.leadtime")
					history := viper.GetDuration("cfp.responsetime.history")
					threshold := float64(viper.GetDuration("cfp.responsetime.threshold") / viper.GetDuration("cfp.responsetime.unit"))
					cfp, err = NewArimaR(comp, interval, leadtime, history, threshold)
					if err != nil {
						log.Print(err)
					}
					c.cfps[comp.UniqName()] = cfp
				}
				cfp.Insert(tsPoint)
				res, err := cfp.Predict()
				if err != nil {
					log.Print(err)
				}
				c.cfpResultCh <- res
			case model, _ := <-c.admCh:
				c.m = model
			}
		}
		close(c.cfpResultCh)
	}()
}
