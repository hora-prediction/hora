package adm

import (
	"io/ioutil"
	"log"
	"os"
	"testing"
)

func TestReadWriteFile(t *testing.T) {
	// Export
	m := CreateSmallADM(t)

	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		log.Fatal("Error creating temp dir", err)
	}
	defer os.RemoveAll(dir) // clean up

	m.WriteFile(dir + "/m.json")

	// Import
	mRead, err := ReadFile(dir + "/m.json")
	if err != nil {
		t.Error("Error importing ADM", err)
	}
	if len(m) != len(mRead) {
		t.Error("Lengths of exported and imported ADMs are not equal")
	}
	for k, v := range m {
		if v.Caller != mRead[k].Caller {
			t.Error("Expected %v but got %v", v.Caller, mRead[k].Caller)
		}
		if len(v.Dependencies) != len(mRead[k].Dependencies) {
			t.Error("Expected %v dependencies but got %v", len(v.Dependencies), len(mRead[k].Dependencies))
		}
		// TODO: check all dependencies
	}
}
