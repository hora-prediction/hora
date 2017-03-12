package fpm

import (
	"testing"

	"github.com/teeratpitakrat/hora/model/adm"
)

func TestCreate(t *testing.T) {
	archmodel := make(adm.ADM)

	depA := adm.DepList{make([]adm.Dep, 2, 2)}
	depA.Deps[0] = adm.Dep{adm.Component{"B", "host2"}, 0.5}
	depA.Deps[1] = adm.Dep{adm.Component{"C", "host3"}, 0.5}
	archmodel[adm.Component{"A", "host1"}] = depA

	depB := adm.DepList{make([]adm.Dep, 1, 1)}
	depB.Deps[0] = adm.Dep{adm.Component{"D", "host4"}, 1}
	archmodel[adm.Component{"B", "host2"}] = depB

	depC := adm.DepList{make([]adm.Dep, 1, 1)}
	depC.Deps[0] = adm.Dep{adm.Component{"D", "host4"}, 1}
	archmodel[adm.Component{"C", "host3"}] = depC

	depD := adm.DepList{}
	archmodel[adm.Component{"D", "host4"}] = depD

	Create(archmodel)
}
