package cfp

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"
)

func CreateLinearTSPoints(t *testing.T) (adm.Component, []mondat.TSPoint) {
	comp := adm.Component{
		Name:     "componentname",
		Hostname: "hostname",
		Type:     "responsetime",
		Called:   0,
	}
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")

	tsPoints := make([]mondat.TSPoint, 20, 20)
	for i := 0; i < 20; i++ {
		tsPoints[i] = mondat.TSPoint{
			Component: comp,
			Timestamp: curtime,
			Value:     float64(i),
		}
		curtime = curtime.Add(time.Minute)
	}
	return comp, tsPoints
}
