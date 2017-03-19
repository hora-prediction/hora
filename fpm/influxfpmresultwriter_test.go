package fpm

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
)

func TestWriteFpmResult(t *testing.T) {
	writer, err := New("http://localhost:8086", "root", "root")
	if err != nil {
		t.Error(err)
	}

	a := adm.Component{"A", "host1"}
	b := adm.Component{"B", "host2"}
	failProbs := make(map[adm.Component]float64)
	failProbs[a] = 0.2
	failProbs[b] = 0.3

	result := Result{
		failProbs,
		time.Now(),
		time.Now().Add(10 * time.Minute),
	}
	writer.WriteFpmResult(result)
}
