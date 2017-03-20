package adm

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestReadWriteFile(t *testing.T) {
	// Export
	m := New()
	compA := Component{"method1()", "host-1", "responsetime"}
	compB := Component{"method2(param)", "host-2", "responsetime"}
	compC := Component{"method3()", "host-3", "responsetime"}
	compD := Component{"method4(param1, param2)", "host-4", "responsetime"}

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

	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	m.WriteFile(dir + "/m.json")

	// Import
	mRead, err := ReadFile(dir + "/m.json")
	if err != nil {
		t.Error(err)
	}
	if len(m) != len(mRead) {
		t.Error("Lengths of exported and imported ADMs are equal")
	}
	for k, v := range m {
		if v.Component != mRead[k].Component {
			t.Error("Expected %v but got %v", v.Component, mRead[k].Component)
		}
		if len(v.Dependencies) != len(mRead[k].Dependencies) {
			t.Error("Expected %v dependencies but got %v", len(v.Dependencies), len(mRead[k].Dependencies))
		}
		// TODO: check all dependencies
	}
}
