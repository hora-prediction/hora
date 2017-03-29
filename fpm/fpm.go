package fpm

import (
	"time"

	"github.com/hora-prediction/hora/adm"
	"github.com/hora-prediction/hora/cfp"
)

type Fpm interface {
	UpdateAdm(adm.ADM)
	UpdateCfpResult(cfpResult cfp.Result)
}

type Result struct {
	FailProbs map[adm.Component]float64
	Timestamp time.Time
	Predtime  time.Time
}
