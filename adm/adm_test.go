package adm

import (
	"testing"
)

func TestADMSmall(t *testing.T) {
	m := New()

	compA := Component{"method1()", "host-1"}
	compB := Component{"method2(param)", "host-2"}
	compC := Component{"method3()", "host-3"}
	compD := Component{"method4(param1, param2)", "host-4"}

	depA := DependencyInfo{compA, make([]Dependency, 2, 2)}
	depA.Component = compA
	depA.Dependencies[0] = Dependency{compB, 0.5}
	depA.Dependencies[1] = Dependency{compC, 0.5}
	m[compA.UniqName()] = depA

	depB := DependencyInfo{compB, make([]Dependency, 1, 1)}
	depB.Component = compB
	depB.Dependencies[0] = Dependency{compD, 1}
	m[compB.UniqName()] = depB

	depC := DependencyInfo{compC, make([]Dependency, 1, 1)}
	depC.Component = compC
	depC.Dependencies[0] = Dependency{compD, 1}
	m[compC.UniqName()] = depC

	depD := DependencyInfo{}
	depD.Component = compD
	m[compD.UniqName()] = depD

	for c, v := range m {
		switch c {
		case compA.UniqName():
			expected := 2
			if len(v.Dependencies) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Dependencies))
			}
			if v.Dependencies[0].Component.UniqName() != "host_2_method2_param_" || v.Dependencies[0].Weight != 0.5 {
				t.Error("Wrong value")
			}
			if v.Dependencies[1].Component.UniqName() != "host_3_method3__" || v.Dependencies[1].Weight != 0.5 {
				t.Error("Wrong value")
			}
		case compB.UniqName():
			expected := 1
			if len(v.Dependencies) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Dependencies))
			}
			if v.Dependencies[0].Component.UniqName() != "host_4_method4_param1__param2_" || v.Dependencies[0].Weight != 1 {
				t.Error("Wrong value")
			}
		case compC.UniqName():
			expected := 1
			if len(v.Dependencies) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Dependencies))
			}
			if v.Dependencies[0].Component.UniqName() != "host_4_method4_param1__param2_" || v.Dependencies[0].Weight != 1 {
				t.Error("Wrong value")
			}
		case compD.UniqName():
			expected := 0
			if len(v.Dependencies) != expected {
				t.Error("Expected: ", expected, " but got ", len(v.Dependencies))
			}
		}
	}
}
