package arima

import (
	"log"
	"testing"

	"github.com/teeratpitakrat/hora/model/adm"
)

var testdat = []float64{60, 43, 67, 50, 56, 42, 50, 65, 68, 43, 65, 34, 47, 34, 49, 41, 13, 35, 53, 56}

func TestInsert(t *testing.T) {
	c := adm.Component{"A", "host1"}
	buflen = 20
	a, err := New(c)
	if err != nil {
		t.Error("Error getting new ARIMAR", err)
		return
	}
	a.Insert(1000) // should be dropped
	a.Insert(2000) // should be dropped
	for _, v := range testdat {
		a.Insert(v)
	}
	dat := a.GetData()
	for i, v := range dat {
		if v != testdat[i] {
			t.Errorf("Expected: %f but got %f", testdat[i], v)
		}
	}
}

func TestPredict(t *testing.T) {
	c := adm.Component{"A", "host1"}
	buflen = 20
	a, err := New(c)
	if err != nil {
		t.Error("Error getting new ARIMAR", err)
		return
	}

	a.Insert(1000) // should be dropped
	a.Insert(2000) // should be dropped
	for _, v := range testdat {
		a.Insert(v)
	}
	log.Print(a.Predict(1))
	// TODO: check result
	// {48.55 21.717926902432662 75.38207309756733}
}
