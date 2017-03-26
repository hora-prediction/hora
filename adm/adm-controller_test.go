package adm

import (
	"bytes"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestFileWatcher(t *testing.T) {
	viper.Set("adm.filewatcher.enabled", true)
	viper.Set("adm.restapi.enabled", false)

	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		t.Errorf("Error creating temp dir: %s", err)
	}
	defer os.RemoveAll(dir) // clean up
	viper.Set("adm.filewatcher.path", dir+"/adm.json")

	refModel := CreateSmallADM(t)
	w := NewFileWatcher()
	w.m = refModel
	// Update ADM file
	w.Write()

	controller := NewController()

	select {
	case newModel := <-controller.AdmCh:
		if newModel.String() != refModel.String() {
			t.Errorf("Expected\n%s\nbut got\n%v", refModel.String(), newModel.String())
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out while reading ADM")
	}
}

func TestRestApi(t *testing.T) {
	viper.Set("adm.filewatcher.enabled", false)
	viper.Set("adm.restapi.enabled", true)
	viper.SetDefault("adm.restapi.port", "8080")

	dir, err := ioutil.TempDir("", "adm")
	if err != nil {
		t.Errorf("Error creating temp dir: %s", err)
	}
	defer os.RemoveAll(dir) // clean up

	controller := NewController()

	refModel := CreateSmallADM(t)
	go func() {
		time.Sleep(500 * time.Millisecond)
		// Update ADM via REST API
		data := url.Values{}
		data.Set("adm", refModel.String())
		port := viper.GetString("adm.restapi.port")
		req, err := http.NewRequest(
			"POST",
			"http://localhost:"+port+"/adm",
			bytes.NewBufferString(data.Encode()),
		)
		req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
		req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
		if err != nil {
			t.Errorf("Error creating http request: %s", err)
		}
		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			t.Errorf("Error sending ADM: %s", err)
		}
		if resp == nil {
			t.Error("Error sending ADM: received empty response. Please check target address and port.")
		}
	}()

	select {
	case newModel := <-controller.AdmCh:
		if newModel.String() != refModel.String() {
			t.Errorf("Expected\n%s\nbut got\n%v", refModel.String(), newModel.String())
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out while reading ADM")
	}
}
