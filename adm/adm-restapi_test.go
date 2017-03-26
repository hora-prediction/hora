package adm

import (
	"bytes"
	"log"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strconv"
	"strings"
	"testing"
	"time"

	"github.com/spf13/viper"
)

func TestRestApiEmptyADM(t *testing.T) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}
	viper.SetDefault("adm.restapi.port", "8080")

	r := NewRestApi()

	port := viper.GetString("adm.restapi.port")
	req, err := http.NewRequest("GET", "http://localhost:"+port, nil)
	if err != nil {
		log.Print(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(r.getHandler)
	handler.ServeHTTP(rr, req)

	if status := rr.Code; status != http.StatusOK {
		t.Errorf("Handler returned wrong status code: got %v want %v", status, http.StatusOK)
	}

	expected := string(htmlHeader) + "{}" + string(htmlFooter)
	if rr.Body.String() != expected {
		t.Errorf("Handler returned unexpected body: got %v want %v",
			rr.Body.String(), expected)
	}
}

func TestRestApiUpdateADM(t *testing.T) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}
	viper.SetDefault("adm.restapi.port", "8080")

	r := NewRestApi()

	m := CreateSmallADM(t)
	data := url.Values{}
	data.Set("adm", m.String())

	port := viper.GetString("adm.restapi.port")
	req, err := http.NewRequest(
		"POST",
		"http://localhost:"+port,
		bytes.NewBufferString(data.Encode()),
	)
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	req.Header.Add("Content-Length", strconv.Itoa(len(data.Encode())))
	if err != nil {
		log.Print(err)
	}

	rr := httptest.NewRecorder()
	handler := http.HandlerFunc(r.postHandler)
	handler.ServeHTTP(rr, req)

	if !strings.Contains(rr.Body.String(), m.String()) {
		t.Errorf("Handler returned unexpected body: got\n%v that does not contain\n%v", rr.Body.String(), m.String())
	}

	select {
	case newModel := <-r.admCh:
		if newModel.String() != m.String() {
			t.Errorf("Expected %v but got %v", m.String(), newModel.String())
		}
	case <-time.After(1 * time.Second):
		t.Errorf("Timed out while updating ADM")
	}
}
