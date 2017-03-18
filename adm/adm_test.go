package adm

import (
	"testing"
)

func TestADMSmall(t *testing.T) {
	m := make(ADM)

	compA := Component{"A", "host1"}
	compB := Component{"B", "host2"}
	compC := Component{"C", "host3"}
	compD := Component{"D", "host4"}

	depA := DepList{compA, make([]Dep, 2, 2)}
	depA.Component = compA
	depA.Deps[0] = Dep{compB, 0.5}
	depA.Deps[1] = Dep{compC, 0.5}
	m[compA.UniqName()] = depA

	depB := DepList{compB, make([]Dep, 1, 1)}
	depB.Component = compB
	depB.Deps[0] = Dep{compD, 1}
	m[compB.UniqName()] = depB

	depC := DepList{compC, make([]Dep, 1, 1)}
	depC.Component = compC
	depC.Deps[0] = Dep{compD, 1}
	m[compC.UniqName()] = depC

	depD := DepList{}
	depD.Component = compD
	m[compD.UniqName()] = depD

	for c, v := range m {
		switch c {
		case compA.UniqName():
			expected := 2
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
			if v.Deps[0].Component.UniqName() != "host2_B" || v.Deps[0].Weight != 0.5 {
				t.Error("Wrong value")
			}
			if v.Deps[1].Component.UniqName() != "host3_C" || v.Deps[1].Weight != 0.5 {
				t.Error("Wrong value")
			}
		case compB.UniqName():
			expected := 1
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
			if v.Deps[0].Component.UniqName() != "host4_D" || v.Deps[0].Weight != 1 {
				t.Error("Wrong value")
			}
		case compC.UniqName():
			expected := 1
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
			if v.Deps[0].Component.UniqName() != "host4_D" || v.Deps[0].Weight != 1 {
				t.Error("Wrong value")
			}
		case compD.UniqName():
			expected := 0
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
		}
	}
}
