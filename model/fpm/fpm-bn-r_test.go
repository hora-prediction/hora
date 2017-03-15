package fpm

import (
	"testing"

	"github.com/teeratpitakrat/hora/model/adm"
	"github.com/teeratpitakrat/hora/rbridge"
)

func TestCreate(t *testing.T) {
	//archmodel := make(adm.ADM)

	//depA := adm.DepList{make([]adm.Dep, 2, 2)}
	//depA.Deps[0] = adm.Dep{adm.Component{"B", "host2"}, 0.5}
	//depA.Deps[1] = adm.Dep{adm.Component{"C", "host3"}, 0.5}
	//archmodel[adm.Component{"A", "host1"}] = depA

	//depC := adm.DepList{make([]adm.Dep, 1, 1)}
	//depC.Deps[0] = adm.Dep{adm.Component{"D", "host4"}, 1}
	//archmodel[adm.Component{"C", "host3"}] = depC

	//depB := adm.DepList{make([]adm.Dep, 1, 1)}
	//depB.Deps[0] = adm.Dep{adm.Component{"D", "host4"}, 1}
	//archmodel[adm.Component{"B", "host2"}] = depB

	//depD := adm.DepList{}
	//archmodel[adm.Component{"D", "host4"}] = depD

	m := make(adm.ADM)

	compA := adm.Component{"A", "host1"}
	compB := adm.Component{"B", "host2"}
	compC := adm.Component{"C", "host3"}
	compD := adm.Component{"D", "host4"}

	depA := adm.DepList{compA, make([]adm.Dep, 2, 2)}
	depA.Component = compA
	depA.Deps[0] = adm.Dep{compB, 0.5}
	depA.Deps[1] = adm.Dep{compC, 0.5}
	m[compA.UniqName()] = depA

	depB := adm.DepList{compB, make([]adm.Dep, 1, 1)}
	depB.Component = compB
	depB.Deps[0] = adm.Dep{compD, 1}
	m[compB.UniqName()] = depB

	depC := adm.DepList{compC, make([]adm.Dep, 1, 1)}
	depC.Component = compC
	depC.Deps[0] = adm.Dep{compD, 1}
	m[compC.UniqName()] = depC

	depD := adm.DepList{}
	depD.Component = compD
	m[compD.UniqName()] = depD

	// Configure R bridge
	rbridge.SetHostname("localhost")
	rbridge.SetPort(6311)

	var f FPMBNR
	f.LoadADM(m)
	err := f.Create()
	if err != nil {
		t.Error("Error creating FPM", err)
	}

	res, err := f.Predict()
	if err != nil {
		t.Error("Error making prediction", err)
	}
	// TODO: more precision checks
	fprobA := res[adm.Component{"A", "host1"}]
	if fprobA != 0 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB := res[adm.Component{"B", "host2"}]
	if fprobB != 0 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC := res[adm.Component{"C", "host3"}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD := res[adm.Component{"D", "host4"}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}

	f.Update(adm.Component{"D", "host4"}, 0.1)
	res, err = f.Predict()
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = res[adm.Component{"A", "host1"}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = res[adm.Component{"B", "host2"}]
	if fprobB > 0.12 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = res[adm.Component{"C", "host3"}]
	if fprobC > 0.12 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = res[adm.Component{"D", "host4"}]
	if fprobD > 0.12 {
		t.Error("Expected: 0 but got", fprobD)
	}

	f.Update(adm.Component{"D", "host4"}, 0.9)
	res, err = f.Predict()
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = res[adm.Component{"A", "host1"}]
	if fprobA < 0.89 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = res[adm.Component{"B", "host2"}]
	if fprobB < 0.89 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = res[adm.Component{"C", "host3"}]
	if fprobC < 0.89 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = res[adm.Component{"D", "host4"}]
	if fprobD < 0.89 {
		t.Error("Expected: 0 but got", fprobD)
	}

	f.Update(adm.Component{"D", "host4"}, 0.0)
	f.Update(adm.Component{"B", "host2"}, 0.1)
	res, err = f.Predict()
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = res[adm.Component{"A", "host1"}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = res[adm.Component{"B", "host2"}]
	if fprobB > 0.12 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = res[adm.Component{"C", "host3"}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = res[adm.Component{"D", "host4"}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}

	f.Update(adm.Component{"B", "host2"}, 0.0)
	f.Update(adm.Component{"A", "host1"}, 0.1)
	res, err = f.Predict()
	if err != nil {
		t.Error("Error making prediction", err)
	}
	fprobA = res[adm.Component{"A", "host1"}]
	if fprobA > 0.12 {
		t.Error("Expected: 0 but got", fprobA)
	}
	fprobB = res[adm.Component{"B", "host2"}]
	if fprobB != 0 {
		t.Error("Expected: 0 but got", fprobB)
	}
	fprobC = res[adm.Component{"C", "host3"}]
	if fprobC != 0 {
		t.Error("Expected: 0 but got", fprobC)
	}
	fprobD = res[adm.Component{"D", "host4"}]
	if fprobD != 0 {
		t.Error("Expected: 0 but got", fprobD)
	}
}
