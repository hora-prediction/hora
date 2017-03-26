package adm

import (
	"io/ioutil"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestReadWrite(t *testing.T) {
	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		t.Errorf("Error creating temp dir: %s", err)
	}
	defer os.RemoveAll(dir) // clean up
	viper.Set("adm.filewatcher.path", dir+"/adm.json")

	w := NewFileWatcher()
	refModel := CreateSmallADM(t)
	w.m = refModel
	w.Write()
	w.Read()
	if w.m.String() != refModel.String() {
		t.Errorf("Error writing and reading ADM. Expected\n%s\n but got\n%s", refModel.String(), w.m.String())
	}
}

func TestWatch(t *testing.T) {
	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		t.Errorf("Error creating temp dir: %s", err)
	}
	defer os.RemoveAll(dir) // clean up
	viper.Set("adm.filewatcher.path", dir+"/adm.json")

	w := NewFileWatcher()
	refModel := CreateSmallADM(t)
	w.m = refModel
	w.Write()
	w.Start()

	select {
	case newModel := <-w.admCh:
		if newModel.String() != refModel.String() {
			t.Errorf("Expected\n%s\nbut got\n%v", refModel.String(), newModel.String())
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out while reading ADM")
	}
}

func TestUpdateADM(t *testing.T) {
}
