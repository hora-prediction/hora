package mondat

import (
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	//"github.com/teeratpitakrat/hora/adm"

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

		viper.Set("influxdb.kieker.addr", "http://localhost:"+influxDBPort)
		viper.Set("influxdb.kieker.username", "root")
		viper.Set("influxdb.kieker.password", "root")
		viper.Set("influxdb.kieker.db", "kieker")

		viper.Set("influxdb.k8s.addr", "http://localhost:"+influxDBPort)
		viper.Set("influxdb.k8s.username", "root")
		viper.Set("influxdb.k8s.password", "root")
		viper.Set("influxdb.k8s.db", "kieker")

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

//func TestReadBatch(t *testing.T) {
//if testing.Short() {
//t.Skip("skipping test in short mode.")
//}
//viper.SetConfigName("config") // name of config file (without extension)
//viper.SetConfigType("toml")
//viper.AddConfigPath("../.")
//err := viper.ReadInConfig() // Find and read the config file
//if err != nil {             // Handle errors reading the config file
//log.Print("Fatal error config file: %s \n", err)
//}

//viper.SetDefault("influxdb.addr", "http://localhost:8086")
//viper.SetDefault("influxdb.username", "root")
//viper.SetDefault("influxdb.password", "root")
//viper.SetDefault("influxdb.db.kieker", "kieker")

////TODO
//}

func TestReadBatch(t *testing.T) {
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

	WriteTestPastData(t)
	// TODO: test

}

func createTestDBs(t *testing.T, clnt client.Client) {
	cmd := fmt.Sprintf("CREATE DATABASE %s", "TestKiekerDB")
	q := client.Query{
		Command:  cmd,
		Database: "TestKiekerDB",
	}
	_, err := clnt.Query(q)
	if err != nil {
		t.Errorf("Cannot create test DB for Kieker. %s", err)
	}
	cmd = fmt.Sprintf("CREATE DATABASE %s", "TestK8sDB")
	q = client.Query{
		Command:  cmd,
		Database: "TestK8sDB",
	}
	_, err = clnt.Query(q)
	if err != nil {
		t.Errorf("Cannot create test DB for k8s. %s", err)
	}
}

func dropTestDBs(t *testing.T, clnt client.Client) {
	cmd := fmt.Sprintf("DROP DATABASE %s", "TestKiekerDB")
	q := client.Query{
		Command:  cmd,
		Database: "TestKiekerDB",
	}
	_, err := clnt.Query(q)
	if err != nil {
		t.Errorf("Cannot drop test DB for kieker. %s", err)
	}
	cmd = fmt.Sprintf("DROP DATABASE %s", "TestK8sDB")
	q = client.Query{
		Command:  cmd,
		Database: "TestK8sDB",
	}
	_, err = clnt.Query(q)
	if err != nil {
		t.Errorf("Cannot drop test DB for k8s. %s", err)
	}
}
