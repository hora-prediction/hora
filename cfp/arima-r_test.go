package cfp

import (
	"log"
	"math"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/ory-am/dockertest.v3"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/mondat"
	"github.com/teeratpitakrat/hora/rbridge"
)

var epsilon = 1e-12

// TODO: build mondat.TSPoint
var testdat = []float64{60, 43, 67, 50, 56, 42, 50, 65, 68, 43, 65, 34, 47, 34, 49, 41, 13, 35, 53, 56}

func TestInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	c := adm.Component{"A", "host1", "responsetime", 0}
	a, err := NewArimaR(c, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 1000}) // should be dropped
	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 2000}) // should be dropped
	for _, v := range testdat {
		a.Insert(mondat.TSPoint{c, time.Unix(0, 0), v})
	}
	dat := a.TSPoints()
	for i, v := range dat {
		expected := mondat.TSPoint{c, time.Unix(0, 0), testdat[i]}
		if v != expected {
			t.Errorf("Expected: %f but got %f", expected, v)
		}
	}
}

func TestPredict(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	c := adm.Component{"A", "host1", "responsetime", 0}
	a, err := NewArimaR(c, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}

	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 1000}) // should be dropped
	a.Insert(mondat.TSPoint{c, time.Unix(0, 0), 2000}) // should be dropped
	for _, v := range testdat {
		a.Insert(mondat.TSPoint{c, time.Unix(0, 0), v})
	}
	//log.Print(a.Predict())
	// TODO: check result
	// {48.55 21.717926902432662 75.38207309756733}
}

func TestPredictLinearData0percFail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateLinearTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 1*time.Minute, 20*time.Minute, 20.5)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, point := range tsPoints {
		arimaCfp.Insert(point)
	}
	cfpRes, err := arimaCfp.Predict()
	if err != nil {
		t.Errorf("Error making prediction: %s", err)
	}
	predtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:20 UTC")
	if !cfpRes.Predtime.Equal(predtime) {
		t.Errorf("Expected prediction time at %s but got %s", predtime, cfpRes.Predtime)
	}
	expected := 0.0
	if math.Abs(cfpRes.FailProb-expected) > epsilon {
		t.Errorf("Expected failure probability of %.12f but got %.12f", expected, cfpRes.FailProb)
	}
}

func TestPredictLinearData50percFail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateLinearTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 1*time.Minute, 20*time.Minute, 20)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, point := range tsPoints {
		arimaCfp.Insert(point)
	}
	cfpRes, err := arimaCfp.Predict()
	if err != nil {
		t.Errorf("Error making prediction: %s", err)
	}
	predtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:20 UTC")
	if !cfpRes.Predtime.Equal(predtime) {
		t.Errorf("Expected prediction time at %s but got %s", predtime, cfpRes.Predtime)
	}
	expected := 0.499999985000
	if math.Abs(cfpRes.FailProb-expected) > epsilon {
		t.Errorf("Expected failure probability of %.12f but got %.12f", expected, cfpRes.FailProb)
	}
}

func TestPredictLinearData100percFail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateLinearTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 2*time.Minute, 20*time.Minute, 20)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, point := range tsPoints {
		arimaCfp.Insert(point)
	}
	cfpRes, err := arimaCfp.Predict()
	if err != nil {
		t.Errorf("Error making prediction: %s", err)
	}
	predtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:21 UTC")
	if !cfpRes.Predtime.Equal(predtime) {
		t.Errorf("Expected prediction time at %s but got %s", predtime, cfpRes.Predtime)
	}
	expected := 1.0
	if math.Abs(cfpRes.FailProb-expected) > epsilon {
		t.Errorf("Expected failure probability of %.12f but got %.12f", expected, cfpRes.FailProb)
	}
}

func TestMain(m *testing.M) {
	if testing.Short() {
		// TODO: skip test in short mode
		code := m.Run()
		os.Exit(code)
	} else {
		// uses a sensible default on windows (tcp/http) and linux/osx (socket)
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}

		// pulls an image, creates a container based on it and runs it
		resource, err := pool.Run("teeratpitakrat/docker-r-hora", "latest", nil)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

		viper.Set("rserve.hostname", "localhost")
		viper.Set("rserve.port", resource.GetPort("6311/tcp"))
		time.Sleep(1 * time.Second)

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			//db, err = sql.Open("mysql", fmt.Sprintf("root:secret@(localhost:%s)/mysql", resource.GetPort("3306/tcp")))
			_, err = rbridge.GetRSession("test")
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Fatalf("Could not connect to docker-r-hora: %s", err)
		}

		code := m.Run()

		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}

		os.Exit(code)
	}
}
