package mondat

import (
	"time"

	"github.com/teeratpitakrat/hora/adm"
)

type TSPoint struct {
	Component adm.Component
	Timestamp time.Time
	Value     float64
}

type TSPoints []TSPoint

func (mondat TSPoints) Len() int {
	return len(mondat)
}

func (mondat TSPoints) Less(i, j int) bool {
	return mondat[i].Timestamp.Before(mondat[j].Timestamp)
}

func (mondat TSPoints) Swap(i, j int) {
	mondat[i], mondat[j] = mondat[j], mondat[i]
}
