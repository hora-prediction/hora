package io

import (
	"time"

	"github.com/teeratpitakrat/hora/model/adm"
)

type MonDataPoint struct {
	Component adm.Component
	Timestamp time.Time
	Value     float64
}

type MonData []MonDataPoint

func (mondat MonData) Len() int {
	return len(mondat)
}

func (mondat MonData) Less(i, j int) bool {
	return mondat[i].Timestamp.Before(mondat[j].Timestamp)
}

func (mondat MonData) Swap(i, j int) {
	mondat[i], mondat[j] = mondat[j], mondat[i]
}
