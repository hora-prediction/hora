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

func (r *InfluxMonDatReader) Read() {
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.addr,
		Username: r.username,
		Password: r.password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		return
	}
	if r.batch {
		r.readBatch(clnt)
	} else {
		r.readRealtime(clnt)
	}
}

func (r *InfluxMonDatReader) readBatch(clnt client.Client) {
	var monDat MonDat
	for c, _ := range r.archdepmod {
		// Get first and last timestamp of this component in influxdb
		var curtimestamp, firsttimestamp, lasttimestamp time.Time
		firsttimestamp, lasttimestamp = r.getFirstAndLastTimestamp(clnt, c)
		// Get the larger starttime
		if r.starttime.After(firsttimestamp) {
			curtimestamp = r.starttime.Add(-time.Nanosecond)
		} else {
			curtimestamp = firsttimestamp.Add(-time.Nanosecond)
		}
		// TODO: query for different types of components

	LoopChunk: // Loop to get all data because InfluxDB return max. 10000 records by default
		for {
			cmd := "select percentile(\"response_time\",95) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "' and time > " + strconv.FormatInt(curtimestamp.UnixNano(), 10) + " group by time(1m)"
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
					break LoopChunk // break chunk loop if timestamp of current query result exceeds the lasttimestamp the defined endtime
				}
				if row[1] != nil {
					val, _ := row[1].(json.Number).Float64()
					point := MonDatPoint{c, t, val}
					monDat = append(monDat, point)
				} else {
					point := MonDatPoint{c, t, 0}
					monDat = append(monDat, point)
				}
				curtimestamp = t
				// TODO: move time forward at least on step
				// there is a bug when start and end time are very close to each other
				// (less than time resolution)
			}
		}
	}
	// sort all data points by time
	sort.Sort(monDat)
	for _, d := range monDat {
		r.ch <- d
	}
	close(r.ch)
	return
}

func (r *InfluxMonDatReader) readRealtime(clnt client.Client) {
}

func (r *InfluxMonDatReader) getFirstAndLastTimestamp(clnt client.Client, c adm.Component) (time.Time, time.Time) {
	var firsttimestamp, lasttimestamp time.Time
	cmd := "select * from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "' order by time limit 1"
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

	cmd = "select * from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "' order by time desc limit 1"
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
