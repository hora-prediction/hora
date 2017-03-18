package fpm

import (
	"github.com/teeratpitakrat/hora/adm"
)

type FPM interface {
	LoadADM(adm.ADM)
	Create() error
	Update(adm.Component, float64)
	Predict() (map[adm.Component]float64, error)
}
