package cfp

import (
	"math"
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

func CreateSeasonalTSPoints(t *testing.T) (adm.Component, []mondat.TSPoint) {
	numPoints := 40
	comp := adm.Component{
		Name:     "componentname",
		Hostname: "hostname",
		Type:     "responsetime",
		Called:   0,
	}
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")

	tsPoints := make([]mondat.TSPoint, numPoints, numPoints)
	for i, pi := 0, math.Pi/10; i < numPoints; i, pi = i+1, pi+math.Pi/10 {
		tsPoints[i] = mondat.TSPoint{
			Component: comp,
			Timestamp: curtime,
			Value:     math.Sin(pi),
		}
		curtime = curtime.Add(time.Minute)
	}
	return comp, tsPoints
}
