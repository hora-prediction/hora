package rbridge

import (
	"flag"
	"log"
	"os"
	"testing"
	"time"

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
		time.Sleep(2 * time.Second)

		// exponential backoff-retry, because the application in the container might not be ready to accept connections yet
		if err := pool.Retry(func() error {
			var err error
			_, err = GetRSession("test-rbridge")
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

func TestGetRSession(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping test in short mode.")
	}

	sessionA, err := GetRSession("A")
	if err != nil {
		t.Error("Error: ", err)
	}

	sessionA.Eval("a <- 2")
	sessionA.Eval("b <- 3")
	ret, err := sessionA.Eval("a+b")
	if err != nil {
		t.Error("Error: ", err)
	}
	if ret != 5.0 {
		t.Error("Expected 5 but got", ret)
	}

	sessionB, err := GetRSession("B")
	if err != nil {
		t.Error("Error: ", err)
	}

	ret, err = sessionB.Eval("a+b")
	if err == nil {
		t.Error("Expected non-nil error but got nil with return value", ret)
	}

	sessionB.Eval("a <- 2.5")
	sessionB.Eval("b <- 3.4")
	ret, err = sessionB.Eval("a+b")
	if err != nil {
		t.Error("Error: ", err)
	}
	if ret != 5.9 {
		t.Error("Expected 5.9 but got", ret)
	}

	ret, err = sessionA.Eval("a+b")
	if err != nil {
		t.Error("Error: ", err)
	}
	if ret != 5.0 {
		t.Error("Expected 5.0 but got", ret)
	}
}
