package fpm

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/rbridge"
)

func TestCreate(t *testing.T) {
	m := make(adm.ADM)

	compA := adm.Component{"A", "host1"}
	compB := adm.Component{"B", "host2"}
	compC := adm.Component{"C", "host3"}
	compD := adm.Component{"D", "host4"}

	depA := adm.DependencyInfo{compA, make([]adm.Dependency, 2, 2)}
	depA.Component = compA
	depA.Dependencies[0] = adm.Dependency{compB, 0.5}
	depA.Dependencies[1] = adm.Dependency{compC, 0.5}
	m[compA.UniqName()] = depA

	depB := adm.DependencyInfo{compB, make([]adm.Dependency, 1, 1)}
	depB.Component = compB
	depB.Dependencies[0] = adm.Dependency{compD, 1}
	m[compB.UniqName()] = depB

	depC := adm.DependencyInfo{compC, make([]adm.Dependency, 1, 1)}
	depC.Component = compC
	depC.Dependencies[0] = adm.Dependency{compD, 1}
	m[compC.UniqName()] = depC

	depD := adm.DependencyInfo{}
	depD.Component = compD
	m[compD.UniqName()] = depD

	// Configure R bridge
	rbridge.SetHostname("localhost")
	rbridge.SetPort(6311)

	f, fpmResultCh, err := NewBayesNetR(m)
	if err != nil {
		t.Error("Error creating BayesNetR", err)
	}

	cfpResult := cfp.Result{adm.Component{"D", "host4"}, time.Unix(0, 0), time.Unix(0, 300), 0.0}
	f.UpdateCfpResult(cfpResult)
	fpmResult := <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	// TODO: more precision checks
	fprobA := fpmResult.FailProbs[adm.Component{"A", "host1"}]
	if fprobA != 0 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB := fpmResult.FailProbs[adm.Component{"B", "host2"}]
	if fprobB != 0 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC := fpmResult.FailProbs[adm.Component{"C", "host3"}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD := fpmResult.FailProbs[adm.Component{"D", "host4"}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResult = cfp.Result{adm.Component{"D", "host4"}, time.Unix(0, 0), time.Unix(0, 300), 0.1}
	f.UpdateCfpResult(cfpResult)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1"}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2"}]
	if fprobB > 0.12 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3"}]
	if fprobC > 0.12 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4"}]
	if fprobD > 0.12 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResult = cfp.Result{adm.Component{"D", "host4"}, time.Unix(0, 0), time.Unix(0, 300), 0.9}
	f.UpdateCfpResult(cfpResult)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1"}]
	if fprobA < 0.89 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2"}]
	if fprobB < 0.89 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3"}]
	if fprobC < 0.89 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4"}]
	if fprobD < 0.89 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResultD := cfp.Result{adm.Component{"D", "host4"}, time.Unix(0, 0), time.Unix(0, 300), 0.0}
	f.UpdateCfpResult(cfpResultD)
	fpmResult = <-fpmResultCh
	cfpResultB := cfp.Result{adm.Component{"B", "host2"}, time.Unix(0, 0), time.Unix(0, 300), 0.1}
	f.UpdateCfpResult(cfpResultB)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1"}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2"}]
	if fprobB > 0.12 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3"}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4"}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}

	cfpResultB = cfp.Result{adm.Component{"B", "host2"}, time.Unix(0, 0), time.Unix(0, 300), 0.0}
	cfpResultA := cfp.Result{adm.Component{"A", "host1"}, time.Unix(0, 0), time.Unix(0, 300), 0.1}
	f.UpdateCfpResult(cfpResultB)
	fpmResult = <-fpmResultCh
	f.UpdateCfpResult(cfpResultA)
	fpmResult = <-fpmResultCh
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = fpmResult.FailProbs[adm.Component{"A", "host1"}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = fpmResult.FailProbs[adm.Component{"B", "host2"}]
	if fprobB > 0.1 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = fpmResult.FailProbs[adm.Component{"C", "host3"}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = fpmResult.FailProbs[adm.Component{"D", "host4"}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}
}
