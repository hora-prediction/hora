package adm

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestExportAndImport(t *testing.T) {
	// Export
	mExport := New()
	compA := Component{"method1()", "host-1"}
	compB := Component{"method2(param)", "host-2"}
	compC := Component{"method3()", "host-3"}
	compD := Component{"method4(param1, param2)", "host-4"}

	depA := DependencyInfo{compA, make([]Dependency, 2, 2)}
	depA.Component = compA
	depA.Dependencies[0] = Dependency{compB, 0.5}
	depA.Dependencies[1] = Dependency{compC, 0.5}
	mExport[compA.UniqName()] = depA

	depB := DependencyInfo{compB, make([]Dependency, 1, 1)}
	depB.Component = compB
	depB.Dependencies[0] = Dependency{compD, 1}
	mExport[compB.UniqName()] = depB

	depC := DependencyInfo{compC, make([]Dependency, 1, 1)}
	depC.Component = compC
	depC.Dependencies[0] = Dependency{compD, 1}
	mExport[compC.UniqName()] = depC

	depD := DependencyInfo{}
	depD.Component = compD
	mExport[compD.UniqName()] = depD

	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		log.Fatal(err)
	}
	defer os.RemoveAll(dir) // clean up

	Export(mExport, dir+"/mExport.json")

	// Import
	mImport, err := Import(dir + "/mExport.json")
	if err != nil {
		t.Error(err)
	}
	if len(mExport) != len(mImport) {
		t.Error("Lengths of exported and imported ADMs are equal")
	}
	for k, v := range mExport {
		if v.Component != mImport[k].Component {
			t.Error("Expected %v but got %v", v.Component, mImport[k].Component)
		}
		if len(v.Dependencies) != len(mImport[k].Dependencies) {
			t.Error("Expected %v dependencies but got %v", len(v.Dependencies), len(mImport[k].Dependencies))
		}
		// TODO: check all dependencies
	}
}
