package resultio

import (
	"log"
	"os"
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"
	"github.com/teeratpitakrat/hora/cfp"
	"github.com/teeratpitakrat/hora/fpm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
	"gopkg.in/ory-am/dockertest.v3"
)

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
		resource, err := pool.Run("influxdb", "alpine", nil)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

		influxDBPort := resource.GetPort("8086/tcp")

		viper.Set("influxdb.hora.addr", "http://localhost:"+influxDBPort)
		viper.Set("influxdb.hora.username", "root")
		viper.Set("influxdb.hora.password", "root")
		viper.Set("influxdb.hora.db", "hora")

		log.Println("Waiting for docker container")
		time.Sleep(2 * time.Second)

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			_, err = client.NewHTTPClient(client.HTTPConfig{
				Addr:     "http://localhost:" + influxDBPort,
				Username: "root",
				Password: "root",
			})
			if err != nil {
				return err
			}
			return nil
		}); err != nil {
			log.Fatalf("Could not connect to influxdb: %s", err)
		}

		code := m.Run()

		// You can't defer this because os.Exit doesn't care for defer
		if err := pool.Purge(resource); err != nil {
			log.Fatalf("Could not purge resource: %s", err)
		}

		os.Exit(code)
	}
}
func TestWriteCfpResult(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	writer, err := New(
		viper.GetString("influxdb.hora.addr"),
		viper.GetString("influxdb.hora.username"),
		viper.GetString("influxdb.hora.password"),
		viper.GetString("influxdb.hora.db"),
	)
	if err != nil {
		t.Error(err)
	}
	a := adm.Component{"A", "host1", "responsetime", 0}
	b := adm.Component{"B", "host2", "responsetime", 0}
	resulta := cfp.Result{
		Component: a,
		Timestamp: time.Now(),
		Predtime:  time.Now().Add(10 * time.Minute),
		PredMean:  0,
		PredLB:    0,
		PredUB:    0,
		PredSd:    1,
		FailProb:  0.555,
	}
	resultb := cfp.Result{
		Component: b,
		Timestamp: time.Now(),
		Predtime:  time.Now().Add(10 * time.Minute),
		PredMean:  0,
		PredLB:    0,
		PredUB:    0,
		PredSd:    1,
		FailProb:  0.666,
	}
	writer.WriteCfpResult(resulta)
	writer.WriteCfpResult(resultb)

	// TODO: read and check
}

func TestWriteFpmResult(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	writer, err := New(
		viper.GetString("influxdb.hora.addr"),
		viper.GetString("influxdb.hora.username"),
		viper.GetString("influxdb.hora.password"),
		viper.GetString("influxdb.hora.db"),
	)
	if err != nil {
		t.Error(err)
	}

	a := adm.Component{"A", "host1", "responsetime", 0}
	b := adm.Component{"B", "host2", "responsetime", 0}
	failProbs := make(map[adm.Component]float64)
	failProbs[a] = 0.2
	failProbs[b] = 0.3

	result := fpm.Result{
		failProbs,
		time.Now(),
		time.Now().Add(10 * time.Minute),
	}
	writer.WriteFpmResult(result)

	// TODO: read and check
}
