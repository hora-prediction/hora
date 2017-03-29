package cfp

import (
	"container/ring"
	"errors"
	"log"
	"strconv"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"
	"github.com/teeratpitakrat/hora/rbridge"

	"github.com/chobie/go-gaussian"
	"github.com/senseyeio/roger"
)

const fillValue float64 = 0.0

type ArimaR struct {
	component adm.Component
	buf       *ring.Ring
	interval  time.Duration
	threshold float64
	leadtime  time.Duration
	rSession  roger.Session
}

func NewArimaR(c adm.Component, interval time.Duration, leadtime time.Duration, history time.Duration, threshold float64) (*ArimaR, error) {
	var a ArimaR
	a.component = c
	buflen := int(history / interval)
	a.buf = ring.New(buflen)
	a.interval = interval
	a.threshold = threshold
	a.leadtime = leadtime
	session, err := rbridge.GetRSession(a.component.UniqName())
	if err != nil {
		log.Print("Error creating new ArimaR predictor: ", err)
		return nil, err
	}
	a.rSession = session
	return &a, nil
}

func (a *ArimaR) Insert(p mondat.TSPoint) {
	if p.Component != a.component {
		log.Printf("Warning: receiving data of another component. Expected %s but got %s. Ignoring data point", a.component.UniqName(), p.Component.UniqName())
		return
	}
	if a.buf.Value != nil {
		lastTSPoint := a.buf.Value.(mondat.TSPoint)
		lastTimestamp := lastTSPoint.Timestamp
		// Drop if received data is older than the latest one
		if p.Timestamp.Equal(lastTimestamp) || p.Timestamp.Before(lastTimestamp) {
			log.Printf("Warning: receiving data older than or as old as the latest one. Latest: %s, Received: %s. Ignoring data point", lastTimestamp, p.Timestamp)
			return
		}
		// Fill missing values if there is a gap between latest point in the buffer and the one just received
		for lastBufTimestamp := a.buf.Value.(mondat.TSPoint).Timestamp; lastBufTimestamp.Add(a.interval).Before(p.Timestamp); lastBufTimestamp = lastBufTimestamp.Add(a.interval) {
			a.buf = a.buf.Next()
			fillTSPoint := mondat.TSPoint{
				Component: p.Component,
				Timestamp: lastBufTimestamp.Add(a.interval),
				Value:     fillValue,
			}
			a.buf.Value = fillTSPoint
		}
		a.buf = a.buf.Next()
	}
	a.buf.Value = p
}

func (a *ArimaR) TSPoints() mondat.TSPoints {
	buflen := a.buf.Len()
	dat := make(mondat.TSPoints, buflen, buflen)
	for i := 0; i < buflen; i++ {
		a.buf = a.buf.Next()
		p := a.buf.Value
		if p == nil {
			continue
		}
		dat[i] = p.(mondat.TSPoint)
	}
	return dat
}

func (a *ArimaR) Predict() (Result, error) {
	var result Result
	// load data
	cmd := "fit <- auto.arima(c("
	for i, v := range a.TSPoints() {
		if i > 0 {
			cmd += ", "
		}
		cmd += strconv.FormatFloat(v.Value, 'f', 6, 64)
	}
	cmd += "))"
	_, err := a.rSession.Eval(cmd)
	if err != nil {
		log.Printf("Cannot evaluate R with cmd=%s\n%s", cmd, err)
		return result, err
	}

	// forecast
	cmd = "forecast(fit, h="
	step := int64(a.leadtime / a.interval)
	cmd += strconv.FormatInt(step, 10)
	cmd += ")"
	ret, err := a.rSession.Eval(cmd)
	if err != nil {
		log.Printf("Cannot evaluate R with cmd=%s\n%s", cmd, err)
		return result, err
	}
	res := ret.(map[string]interface{})

	// parse result
	var mean float64
	if step == 1 { // if step == 1, the returned mean is float64
		mean = res["mean"].(float64)
	} else { // if step > 1, the returned mean is []float64
		meanArray := res["mean"].([]float64)
		mean = meanArray[len(meanArray)-1]
	}
	lowerArray := res["lower"].([]float64)
	lower := lowerArray[len(lowerArray)-1]
	upperArray := res["upper"].([]float64)
	upper := upperArray[len(upperArray)-1]
	sd := (upper - lower) / 3.92

	if sd <= 0 {
		return result, errors.New("Standard deviation <= 0")
	}
	distribution := gaussian.NewGaussian(mean, sd*sd)
	failProb := 1 - distribution.Cdf(a.threshold)

	result = Result{
		a.component,
		a.buf.Value.(mondat.TSPoint).Timestamp,
		a.buf.Value.(mondat.TSPoint).Timestamp.Add(a.leadtime),
		mean,
		lower,
		upper,
		sd,
		failProb,
	}
	return result, nil
}
