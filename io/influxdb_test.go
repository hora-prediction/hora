package influxdb

import (
	"encoding/json"
	"log"
	"testing"
	"time"
)

func TestQuery(t *testing.T) {
	clnt, err := NewClient("http://localhost:8086", "root", "root")
	if err != nil {
		t.Error("Error connecting to influxdb")
	}
	res, err := Query(*clnt, "select mean(\"response_time\") from operation_execution where \"hostname\" = 'middletier-wbfr4' and time >= 1487351488081472373 group by time(1m) limit 20", "kieker")
	if err != nil {
		t.Error("Error querying influxdb")
	}
	log.Print(res)
	for i, row := range res[0].Series[0].Values {
		t, err := time.Parse(time.RFC3339, row[0].(string))
		if err != nil {
			log.Fatal(err)
		}
		log.Print(row[1])
		if row[1] != nil {
			val, _ := row[1].(json.Number).Float64()
			log.Printf("[%2d] %s: %f\n", i, t.Format(time.Stamp), val)
		}
	}
}
