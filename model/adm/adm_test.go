package adm

import (
	"testing"
)

func TestADMSmall(t *testing.T) {
	m := make(ADM)

	depA := DepList{make([]Dep, 2, 2)}
	depA.Deps[0] = Dep{Component{"B", "host2"}, 0.5}
	depA.Deps[1] = Dep{Component{"C", "host3"}, 0.5}
	m[Component{"A", "host1"}] = depA

	depB := DepList{make([]Dep, 1, 1)}
	depB.Deps[0] = Dep{Component{"D", "host4"}, 1}
	m[Component{"B", "host2"}] = depB

	depC := DepList{make([]Dep, 1, 1)}
	depC.Deps[0] = Dep{Component{"D", "host4"}, 1}
	m[Component{"C", "host3"}] = depC

	depD := DepList{}
	m[Component{"D", "host4"}] = depD

	for c, v := range m {
		switch c {
		case Component{"A", "host1"}:
			expected := 2
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
			if v.Deps[0].Component.GetName() != "host2_B" || v.Deps[0].Weight != 0.5 {
				t.Error("Wrong value")
			}
			if v.Deps[1].Component.GetName() != "host3_C" || v.Deps[1].Weight != 0.5 {
				t.Error("Wrong value")
			}
		case Component{"B", "host2"}:
			expected := 1
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
			if v.Deps[0].Component.GetName() != "host4_D" || v.Deps[0].Weight != 1 {
				t.Error("Wrong value")
			}
		case Component{"C", "host3"}:
			expected := 1
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
			if v.Deps[0].Component.GetName() != "host4_D" || v.Deps[0].Weight != 1 {
				t.Error("Wrong value")
			}
		case Component{"D", "host4"}:
			expected := 0
			if len(v.Deps) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Deps))
			}
		}
	}
}
