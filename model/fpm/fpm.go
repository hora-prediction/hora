package fpm

import (
	"github.com/teeratpitakrat/hora/model/adm"
)

type FPM interface {
	Create(adm.ADM)
	Predict() map[adm.Component]float64
}
