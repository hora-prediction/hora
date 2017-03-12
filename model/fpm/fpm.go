package fpm

import (
	"github.com/teeratpitakrat/hora/model/adm"
)

type FPM interface {
	Create(adm.ADM)
	Update(adm.ADM)
	Predict() map[string]float64
}
