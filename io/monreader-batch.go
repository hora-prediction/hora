package io

import (
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/teeratpitakrat/hora/influxdb"
	"github.com/teeratpitakrat/hora/model/adm"
)

func ReadMonDat(m adm.ADM, ch chan MonDataPoint) {
	var monData MonData
	monData = make([]MonDataPoint, 0, 0)

	influxClnt, err := influxdb.NewClient("http://localhost:8086", "root", "root")
	if err != nil {
		log.Fatal("Cannot get new influxdb client", err)
		return
	}

	for c, _ := range m {
		// TODO: get first timestamp in db for this component
		// and read until the end
		cmd := "select percentile(\"response_time\",95) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "' and time >= 1487341677665666724 group by time(1m)"
		res, err := influxdb.Query(*influxClnt, cmd, "kieker")
		if err != nil {
			log.Fatal("Cannot query data", err)
			return
		}

		// Parse time and response time
		for _, row := range res[0].Series[0].Values {
			t, err := time.Parse(time.RFC3339, row[0].(string))
			if err != nil {
				log.Fatal(err)
			}
			if row[1] != nil {
				val, _ := row[1].(json.Number).Float64()
				point := MonDataPoint{c, t, val}
				monData = append(monData, point)
			} else {
				point := MonDataPoint{c, t, 0}
				monData = append(monData, point)
			}
		}
	}
	// sort all data points by time
	sort.Sort(monData)
	for _, d := range monData {
		ch <- d
	}
	close(ch)
	return
}
