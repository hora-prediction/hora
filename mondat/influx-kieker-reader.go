package mondat

import (
	"encoding/json"
	"log"
	"sort"
	"strconv"
	"time"

	"github.com/teeratpitakrat/hora/adm"

	"github.com/influxdata/influxdb/client/v2"
	"github.com/spf13/viper"
)

type InfluxKiekerReader struct {
	Archdepmod adm.ADM
	Addr       string
	Username   string
	Password   string
	Db         string
	Batch      bool
	Starttime  time.Time
	Endtime    time.Time
	Interval   time.Duration
	// aggregation type for each component type
}

func (r *InfluxKiekerReader) Read() <-chan TSPoint {
	ch := make(chan TSPoint)
	clnt, err := client.NewHTTPClient(client.HTTPConfig{
		Addr:     r.Addr,
		Username: r.Username,
		Password: r.Password,
	})
	if err != nil {
		log.Fatal("Error: cannot create new influxdb client", err)
		close(ch)
		return ch
	}
	if r.Batch {
		log.Print("Read monitoring data in batch mode")
		go r.readBatch(clnt, ch)
	} else {
		log.Print("Read monitoring data in realtime mode")
		go r.readRealtime(clnt, ch)
	}
	return ch
}

func (r *InfluxKiekerReader) readBatch(clnt client.Client, ch chan TSPoint) {
	var tsPoints TSPoints
	for _, d := range r.Archdepmod {
		// Get first and last timestamp of this component in influxdb
		var curtimestamp, firsttimestamp, lasttimestamp time.Time
		firsttimestamp, lasttimestamp = r.getFirstAndLastTimestamp(clnt, d.Component)
		// Get the larger starttime
		if r.Starttime.After(firsttimestamp) {
			curtimestamp = r.Starttime.Add(-time.Nanosecond)
		} else {
			curtimestamp = firsttimestamp.Add(-time.Nanosecond)
		}

		// TODO: query for different types of components

	LoopChunk: // Loop to get all data because InfluxDB return max. 10000 records by default
		for {
			//cmd := "select percentile(\"response_time\",95) from operation_execution where \"hostname\" = '" + d.Component.Hostname + "' and \"operation_signature\" = '" + d.Component.Name + "' and time > " + strconv.FormatInt(curtimestamp.UnixNano(), 10) + " and time <= " + strconv.FormatInt(lasttimestamp.UnixNano(), 10) + " group by time(1m)"
			aggregation := viper.GetString("cfp.responsetime.aggregation")
			aggregationvalue := viper.GetString("cfp.responsetime.aggregationvalue")
			cmd := "select " + aggregation + "(\"response_time\"," + aggregationvalue + ") from operation_execution where \"hostname\" = '" + d.Component.Hostname + "' and \"operation_signature\" = '" + d.Component.Name + "' and time > " + strconv.FormatInt(curtimestamp.UnixNano(), 10) + " and time <= " + strconv.FormatInt(lasttimestamp.UnixNano(), 10) + " group by time(" + r.Interval.String() + ")"
			q := client.Query{
				Command:  cmd,
				Database: r.Db,
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

				if t.After(lasttimestamp) || (!r.Endtime.IsZero() && t.After(r.Endtime)) {
					break LoopChunk // break chunk loop if timestamp of current query result exceeds the lasttimestamp or the defined endtime
				}
				if row[1] != nil {
					val, _ := row[1].(json.Number).Float64()
					point := TSPoint{d.Component, t, val}
					tsPoints = append(tsPoints, point)
				} else {
					point := TSPoint{d.Component, t, 0}
					tsPoints = append(tsPoints, point)
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
	sort.Sort(tsPoints)
	for _, d := range tsPoints {
		ch <- d
	}
	close(ch)
	return
}

func (r *InfluxKiekerReader) readRealtime(clnt client.Client, ch chan TSPoint) {
	// Wait until a full minute has passed
	// TODO: wait according to r.Interval
	remainingSeconds := time.Duration(60 - time.Now().Second())
	time.Sleep(remainingSeconds * time.Second)
	// Wait a few more seconds for data to arrive at influxdb
	time.Sleep(5 * time.Second)
	ticker := time.NewTicker(r.Interval)
	curtime := time.Now().Truncate(time.Minute)
	for {
		for _, d := range r.Archdepmod {
			// TODO: query for different types of components
			// TODO: change group by time according to r.Interval
			aggregation := viper.GetString("cfp.responsetime.aggregation")
			aggregationvalue := viper.GetString("cfp.responsetime.aggregationvalue")
			cmd := "select " + aggregation + "(\"response_time\"," + aggregationvalue + ") from operation_execution where \"hostname\" = '" + d.Component.Hostname + "' and \"operation_signature\" = '" + d.Component.Name + "' and time >= " + strconv.FormatInt(curtime.Add(-1*r.Interval).UnixNano(), 10) + " and time < " + strconv.FormatInt(curtime.UnixNano(), 10) + " group by time(" + r.Interval.String() + ")"
			q := client.Query{
				Command:  cmd,
				Database: r.Db,
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
				continue // no data - try next component
			}
			// Parse time and response time
			for _, row := range res[0].Series[0].Values {
				t, err := time.Parse(time.RFC3339, row[0].(string))
				if err != nil {
					log.Fatal(err)
				}

				if row[1] != nil {
					val, _ := row[1].(json.Number).Float64()
					point := TSPoint{d.Component, t, val}
					ch <- point
				}
			}
		}
		curtime = <-ticker.C
		curtime = curtime.Truncate(time.Minute)
	}
}

func (r *InfluxKiekerReader) getFirstAndLastTimestamp(clnt client.Client, c adm.Component) (time.Time, time.Time) {
	var firsttimestamp, lasttimestamp time.Time
	cmd := "select first(response_time) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "'"
	q := client.Query{
		Command:  cmd,
		Database: r.Db,
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

	// TODO: query for different components
	cmd = "select last(response_time) from operation_execution where \"hostname\" = '" + c.Hostname + "' and \"operation_signature\" = '" + c.Name + "'"
	q = client.Query{
		Command:  cmd,
		Database: r.Db,
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
