package cfp

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
)

func TestWriteCfpResult(t *testing.T) {
	writer, err := NewCfpResultWriter("http://localhost:8086", "root", "root")
	if err != nil {
		t.Error(err)
	}
	a := adm.Component{"A", "host1"}
	b := adm.Component{"B", "host2"}
	resulta := Result{
		a,
		time.Now(),
		time.Now().Add(10 * time.Minute),
		0.555,
	}
	resultb := Result{
		b,
		time.Now(),
		time.Now().Add(10 * time.Minute),
		0.666,
	}
	writer.WriteCfpResult(resulta)
	writer.WriteCfpResult(resultb)
}
