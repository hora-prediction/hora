package predictor

import (
	"log"
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/io"
	"github.com/teeratpitakrat/hora/model/adm"
)

var testdat = []float64{60, 43, 67, 50, 56, 42, 50, 65, 68, 43, 65, 34, 47, 34, 49, 41, 13, 35, 53, 56}

func TestInsert(t *testing.T) {
	c := adm.Component{"A", "host1"}
	buflen = 20
	a, err := New(c, time.Minute, 5*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ARIMAR", err)
		return
	}
	a.Insert(io.MonDatPoint{c, time.Unix(0, 0), 1000}) // should be dropped
	a.Insert(io.MonDatPoint{c, time.Unix(0, 0), 2000}) // should be dropped
	for _, v := range testdat {
		a.Insert(io.MonDatPoint{c, time.Unix(0, 0), v})
	}
	dat := a.GetData()
	for i, v := range dat {
		expected := io.MonDatPoint{c, time.Unix(0, 0), testdat[i]}
		if v != expected {
			t.Errorf("Expected: %f but got %f", expected, v)
		}
	}
}

func TestPredict(t *testing.T) {
	c := adm.Component{"A", "host1"}
	buflen = 20
	a, err := New(c, time.Minute, 5*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ARIMAR", err)
		return
	}

	a.Insert(io.MonDatPoint{c, time.Unix(0, 0), 1000}) // should be dropped
	a.Insert(io.MonDatPoint{c, time.Unix(0, 0), 2000}) // should be dropped
	for _, v := range testdat {
		a.Insert(io.MonDatPoint{c, time.Unix(0, 0), v})
	}
	log.Print(a.Predict())
	// TODO: check result
	// {48.55 21.717926902432662 75.38207309756733}
}
