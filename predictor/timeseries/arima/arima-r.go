package arima

import (
	"container/ring"
	"log"
	"strconv"

	"github.com/senseyeio/roger"

	"github.com/teeratpitakrat/hora/model/adm"
	"github.com/teeratpitakrat/hora/rbridge"
)

var buflen = 20

type ARIMAR struct {
	component adm.Component
	buf       *ring.Ring
	rSession  roger.Session
}

type Result struct {
	mean  float64
	lower float64
	upper float64
}

func New(c adm.Component) (*ARIMAR, error) {
	var a ARIMAR
	a.component = c
	a.buf = ring.New(buflen)
	session, err := rbridge.GetRSession(a.component.UniqName())
	if err != nil {
		log.Print("Error creating new ARIMAR predictor: ", err)
		return nil, err
	}
	a.rSession = session
	return &a, nil
}

func (a *ARIMAR) Insert(p float64) {
	if a.buf == nil {
		a.buf = ring.New(buflen)
	}
	a.buf = a.buf.Next()
	a.buf.Value = p
}

func (a *ARIMAR) GetData() []float64 {
	dat := make([]float64, buflen, buflen)
	for i := 0; i < buflen; i++ {
		a.buf = a.buf.Next()
		v := a.buf.Value
		if v == nil {
			continue
		}
		dat[i] = v.(float64)
	}
	return dat
}

func (a *ARIMAR) Predict(step int64) (*Result, error) {
	// load data
	cmd := "fit <- auto.arima(c("
	for i, v := range a.GetData() {
		if i > 0 {
			cmd += ", "
		}
		cmd += strconv.FormatFloat(v, 'f', 6, 64)
	}
	cmd += "))"
	_, err := a.rSession.Eval(cmd)
	if err != nil {
		log.Print("Error: ", err)
		return nil, err
	}

	// forecast
	cmd = "forecast(fit, h="
	cmd += strconv.FormatInt(step, 10)
	cmd += ")"
	ret, err := a.rSession.Eval(cmd)
	if err != nil {
		log.Print("Error: ", err)
		return nil, err
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

	result := &Result{mean, lower, upper}
	return result, nil
}
