package mondat

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

func WriteTestDataInThePast(t *testing.T) {
	kiekerClnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     viper.GetString("influxdb.kieker.addr"),
		Username: "root",
		Password: "root",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB client for kieker. %s", err)
	}
	k8sClnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     viper.GetString("influxdb.k8s.addr"),
		Username: "root",
		Password: "root",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB client for k8s. %s", err)
	}

	// Create a new point batch
	kiekerBp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "kiekerTest",
		Precision: "ns",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB batch point for kieker. %s", err)
	}
	k8sBp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "k8sTest",
		Precision: "ns",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB batch point for k8s. %s", err)
	}

	// Create a point and add to batch
	m := adm.CreateSmallADMWithHW(t)
	for _, depInfo := range m {
		caller := depInfo.Caller
		switch caller.Type {
		case "responsetime":
			timestamp, err := time.Parse("02 Jan 06 15:04:05 MST", "01 Jan 17 00:00:10 UTC")
			if err != nil {
				t.Errorf("Error creating time for test data: %s", err)
			}
			responseTime, err := time.ParseDuration("1s")
			if err != nil {
				t.Errorf("Error creating test data: %s", err)
			}
			for i := 0; i < 200; i++ {
				tags := map[string]string{
					"hostname":           caller.Hostname,
					"operationSignature": caller.Name,
				}
				fields := map[string]interface{}{
					"eoi":          0,
					"ess":          0,
					"responseTime": responseTime.Nanoseconds(),
					"sessionId":    "<no-session-id>",
					"tin":          timestamp.Add(-responseTime).UnixNano(),
					"tout":         timestamp.UnixNano(),
					"traceId":      759700962142060545,
				}
				pt, err := client.NewPoint("OperationExecution", tags, fields, timestamp)
				if err != nil {
					t.Errorf("Cannot create InfluxDB point. %s", err)
				}
				kiekerBp.AddPoint(pt)
				timestamp = timestamp.Add(500 * time.Millisecond)
				responseTime = responseTime + time.Second
			}
		case "cpu":
			timestamp, err := time.Parse("02 Jan 06 15:04:05 MST", "01 Jan 17 00:00:00 UTC")
			if err != nil {
				t.Errorf("Error creating time for test data: %s", err)
			}
			if err != nil {
				t.Errorf("Error creating test data: %s", err)
			}
			for i := 0; i < 3; i++ {
				tags := map[string]string{
					"container_base_image": "container-base-image",
					"container_name":       "container-name",
					"host_id":              "10.0.12.148",
					"hostname":             "10.0.12.148",
					"label":                "name:label",
					"namespace_id":         "4aa00d64-100b-11e7-b602-fa163e06ca8c",
					"namespace_name":       "default",
					"nodename":             "10.0.12.148",
					"pod_id":               "55938b27-101a-11e7-88c4-fa163e06ca8c",
					"pod_name":             caller.Hostname,
					"type":                 "pod_container",
				}
				fields := map[string]interface{}{
					"value": 100,
				}
				pt, err := client.NewPoint("cpu/usage_rate", tags, fields, timestamp)
				if err != nil {
					t.Errorf("Cannot create InfluxDB point. %s", err)
				}
				k8sBp.AddPoint(pt)
				timestamp = timestamp.Add(1 * time.Minute)
			}
		case "memory":
			timestamp, err := time.Parse("02 Jan 06 15:04:05 MST", "01 Jan 17 00:00:00 UTC")
			if err != nil {
				t.Errorf("Error creating time for test data: %s", err)
			}
			if err != nil {
				t.Errorf("Error creating test data: %s", err)
			}
			for i := 0; i < 3; i++ {
				tags := map[string]string{
					"container_base_image": "container-base-image",
					"container_name":       "container-name",
					"host_id":              "10.0.12.148",
					"hostname":             "10.0.12.148",
					"label":                "name:label",
					"namespace_id":         "4aa00d64-100b-11e7-b602-fa163e06ca8c",
					"namespace_name":       "default",
					"nodename":             "10.0.12.148",
					"pod_id":               "55938b27-101a-11e7-88c4-fa163e06ca8c",
					"pod_name":             caller.Hostname,
					"type":                 "pod_container",
				}
				fields := map[string]interface{}{
					"value": 1e9,
				}
				pt, err := client.NewPoint("memory/usage", tags, fields, timestamp)
				if err != nil {
					t.Errorf("Cannot create InfluxDB point. %s", err)
				}
				k8sBp.AddPoint(pt)
				timestamp = timestamp.Add(1 * time.Minute)
			}
		}
	}

	// Write the batch
	if err := kiekerClnt.Write(kiekerBp); err != nil {
		t.Errorf("Cannot write kieker data to InfluxDB. %s", err)
	}
	if err = k8sClnt.Write(k8sBp); err != nil {
		t.Errorf("Cannot write k8s data to InfluxDB. %s", err)
	}
	//time.Sleep(60 * time.Second)
}
