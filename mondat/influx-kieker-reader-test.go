package mondat

import (
	"testing"
	"time"

	"github.com/teeratpitakrat/hora/adm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

func WriteTestPastData(t *testing.T) {
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     viper.GetString("influxdb.kieker.addr"),
		Username: "root",
		Password: "root",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB client. %s", err)
	}

	// Create a new point batch
	bp, err := client.NewBatchPoints(client.BatchPointsConfig{
		Database:  "TestKiekerDB",
		Precision: "ns",
	})
	if err != nil {
		t.Errorf("Cannot create InfluxDB batch point. %s", err)
	}

	// Create a point and add to batch
	m := adm.CreateSmallADM(t)
	for _, depInfo := range m {
		// TODO: create kieker records and write to influxdb

		//tags := map[string]string{"cpu": "cpu-total"}
		//fields := map[string]interface{}{
		//"idle":   10.1,
		//"system": 53.3,
		//"user":   46.6,
		//}
		pt, err := client.NewPoint("OperationExecution", tags, fields, time.Now())
		if err != nil {
			t.Errorf("Cannot create InfluxDB point. %s", err)
		}
		bp.AddPoint(pt)
	}

	// Write the batch
	if err := clnt.Write(bp); err != nil {
		t.Errorf("Cannot write data to InfluxDB. %s", err)
	}
	time.Sleep(60 * time.Second)
}
