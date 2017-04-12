package mondat

import (
	"flag"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/hora-prediction/hora/adm"

	"github.com/influxdata/influxdb/client/v2"
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
		resource, err := pool.Run("influxdb", "alpine", nil)
		if err != nil {
			log.Fatalf("Could not start resource: %s", err)
		}

		influxDBPort := resource.GetPort("8086/tcp")

		viper.Set("influxdb.kieker.addr", "http://localhost:"+influxDBPort)
		viper.Set("influxdb.kieker.username", "root")
		viper.Set("influxdb.kieker.password", "root")
		viper.Set("influxdb.kieker.db", "kieker")

		viper.Set("influxdb.k8s.addr", "http://localhost:"+influxDBPort)
		viper.Set("influxdb.k8s.username", "root")
		viper.Set("influxdb.k8s.password", "root")
		viper.Set("influxdb.k8s.db", "kieker")

		viper.Set("influxdb.locust.addr", "http://localhost:"+influxDBPort)
		viper.Set("influxdb.locust.username", "root")
		viper.Set("influxdb.locust.password", "root")
		viper.Set("influxdb.locust.db", "kieker")

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

func TestReadBatchWithZeroStarttime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	m := adm.CreateSmallADM(t)
	influxKiekerReader := InfluxKiekerReader{
		Archdepmod: m,
		KiekerDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "kiekerTest",
		},
		K8sDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "k8sTest",
		},
		LocustDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.locust.addr"),
			Username: "root",
			Password: "root",
			DbName:   "locustTest",
		},
		Batch:    true,
		Endtime:  time.Now(),
		Interval: time.Minute,
	}
	mondatCh := influxKiekerReader.Read()
	if _, ok := <-mondatCh; ok {
		t.Errorf("Expected mondatCh to be closed")
	}
}

func TestReadBatchWithZeroEndtime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	m := adm.CreateSmallADM(t)
	influxKiekerReader := InfluxKiekerReader{
		Archdepmod: m,
		KiekerDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "kiekerTest",
		},
		K8sDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "k8sTest",
		},
		LocustDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.locust.addr"),
			Username: "root",
			Password: "root",
			DbName:   "locustTest",
		},
		Batch:     true,
		Starttime: time.Now(),
		Interval:  time.Minute,
	}
	mondatCh := influxKiekerReader.Read()
	if _, ok := <-mondatCh; ok {
		t.Errorf("Expected mondatCh to be closed")
	}
}

func TestReadBatch(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     viper.GetString("influxdb.kieker.addr"),
		Username: "root",
		Password: "root",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB client. %s", err)
	}
	createTestDBs(t, clnt)
	defer dropTestDBs(t, clnt)

	WriteTestDataInThePast(t)
	m := adm.CreateSmallADMWithHW(t)
	starttime, err := time.Parse("02 Jan 06 15:04:05 MST", "01 Jan 17 00:01:02 UTC")
	endtime, err := time.Parse("02 Jan 06 15:04:05 MST", "01 Jan 17 00:03:02 UTC")
	influxKiekerReader := InfluxKiekerReader{
		Archdepmod: m,
		KiekerDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "kiekerTest",
		},
		K8sDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "k8sTest",
		},
		LocustDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.locust.addr"),
			Username: "root",
			Password: "root",
			DbName:   "locustTest",
		},
		Batch:     true,
		Starttime: starttime,
		Endtime:   endtime,
		Interval:  time.Minute,
	}
	mondatCh := influxKiekerReader.Read()
	for i := 0; i < len(m); i++ {
		select {
		case dat, ok := <-mondatCh:
			if !ok {
				t.Fatalf("Expected monitoring data but the channel is closed")
			}
			switch dat.Component.Type {
			case "responsetime":
				expected := 9.5e10
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "cpu":
				expected := 100.0
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "memory":
				expected := 1e9
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			default:
				t.Fatalf("Unknown component type: %v", dat.Component)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out while reading monitoring data from InfluxDB")
		}
	}
	for i := 0; i < len(m); i++ {
		select {
		case dat, ok := <-mondatCh:
			if !ok {
				t.Fatalf("Expected monitoring data but the channel is closed")
			}
			switch dat.Component.Type {
			case "responsetime":
				expected := 19.5e10
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "cpu":
				expected := 100.0
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "memory":
				expected := 1e9
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			default:
				t.Fatalf("Unknown component type: %v", dat.Component)
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out while reading monitoring data from InfluxDB")
		}
	}
	for i := 0; i < len(m); i++ {
		select {
		case _, ok := <-mondatCh:
			if ok {
				t.Fatalf("Expected channel to be closed.")
			}
		case <-time.After(2 * time.Second):
			t.Fatalf("Timed out while reading monitoring data from InfluxDB")
		}
	}
}

func TestReadRealtime(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping test in short mode.")
	}
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     viper.GetString("influxdb.kieker.addr"),
		Username: "root",
		Password: "root",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB client. %s", err)
	}
	createTestDBs(t, clnt)
	defer dropTestDBs(t, clnt)

	m := adm.CreateSmallADMWithHW(t)
	influxKiekerReader := InfluxKiekerReader{
		Archdepmod: m,
		KiekerDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "kiekerTest",
		},
		K8sDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.kieker.addr"),
			Username: "root",
			Password: "root",
			DbName:   "k8sTest",
		},
		LocustDb: InfluxDBConfig{
			Addr:     viper.GetString("influxdb.locust.addr"),
			Username: "root",
			Password: "root",
			DbName:   "locustTest",
		},
		Batch:    false,
		Interval: time.Minute,
	}
	log.Printf("Testing reading monitoring data in realtime. The test will take approximately 3 minutes.")
	WriteTestDataInTheFuture(t)
	mondatCh := influxKiekerReader.Read()
	for i := 0; i < len(m); i++ {
		select {
		case dat, ok := <-mondatCh:
			if !ok {
				t.Fatalf("Expected monitoring data but the channel is closed")
			}
			switch dat.Component.Type {
			case "responsetime":
				expected := 9.5e10
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "cpu":
				expected := 100.0
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "memory":
				expected := 1e9
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			default:
				t.Fatalf("Unknown component type: %v", dat.Component)
			}
		case <-time.After(70 * time.Second):
			t.Fatalf("Timed out while reading monitoring data from InfluxDB")
		}
	}
	for i := 0; i < len(m); i++ {
		select {
		case dat, ok := <-mondatCh:
			if !ok {
				t.Fatalf("Expected monitoring data but the channel is closed")
			}
			switch dat.Component.Type {
			case "responsetime":
				expected := 19.5e10
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "cpu":
				expected := 100.0
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			case "memory":
				expected := 1e9
				if dat.Value != expected {
					t.Fatalf("Expected %.0f but got %.0f", expected, dat.Value)
				}
			default:
				t.Fatalf("Unknown component type: %v", dat.Component)
			}
		case <-time.After(70 * time.Second):
			t.Fatalf("Timed out while reading monitoring data from InfluxDB")
		}
	}
}

func createTestDBs(t *testing.T, clnt client.Client) {
	cmd := fmt.Sprintf("CREATE DATABASE %s", "kiekerTest")
	q := client.Query{
		Command:  cmd,
		Database: "kiekerTest",
	}
	_, err := clnt.Query(q)
	if err != nil {
		t.Fatalf("Cannot create test DB for Kieker. %s", err)
	}
	cmd = fmt.Sprintf("CREATE DATABASE %s", "k8sTest")
	q = client.Query{
		Command:  cmd,
		Database: "k8sTest",
	}
	_, err = clnt.Query(q)
	if err != nil {
		t.Fatalf("Cannot create test DB for k8s. %s", err)
	}
}

func dropTestDBs(t *testing.T, clnt client.Client) {
	cmd := fmt.Sprintf("DROP DATABASE %s", "kiekerTest")
	q := client.Query{
		Command:  cmd,
		Database: "kiekerTest",
	}
	_, err := clnt.Query(q)
	if err != nil {
		t.Errorf("Cannot drop test DB for kieker. %s", err)
	}
	cmd = fmt.Sprintf("DROP DATABASE %s", "k8sTest")
	q = client.Query{
		Command:  cmd,
		Database: "k8sTest",
	}
	_, err = clnt.Query(q)
	if err != nil {
		t.Errorf("Cannot drop test DB for k8s. %s", err)
	}
}
