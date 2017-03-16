package io

import (
	"encoding/json"
	"log"
	"sort"
	"strconv"
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
	ch         chan MonDatPoint
	// aggregation type for each component type
	// time resolution
}

func (r *InfluxMonDatReader) Read() <-chan MonDatPoint {
	ch := make(chan MonDatPoint)
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.addr,
		Username: r.username,
		Password: r.password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		close(ch)
		return ch
	}
	if r.batch {
		go r.readBatch(clnt, ch)
	} else {
		go r.readRealtime(clnt, ch)
	}
	return ch
}

func (r *InfluxMonDatReader) readBatch(clnt client.Client, ch chan MonDatPoint) {
	var monDat MonDat
	for _, d := range r.archdepmod {
		// Get first and last timestamp of this component in influxdb
		var curtimestamp, firsttimestamp, lasttimestamp time.Time
		firsttimestamp, lasttimestamp = r.getFirstAndLastTimestamp(clnt, d.Component)
		// Get the larger starttime
		if r.starttime.After(firsttimestamp) {
			curtimestamp = r.starttime.Add(-time.Nanosecond)
		} else {
			curtimestamp = firsttimestamp.Add(-time.Nanosecond)
		}

		// TODO: query for different types of components

	LoopChunk: // Loop to get all data because InfluxDB return max. 10000 records by default
		for {
			cmd := "select percentile(\"response_time\",95) from operation_execution where \"hostname\" = '" + d.Component.Hostname + "' and \"operation_signature\" = '" + d.Component.Name + "' and time > " + strconv.FormatInt(curtimestamp.UnixNano(), 10) + " and time <= " + strconv.FormatInt(lasttimestamp.UnixNano(), 10) + " group by time(1m)"
			q := client.Query{
				Command:  cmd,
				Database: r.db,
			}
			response, err := clnt.Query(q)
			if err != nil {
				log.Fatal("Error: cannot query data with cmd=", cmd, err)
				break
			}
			if response.Error() != nil {
				log.Fatal("Error: bad response with cmd=", cmd, response.Error())
				break
			}
			res := response.Results

			if len(res[0].Series) == 0 {
				break // break if no more data is returned
			}
			// Parse time and response time
			for _, row := range res[0].Series[0].Values {
				t, err := time.Parse(time.RFC3339, row[0].(string))
				if err != nil {
					log.Fatal(err)
				}

				if t.After(lasttimestamp) || (!r.endtime.IsZero() && t.After(r.endtime)) {
					break LoopChunk // break chunk loop if timestamp of current query result exceeds the lasttimestamp or the defined endtime
				}
				if row[1] != nil {
					val, _ := row[1].(json.Number).Float64()
					point := MonDatPoint{d.Component, t, val}
					monDat = append(monDat, point)
				} else {
					point := MonDatPoint{d.Component, t, 0}
					monDat = append(monDat, point)
				}
				// preventing querying the same record forever
				if t.Sub(curtimestamp) < time.Minute {
					curtimestamp = curtimestamp.Add(time.Minute)
				} else {
					curtimestamp = t
				}
			}
		}
	}
	// sort all data points by time
	sort.Sort(monDat)
	for _, d := range monDat {
		ch <- d
	}
	close(ch)
	return
}

func (r *InfluxMonDatReader) readRealtime(clnt client.Client, ch chan MonDatPoint) {
}

func (r *InfluxMonDatReader) getFirstAndLastTimestamp(clnt client.Client, c adm.Component) (time.Time, time.Time) {
	var firsttimestamp, lasttimestamp time.Time
	cmd := "select first(response_time) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "'"
	q := client.Query{
		Command:  cmd,
		Database: r.db,
	}
	response, err := clnt.Query(q)
	if err != nil {
		log.Fatal("Error: cannot query data with cmd=", cmd, err)
		return time.Unix(0, 0), time.Unix(0, 0) // TODO: get last timestamp
	}
	if response.Error() != nil {
		log.Fatal("Error: bad response with cmd=", cmd, response.Error())
		return time.Unix(0, 0), time.Unix(0, 0)
	}
	res := response.Results
	firsttimestamp, err = time.Parse(time.RFC3339, res[0].Series[0].Values[0][0].(string))

	cmd = "select last(response_time) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "'"
	q = client.Query{
		Command:  cmd,
		Database: r.db,
	}
	response, err = clnt.Query(q)
	if err != nil {
		log.Fatal("Error: cannot query data with cmd=", cmd, err)
		return time.Unix(0, 0), time.Unix(0, 0)
	}
	if response.Error() != nil {
		log.Fatal("Error: bad response with cmd=", cmd, response.Error())
		return time.Unix(0, 0), time.Unix(0, 0)
	}
	res = response.Results
	lasttimestamp, err = time.Parse(time.RFC3339, res[0].Series[0].Values[0][0].(string))

	return firsttimestamp, lasttimestamp
}
