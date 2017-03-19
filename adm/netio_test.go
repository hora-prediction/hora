package adm

import (
	"log"
	"net/http"
	"testing"
	"time"

	"github.com/spf13/viper"
)

var html = []byte(`
<html><body>
<form action="/adm">
<textarea name="adm" cols="150" rows="40">
{
	"host_1_method1__": {
		"component": {
			"name": "method1()",
			"hostname": "host-1"
		},
		"dependencies": [
		{
			"component": {
				"name": "method2(param)",
				"hostname": "host-2"
			},
			"weight": 0.5
		},
		{
			"component": {
				"name": "method3()",
				"hostname": "host-3"
			},
			"weight": 0.5
		}
		]
	},
	"host_2_method2_param_": {
		"component": {
			"name": "method2(param)",
			"hostname": "host-2"
		},
		"dependencies": [
		{
			"component": {
				"name": "method4(param1, param2)",
				"hostname": "host-4"
			},
			"weight": 1
		}
		]
	},
	"host_3_method3__": {
		"component": {
			"name": "method3()",
			"hostname": "host-3"
		},
		"dependencies": [
		{
			"component": {
				"name": "method4(param1, param2)",
				"hostname": "host-4"
			},
			"weight": 1
		}
		]
	},
	"host_4_method4_param1__param2_": {
		"component": {
			"name": "method4(param1, param2)",
			"hostname": "host-4"
		},
		"dependencies": null
	}
}
</textarea>
<br />
<input type="submit" value="Update"/>
</form>
</body></html>
`)

func TestNetReader(t *testing.T) {
	viper.SetConfigName("config") // name of config file (without extension)
	viper.SetConfigType("toml")
	viper.AddConfigPath("../.")
	err := viper.ReadInConfig() // Find and read the config file
	if err != nil {             // Handle errors reading the config file
		log.Print("Fatal error config file: %s \n", err)
	}

	m := New()

	compA := Component{"method1()", "host-1"}
	compB := Component{"method2(param)", "host-2"}
	compC := Component{"method3()", "host-3"}
	compD := Component{"method4(param1, param2)", "host-4"}

	depA := DependencyInfo{compA, make([]Dependency, 2, 2)}
	depA.Component = compA
	depA.Dependencies[0] = Dependency{compB, 0.5}
	depA.Dependencies[1] = Dependency{compC, 0.5}
	m[compA.UniqName()] = depA

	depB := DependencyInfo{compB, make([]Dependency, 1, 1)}
	depB.Component = compB
	depB.Dependencies[0] = Dependency{compD, 1}
	m[compB.UniqName()] = depB

	depC := DependencyInfo{compC, make([]Dependency, 1, 1)}
	depC.Component = compC
	depC.Dependencies[0] = Dependency{compD, 1}
	m[compC.UniqName()] = depC

	depD := DependencyInfo{}
	depD.Component = compD
	m[compD.UniqName()] = depD

	r := NewNetReader(m)
	r.Serve()

	time.Sleep(100 * time.Millisecond) // Wait for server to start

	port := viper.GetString("webui.port")
	req, err := http.NewRequest("GET", "http://localhost:"+port, nil)
	if err != nil {
		log.Print(err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Error(err)
	}
	defer resp.Body.Close()
	// TODO: check response
	//body, err := ioutil.ReadAll(resp.Body)
}
