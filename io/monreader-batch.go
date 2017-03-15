package io

import (
	"encoding/json"
	"log"
	"sort"
	"time"

	"github.com/teeratpitakrat/hora/model/adm"

	"github.com/influxdata/influxdb/client/v2"
)

type InfluxMonDatReader struct {
	archdepmod adm.ADM
	addr       string
	username   string
	password   string
	db         string
	batch      bool
	starttime  time.Time
	endtime    time.Time
	ch         chan MonDataPoint
}

func (r *InfluxMonDatReader) Read() {
	var monData MonData

	// map to store last timestamp of each component

	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.addr,
		Username: r.username,
		Password: r.password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		return
	}

	for c, _ := range r.archdepmod {
		// TODO: for batch mode, get first timestamp in db for this component
		// and read until the end
		cmd := "select percentile(\"response_time\",95) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "' and time >= 1487341677665666724 group by time(1m)"
		q := client.Query{
			Command:  cmd,
			Database: r.db,
		}
		response, err := clnt.Query(q)
		if err != nil {
			log.Fatal("Error: cannot query data with cmd=", cmd, err)
			continue
		}
		if response.Error() != nil {
			log.Fatal("Error: bad response with cmd=", cmd, response.Error())
			continue
		}
		res := response.Results

		// TODO: check if res is nil

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
		r.ch <- d
	}
	close(r.ch)
	return
}
