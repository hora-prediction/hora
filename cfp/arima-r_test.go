package cfp

import (
	"log"
	"math"
	"os"
	"testing"
	"time"

	"github.com/spf13/viper"
	"gopkg.in/ory-am/dockertest.v3"

	"github.com/teeratpitakrat/hora/rbridge"
)

var epsilon = 1e-12

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
		time.Sleep(2 * time.Second)

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			_, err = rbridge.GetRSession("test-arima")
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

func TestInsert(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateLinearTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, tsPoint := range tsPoints {
		arimaCfp.Insert(tsPoint)
	}
	bufTSPoints := arimaCfp.TSPoints()
	for i, bufTSPoint := range bufTSPoints {
		expected := tsPoints[i]
		if bufTSPoint != expected {
			t.Errorf("Expected: %v but got %v", expected, bufTSPoint)
		}
	}
}

func TestInsertTSPointsWithWrongTimestamp(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateLinearTSPointsWithWrongTimestamp(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, tsPoint := range tsPoints {
		arimaCfp.Insert(tsPoint)
	}
	bufTSPoints := arimaCfp.TSPoints()
	for i, bufTSPoint := range bufTSPoints {
		expected := tsPoints[i]
		if bufTSPoint != expected {
			t.Errorf("Expected: %v but got %v", expected, bufTSPoint)
		}
	}
}

func TestInsertMoreThanBufferLength(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateLinearTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 5*time.Minute, 10*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, tsPoint := range tsPoints {
		arimaCfp.Insert(tsPoint)
	}
	bufTSPoints := arimaCfp.TSPoints()
	for i, bufTSPoint := range bufTSPoints {
		expected := tsPoints[i+10]
		if bufTSPoint != expected {
			t.Errorf("Expected: %v but got %v", expected, bufTSPoint)
		}
	}
}
func TestInsertMissingTSPoints(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, missingTSPoints, completeTSPoints := CreateMissingTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 5*time.Minute, 20*time.Minute, 70)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for _, tsPoint := range missingTSPoints {
		arimaCfp.Insert(tsPoint)
	}
	bufTSPoints := arimaCfp.TSPoints()
	for i, bufTSPoint := range bufTSPoints {
		expected := completeTSPoints[i]
		if bufTSPoint != expected {
			t.Errorf("Expected: %v but got %v", expected, bufTSPoint)
		}
	}
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

func TestPredictSeasonalData0percFail(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	comp, tsPoints := CreateSeasonalTSPoints(t)
	arimaCfp, err := NewArimaR(comp, time.Minute, 2*time.Minute, 40*time.Minute, 0)
	if err != nil {
		t.Error("Error getting new ArimaR", err)
		return
	}
	for i := range tsPoints {
		arimaCfp.Insert(tsPoints[i])
	}
	cfpRes, err := arimaCfp.Predict()
	if err != nil {
		t.Errorf("Error making prediction: %s", err)
	}
	predtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:41 UTC")
	if !cfpRes.Predtime.Equal(predtime) {
		t.Errorf("Expected prediction time at %s but got %s", predtime, cfpRes.Predtime)
	}
	expected := 0.499999985000 // TODO: double-check
	if math.Abs(cfpRes.FailProb-expected) > epsilon {
		t.Errorf("Expected failure probability of %.12f but got %.12f", expected, cfpRes.FailProb)
	}
}
