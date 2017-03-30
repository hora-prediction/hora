package fpm

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

	//"github.com/hora-prediction/hora/adm"
	//"github.com/hora-prediction/hora/cfp"
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
			_, err = rbridge.GetRSession("test-pm")
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
func TestCreate(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}
	//m := adm.CreateSmallADMWithHW(t)

	//f, fpmResultCh, err := NewBayesNetR(m)
	//if err != nil {
	//t.Fatalf("Error creating BayesNetR. %s", err)
	//}

	//cfpResult := cfp.Result{
	//Component: adm.Component{
	//Name:     "D",
	//Hostname: "host4",
	//Type:     "responsetime",
	//Called:   0,
	//},
	//Timestamp: time.Unix(0, 0),
	//Predtime:  time.Unix(0, 300),
	//PredMean:  0,
	//PredLB:    0,
	//PredUB:    0,
	//PredSd:    1,
	//FailProb:  0.0,
	//}
	//f.UpdateCfpResult(cfpResult)
	//fpmResult := <-fpmResultCh
	//if err != nil {
	//t.Error("Error making prediction", err)
	//}
	//// TODO: more precision checks
	//fprobA := fpmResult.FailProbs[adm.Component{"A", "host1", "responsetime", 0}]
	//if fprobA != 0 {
	//t.Error("Expected: 0 but got", fprobA)
	//}
	//fprobB := fpmResult.FailProbs[adm.Component{"B", "host2", "responsetime", 0}]
	//if fprobB != 0 {
	//t.Error("Expected: 0 but got", fprobB)
	//}
	//fprobC := fpmResult.FailProbs[adm.Component{"C", "host3", "responsetime", 0}]
	//if fprobC != 0 {
	//t.Error("Expected: 0 but got", fprobC)
	//}
	//fprobD := fpmResult.FailProbs[adm.Component{"D", "host4", "responsetime", 0}]
	//if fprobD != 0 {
	//t.Error("Expected: 0 but got", fprobD)
	//}
}
