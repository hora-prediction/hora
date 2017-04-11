package eval

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hora-prediction/hora/adm"
	"github.com/hora-prediction/hora/cfp"
	"github.com/hora-prediction/hora/fpm"
	"github.com/hora-prediction/hora/mondat"

	"github.com/hora-prediction/hora/rbridge"
	"github.com/spf13/viper"
	"gopkg.in/ory-am/dockertest.v3"
)

func TestMain(m *testing.M) {
	flag.Parse()
	if testing.Short() {
		code := m.Run()
		os.Exit(code)
	} else {
		// uses a sensible default on windows (tcp/http) and linux/osx (socket)
		pool, err := dockertest.NewPool("")
		if err != nil {
			log.Fatalf("Could not connect to docker: %s", err)
		}

		// pulls an image, creates a container based on it and runs it
		resource, err := pool.Run("hora/docker-r-hora", "latest", nil)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

		viper.Set("rserve.hostname", "localhost")
		viper.Set("rserve.port", resource.GetPort("6311/tcp"))

		log.Println("Waiting for docker container")
		time.Sleep(2 * time.Second)

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			_, err = rbridge.GetRSession("test-eval")
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

func TestUpdateADM(t *testing.T) {
	viper.Set("cfp.responsetime.threshold", "500ms")
	viper.Set("cfp.responsetime.unit", "1ms")
	viper.Set("cfp.cpu.threshold", "800")
	viper.Set("cfp.memory.threshold", "800")

	evtr := New()
	m := adm.CreateSmallADM(t)
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")
	monVal := 10.0
	cfpFailProb := 0.2
	fpmFailProb := 0.3
	for i := 0; i < 60; i++ {
		fpmResult := fpm.Result{
			FailProbs: make(map[adm.Component]float64),
			Predtime:  curtime,
		}
		for _, depInfo := range m {
			component := depInfo.Caller
			tsPoint := mondat.TSPoint{
				Component: component,
				Timestamp: curtime,
				Value:     monVal,
			}
			evtr.UpdateMondat(tsPoint)
			cfpResult := cfp.Result{
				Component: component,
				Predtime:  curtime,
				FailProb:  cfpFailProb,
			}
			evtr.UpdateCfpResult(cfpResult)
			fpmResult.FailProbs[component] = fpmFailProb
		}
		evtr.UpdateFpmResult(fpmResult)
		curtime = curtime.Add(1 * time.Minute)
		monVal += 10
	}
	threshold := float64(viper.GetDuration("cfp.responsetime.threshold") / viper.GetDuration("cfp.responsetime.unit"))
	for _, timestamps := range evtr.result {
		for _, point := range timestamps {
			if point.TSPoint.Value > threshold && point.Label == 0 {
				t.Errorf("Expected label to be true for %s", point.String())
			} else if point.TSPoint.Value < threshold && point.Label == 1 {
				t.Errorf("Expected label to be false for %s", point.String())
			}
		}
	}
}

func TestROC(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}

	viper.Set("cfp.responsetime.threshold", "500ms")
	viper.Set("cfp.responsetime.unit", "1ms")
	viper.Set("cfp.cpu.threshold", "800")
	viper.Set("cfp.memory.threshold", "800")
	// TODO: tmpdir
	viper.Set("eval.outdir", "/tmp/hora")

	evtr := New()
	m := adm.CreateSmallADM(t)
	curtime, _ := time.Parse("02 Jan 06 15:04 MST", "01 Jan 17 00:00 UTC")
	monVal := 10.0
	cfpFailProb := 0.0
	fpmFailProb := 0.0
	fuzzyValue := 0.1
	for i := 0; i < 60; i++ {
		fpmResult := fpm.Result{
			FailProbs: make(map[adm.Component]float64),
			Predtime:  curtime,
		}
		for _, depInfo := range m {
			component := depInfo.Caller
			tsPoint := mondat.TSPoint{
				Component: component,
				Timestamp: curtime,
				Value:     monVal,
			}
			evtr.UpdateMondat(tsPoint)
			cfpResult := cfp.Result{
				Component: component,
				Predtime:  curtime,
				FailProb:  cfpFailProb,
			}
			evtr.UpdateCfpResult(cfpResult)
			fpmResult.FailProbs[component] = fpmFailProb
		}
		evtr.UpdateFpmResult(fpmResult)
		curtime = curtime.Add(1 * time.Minute)
		monVal += 10
		cfpFailProb += 0.01 - fuzzyValue
		fpmFailProb += 0.015 + fuzzyValue
		fuzzyValue = -fuzzyValue
	}
	evtr.ComputeROC()
	// TODO: check results
}
