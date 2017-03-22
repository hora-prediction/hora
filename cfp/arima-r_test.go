package cfp

import (
	"log"
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"
)

// TODO: build mondat.TSPoint
var testdat = []float64{60, 43, 67, 50, 56, 42, 50, 65, 68, 43, 65, 34, 47, 34, 49, 41, 13, 35, 53, 56}

func TestInsert(t *testing.T) {
	c := adm.Component{"A", "host1", "responsetime", 0}
	a, err := NewArimaR(c, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 1000}) // should be dropped
	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 2000}) // should be dropped
	for _, v := range testdat {
		a.Insert(mondat.TSPoint{c, time.Unix(0, 0), v})
	}
	dat := a.TSPoints()
	for i, v := range dat {
		expected := mondat.TSPoint{c, time.Unix(0, 0), testdat[i]}
		if v != expected {
			t.Errorf("Expected: %f but got %f", expected, v)
		}
	}
}

func TestPredict(t *testing.T) {
	c := adm.Component{"A", "host1", "responsetime", 0}
	a, err := NewArimaR(c, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}

	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 1000}) // should be dropped
	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 2000}) // should be dropped
	for _, v := range testdat {
		a.Insert(mondat.TSPoint{c, time.Unix(0, 0), v})
	}
	log.Print(a.Predict())
	// TODO: check result
	// {48.55 21.717926902432662 75.38207309756733}
}
