package resultio

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"
)

func TestWriteCfpResult(t *testing.T) {
	writer, err := New("http://localhost:8086", "root", "root")
	if err != nil {
		t.Error(err)
	}
	a := adm.Component{"A", "host1"}
	b := adm.Component{"B", "host2"}
	resulta := cfp.Result{
		a,
		time.Now(),
		time.Now().Add(10 * time.Minute),
		0.555,
	}
	resultb := cfp.Result{
		b,
		time.Now(),
		time.Now().Add(10 * time.Minute),
		0.666,
	}
	writer.WriteCfpResult(resulta)
	writer.WriteCfpResult(resultb)
}

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

	result := fpm.Result{
		failProbs,
		time.Now(),
		time.Now().Add(10 * time.Minute),
	}
	writer.WriteFpmResult(result)
}
