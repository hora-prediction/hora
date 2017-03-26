package fpm

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
)

func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	m := adm.CreateSmallADM(t)

	f, fpmResultCh, err := NewBayesNetR(m)
	if err != nil {
		t.Error("Error creating BayesNetR", err)
	}

	cfpResult := cfp.Result{adm.Component{"D", "host4", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.0}
	f.UpdateCfpResult(cfpResult)
	fpmResult := <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	// TODO: more precision checks
	fprobA := fpmResult.FailProbs[adm.Component{"A", "host1", "responsetime", 0}]
	if fprobA != 0 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB := fpmResult.FailProbs[adm.Component{"B", "host2", "responsetime", 0}]
	if fprobB != 0 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC := fpmResult.FailProbs[adm.Component{"C", "host3", "responsetime", 0}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD := fpmResult.FailProbs[adm.Component{"D", "host4", "responsetime", 0}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResult = cfp.Result{adm.Component{"D", "host4", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.1}
	f.UpdateCfpResult(cfpResult)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1", "responsetime", 0}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2", "responsetime", 0}]
	if fprobB > 0.12 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3", "responsetime", 0}]
	if fprobC > 0.12 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4", "responsetime", 0}]
	if fprobD > 0.12 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResult = cfp.Result{adm.Component{"D", "host4", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.9}
	f.UpdateCfpResult(cfpResult)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1", "responsetime", 0}]
	if fprobA < 0.89 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2", "responsetime", 0}]
	if fprobB < 0.89 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3", "responsetime", 0}]
	if fprobC < 0.89 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4", "responsetime", 0}]
	if fprobD < 0.89 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResultD := cfp.Result{adm.Component{"D", "host4", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.0}
	f.UpdateCfpResult(cfpResultD)
	fpmResult = <-fpmResultCh
	cfpResultB := cfp.Result{adm.Component{"B", "host2", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.1}
	f.UpdateCfpResult(cfpResultB)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1", "responsetime", 0}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2", "responsetime", 0}]
	if fprobB > 0.12 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3", "responsetime", 0}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4", "responsetime", 0}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResultB = cfp.Result{adm.Component{"B", "host2", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.0}
	cfpResultA := cfp.Result{adm.Component{"A", "host1", "responsetime", 0}, time.Unix(0, 0), time.Unix(0, 300), 0.1}
	f.UpdateCfpResult(cfpResultB)
	fpmResult = <-fpmResultCh
	f.UpdateCfpResult(cfpResultA)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1", "responsetime", 0}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2", "responsetime", 0}]
	if fprobB > 0.1 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3", "responsetime", 0}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4", "responsetime", 0}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}
}
