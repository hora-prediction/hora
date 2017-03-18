package mondat

import (
	"time"

	"github.com/teeratpitakrat/hora/adm"
)

type MonDatPoint struct {
	Component adm.Component
	Timestamp time.Time
	Value     float64
}

type MonDat []MonDatPoint

func (mondat MonDat) Len() int {
	return len(mondat)
}

func (mondat MonDat) Less(i, j int) bool {
	return mondat[i].Timestamp.Before(mondat[j].Timestamp)
}

func (mondat MonDat) Swap(i, j int) {
	mondat[i], mondat[j] = mondat[j], mondat[i]
}
