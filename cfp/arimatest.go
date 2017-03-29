package cfp

import (
	"math"
	"testing"
	"time"

	"github.com/hora-prediction/hora/adm"
	"github.com/hora-prediction/hora/mondat"
)

func CreateLinearTSPoints(t *testing.T) (adm.Component, []mondat.TSPoint) {
	numPoints := 20
	comp := adm.Component{
		Name:     "componentname",
		Hostname: "hostname",
		Type:     "responsetime",
		Called:   0,
	}
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")

	tsPoints := make([]mondat.TSPoint, numPoints, numPoints)
	for i := 0; i < numPoints; i++ {
		tsPoints[i] = mondat.TSPoint{
			Component: comp,
			Timestamp: curtime,
			Value:     float64(i),
		}
		curtime = curtime.Add(time.Minute)
	}
	return comp, tsPoints
}

func CreateLinearTSPointsWithWrongTimestamp(t *testing.T) (adm.Component, []mondat.TSPoint) {
	numPoints := 22
	comp := adm.Component{
		Name:     "componentname",
		Hostname: "hostname",
		Type:     "responsetime",
		Called:   0,
	}
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")

	tsPoints := make([]mondat.TSPoint, numPoints, numPoints)
	for i := 0; i < numPoints; i++ {
		tsPoints[i] = mondat.TSPoint{
			Component: comp,
			Timestamp: curtime,
			Value:     float64(i),
		}
		curtime = curtime.Add(time.Minute)
	}
	wrongtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:18 UTC")
	tsPoints[20] = mondat.TSPoint{
		Component: comp,
		Timestamp: wrongtime,
		Value:     0,
	}
	tsPoints[21] = mondat.TSPoint{
		Component: comp,
		Timestamp: wrongtime.Add(time.Minute),
		Value:     0,
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

func CreateMissingTSPoints(t *testing.T) (adm.Component, []mondat.TSPoint, []mondat.TSPoint) {
	comp := adm.Component{
		Name:     "componentname",
		Hostname: "hostname",
		Type:     "responsetime",
		Called:   0,
	}
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")

	missingTSPointsValues := []float64{1, 2, 3, 4, 5}
	missingTSPoints := make([]mondat.TSPoint, 5, 5)
	for i := 0; i < 5; i++ {
		missingTSPoints[i] = mondat.TSPoint{
			Component: comp,
			Timestamp: curtime,
			Value:     missingTSPointsValues[i],
		}
		curtime = curtime.Add(5 * time.Minute)
	}

	curtime, _ = time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:01 UTC")
	completeTSPointsValues := []float64{0, 0, 0, 0, 2, 0, 0, 0, 0, 3, 0, 0, 0, 0, 4, 0, 0, 0, 0, 5}
	completeTSPoints := make([]mondat.TSPoint, 20, 20)
	for i := 0; i < 20; i++ {
		completeTSPoints[i] = mondat.TSPoint{
			Component: comp,
			Timestamp: curtime,
			Value:     completeTSPointsValues[i],
		}
		curtime = curtime.Add(time.Minute)
	}
	return comp, missingTSPoints, completeTSPoints
}
